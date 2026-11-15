// Copyright (c) 2026 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exsync

import (
	"iter"
	"sync"
)

// Queue is an unbounded thread-safe FIFO queue.
type Queue[T any] struct {
	l       sync.Mutex
	input   []T
	read    []T
	readPtr int
	maxCap  int
}

func NewQueue[T any]() *Queue[T] {
	return NewQueueWithMaxCap[T](16)
}

func NewQueueWithMaxCap[T any](maxCap int) *Queue[T] {
	return &Queue[T]{maxCap: maxCap}
}

// Push adds an item to the queue.
func (q *Queue[T]) Push(val T) {
	q.l.Lock()
	defer q.l.Unlock()
	q.input = append(q.input, val)
}

// Len returns the number of items in the queue.
func (q *Queue[T]) Len() int {
	q.l.Lock()
	defer q.l.Unlock()
	return len(q.input) + len(q.read) - q.readPtr
}

// Iter is an [iter.Seq] that pops items from the queue until the queue is empty.
func (q *Queue[T]) Iter(yield func(T) bool) {
	for {
		val, ok := q.Pop()
		if !ok || !yield(val) {
			return
		}
	}
}

var _ iter.Seq[any] = (*Queue[any])(nil).Iter

// Pop gets the next item in the queue.
func (q *Queue[T]) Pop() (val T, ok bool) {
	q.l.Lock()
	defer q.l.Unlock()

	if len(q.read) <= q.readPtr {
		if len(q.input) == 0 {
			return
		} else if len(q.input) == 1 {
			// TODO determine if this fast path is actually faster than the default path
			val, q.input[0] = q.input[0], val
			q.input = q.input[:0]
			return val, true
		}
		q.read, q.input = q.input, q.read[:0]
		q.readPtr = 0
	}

	val, q.read[q.readPtr] = q.read[q.readPtr], val
	q.readPtr++
	// If we read to the end of an oversized array, free it immediately
	if q.readPtr >= len(q.read) && q.maxCap >= 0 && len(q.read) > q.maxCap {
		q.read = nil
		q.readPtr = 0
	}
	return val, true
}
