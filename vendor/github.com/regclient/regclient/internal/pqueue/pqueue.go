// Package pqueue implements a priority queue.
package pqueue

import (
	"context"
	"fmt"
	"slices"
	"sync"
)

type Queue[T any] struct {
	mu     sync.Mutex
	max    int
	next   func(queued, active []*T) int
	active []*T
	queued []*T
	wait   []*chan struct{}
}

// Opts is used to configure a new priority queue.
type Opts[T any] struct {
	Max  int                           // maximum concurrent entries, defaults to 1.
	Next func(queued, active []*T) int // function to lookup index of next queued entry to release, defaults to oldest entry.
}

// New creates a new priority queue.
func New[T any](opts Opts[T]) *Queue[T] {
	if opts.Max <= 0 {
		opts.Max = 1
	}
	return &Queue[T]{
		max:  opts.Max,
		next: opts.Next,
	}
}

// Acquire adds a new entry to the queue and returns once it is ready.
// The returned function must be called when the queued job completes to release the next entry.
// If there is any error, the returned function will be nil.
func (q *Queue[T]) Acquire(ctx context.Context, e T) (func(), error) {
	if q == nil {
		return func() {}, nil
	}
	found, err := q.checkContext(ctx)
	if err != nil {
		return nil, err
	}
	if found {
		return func() {}, nil
	}
	q.mu.Lock()
	if len(q.active)+len(q.queued) < q.max {
		q.active = append(q.active, &e)
		q.mu.Unlock()
		return q.releaseFn(&e), nil
	}
	// limit reached, add to queue and wait
	w := make(chan struct{}, 1)
	q.queued = append(q.queued, &e)
	q.wait = append(q.wait, &w)
	q.mu.Unlock()
	// wait on both context and queue
	select {
	case <-ctx.Done():
		// context abort, remove queued entry
		q.mu.Lock()
		if i := slices.Index(q.queued, &e); i >= 0 {
			q.queued = slices.Delete(q.queued, i, i+1)
			q.wait = slices.Delete(q.wait, i, i+1)
			q.mu.Unlock()
			return nil, ctx.Err()
		}
		q.mu.Unlock()
		// queued entry found, assume race condition with context and entry being released, release next entry
		q.release(&e)
		return nil, ctx.Err()
	case <-w:
		return q.releaseFn(&e), nil
	}
}

// TryAcquire attempts to add an entry on to the list of active entries.
// If the returned function is nil, the queue was not available.
// If the returned function is not nil, it must be called when the job is complete to release the next entry.
func (q *Queue[T]) TryAcquire(ctx context.Context, e T) (func(), error) {
	if q == nil {
		return func() {}, nil
	}
	found, err := q.checkContext(ctx)
	if err != nil {
		return nil, err
	}
	if found {
		return func() {}, nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.active)+len(q.queued) < q.max {
		q.active = append(q.active, &e)
		return q.releaseFn(&e), nil
	}
	return nil, nil
}

// release next entry or noop.
func (q *Queue[T]) release(prev *T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	// remove prev entry from active list
	if i := slices.Index(q.active, prev); i >= 0 {
		q.active = slices.Delete(q.active, i, i+1)
	}
	// skip checks when at limit or nothing queued
	if len(q.queued) == 0 {
		if len(q.active) == 0 {
			// free up slices if this was the last active entry
			q.active = nil
			q.queued = nil
			q.wait = nil
		}
		return
	}
	if len(q.active) >= q.max {
		return
	}
	i := 0
	if q.next != nil && len(q.queued) > 1 {
		i = q.next(q.queued, q.active)
		// validate response
		i = max(min(i, len(q.queued)-1), 0)
	}
	// release queued entry, move to active list, and remove from queued/wait lists
	close(*q.wait[i])
	q.active = append(q.active, q.queued[i])
	q.queued = slices.Delete(q.queued, i, i+1)
	q.wait = slices.Delete(q.wait, i, i+1)
}

// releaseFn is a convenience wrapper around [release].
func (q *Queue[T]) releaseFn(prev *T) func() {
	return func() {
		q.release(prev)
	}
}

// TODO: is there a way to make a different context key for each generic type?
type ctxType int

var ctxKey ctxType

type valMulti[T any] struct {
	qList []*Queue[T]
}

// AcquireMulti is used to simultaneously lock multiple queues without the risk of deadlock.
// The returned context needs to be used on calls to [Acquire] or [TryAcquire] which will immediately succeed since the resource is already acquired.
// Attempting to acquire other resources with [Acquire], [TryAcquire], or [AcquireMulti] using the returned context and will fail for being outside of the transaction.
// The returned function must be called to release the resources.
// The returned function is not thread safe, ensure no other simultaneous calls to [Acquire] or [TryAcquire] using the returned context have finished before it is called.
func AcquireMulti[T any](ctx context.Context, e T, qList ...*Queue[T]) (context.Context, func(), error) {
	// verify context not already holding locks
	qCtx := ctx.Value(ctxKey)
	if qCtx != nil {
		if qCtxVal, ok := qCtx.(*valMulti[T]); !ok || qCtxVal.qList != nil {
			return ctx, nil, fmt.Errorf("context already used by another AcquireMulti request")
		}
	}
	// delete nil entries
	for i := len(qList) - 1; i >= 0; i-- {
		if qList[i] == nil {
			qList = slices.Delete(qList, i, i+1)
		}
	}
	// empty/nil list is a noop
	if len(qList) == 0 {
		return ctx, func() {}, nil
	}
	// dedup entries from the list
	for i := len(qList) - 2; i >= 0; i-- {
		for j := len(qList) - 1; j > i; j-- {
			if qList[i] == qList[j] {
				qList[j] = qList[len(qList)-1]
				qList = qList[:len(qList)-1]
			}
		}
	}
	// Loop through queues to acquire, waiting on the first, and attempting the remaining.
	// If any of the remaining entries cannot be immediately acquired, reset and make it the new queue to wait on.
	lockI := 0
	doneList := make([]func(), len(qList))
	for {
		acquired := true
		i := 0
		done, err := qList[lockI].Acquire(ctx, e)
		if err != nil {
			return ctx, nil, err
		}
		doneList[lockI] = done
		for i < len(qList) {
			if i != lockI {
				doneList[i], err = qList[i].TryAcquire(ctx, e)
				if doneList[i] == nil || err != nil {
					acquired = false
					break
				}
			}
			i++
		}
		if err == nil && acquired {
			break
		}
		// cleanup on failed attempt
		if lockI > i {
			doneList[lockI]()
		}
		// track blocking index for a retry
		lockI = i
		for i > 0 {
			i--
			doneList[i]()
		}
		// abort on errors
		if err != nil {
			return ctx, nil, err
		}
	}
	// success, update context
	ctxVal := valMulti[T]{qList: qList}
	newCtx := context.WithValue(ctx, ctxKey, &ctxVal)
	cleanup := func() {
		ctxVal.qList = nil
		// dequeue in reverse order to minimize chance of another AcquireMulti being freed and immediately blocking on the next queue
		for i := len(doneList) - 1; i >= 0; i-- {
			doneList[i]()
		}
	}
	return newCtx, cleanup, nil
}

func (q *Queue[T]) checkContext(ctx context.Context) (bool, error) {
	qCtx := ctx.Value(ctxKey)
	if qCtx == nil {
		return false, nil
	}
	qCtxVal, ok := qCtx.(*valMulti[T])
	if !ok {
		return false, nil // another type is using the context, treat it as unset
	}
	if qCtxVal.qList == nil {
		return false, nil
	}
	if slices.Contains(qCtxVal.qList, q) {
		// instance already locked
		return true, nil
	}
	return true, fmt.Errorf("cannot acquire new locks during a transaction")
}
