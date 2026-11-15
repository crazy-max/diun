package amqp091

import (
	"fmt"
	"sync"
)

// LifeCycleState defines the connection or channel state type.
type LifeCycleState byte

const (
	// StateOpen represents a connection or channel that is active and open.
	StateOpen LifeCycleState = iota
	// StateReconnecting represents a connection or channel that is actively undergoing automatic recovery.
	StateReconnecting
	// StateClosing represents a connection or channel that is in the process of closing down.
	StateClosing
	// StateClosed represents a connection or channel that is fully closed and shut down.
	StateClosed
)

func (s LifeCycleState) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateReconnecting:
		return "reconnecting"
	case StateClosing:
		return "closing"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// StateChanged defines the connection or channel life cycle transitions.
type StateChanged struct {
	From LifeCycleState
	To   LifeCycleState

	// Err is set only on the →StateClosed transition and carries the fatal error
	// that ended recovery (exhausted retries, non-recoverable error, etc.).
	Err error

	// SkippedTopologyEntities is set only on the StateReconnecting→StateOpen transition when
	// one or more topology entities were skipped via Recovery.OnTopologyEntityError.
	// Nil on all other transitions and when all entities recovered successfully.
	SkippedTopologyEntities []TopologyRecoveryEntity
}

func (s StateChanged) String() string {
	if s.Err != nil {
		return fmt.Sprintf("From: %s, To: %s, Error: %v", s.From, s.To, s.Err)
	}
	return fmt.Sprintf("From: %s, To: %s", s.From, s.To)
}

type stateListener struct {
	ch      chan *StateChanged
	queue   []*StateChanged
	sending bool
}

// enqueue appends a state change to the listener's bounded queue using a sliding window.
// If the queue size exceeds maxQueueSize, the oldest state change is dropped.
// This assumes the caller holds the lifeCycle mutex.
func (sl *stateListener) enqueue(sc *StateChanged) {
	const maxQueueSize = 50
	if len(sl.queue) < maxQueueSize {
		sl.queue = append(sl.queue, sc)
	} else {
		sl.queue[0] = nil // Allow GC to reclaim the dropped element
		sl.queue = append(sl.queue[1:], sc)
	}
}

// lifeCycle is the lifecycle of the connection or channel.
//
// The listener framework manages state change notifications for registered channels.
// Key characteristics:
//   - Multiple listeners can register concurrently via NotifyStateChange.
//   - Each listener operates concurrently and is isolated from others; a blocked or slow
//     listener will not block other listeners or the main SetState() execution.
//   - Strict FIFO ordering of state transitions is guaranteed for each listener by spawning
//     at most one dedicated delivery goroutine per listener.
//   - Memory usage is strictly bounded by a sliding-window queue per listener (maxQueueSize = 50).
//   - Listener channels are cleanly closed as soon as the final StateClosed transition is sent.
type lifeCycle struct {
	state     LifeCycleState   // The current state of the connection or channel.
	listeners []*stateListener // The registered state change listeners.
	mutex     *sync.Mutex      // The mutex to protect the state changes.
}

func newLifeCycle() *lifeCycle {
	return &lifeCycle{
		state: StateClosed,
		mutex: &sync.Mutex{},
	}
}

func (l *lifeCycle) State() LifeCycleState {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.state
}

func (l *lifeCycle) SetState(value LifeCycleState, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.state == value {
		return
	}

	oldState := l.state
	l.state = value

	sc := &StateChanged{
		From: oldState,
		To:   value,
		Err:  err,
	}

	for _, listener := range l.listeners {
		listener.enqueue(sc)
		if !listener.sending {
			listener.sending = true
			go l.deliverToListener(listener)
		}
	}
}

// setStateOpen transitions to StateOpen and attaches any topology entities that
// were skipped during recovery. Called only from Connection.Reconnect.
func (l *lifeCycle) setStateOpen(skippedTopologyEntities []TopologyRecoveryEntity) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.state == StateOpen {
		return
	}
	oldState := l.state
	l.state = StateOpen
	sc := &StateChanged{
		From: oldState,
		To:   StateOpen,
	}
	if len(skippedTopologyEntities) != 0 {
		sc.SkippedTopologyEntities = skippedTopologyEntities
	}
	for _, listener := range l.listeners {
		listener.enqueue(sc)
		if !listener.sending {
			listener.sending = true
			go l.deliverToListener(listener)
		}
	}
}

func (l *lifeCycle) deliverToListener(listener *stateListener) {
	for {
		l.mutex.Lock()
		if len(listener.queue) == 0 {
			listener.sending = false
			l.mutex.Unlock()
			return
		}
		sc := listener.queue[0]
		listener.queue[0] = nil // Allow GC to reclaim the popped element
		listener.queue = listener.queue[1:]
		ch := listener.ch
		l.mutex.Unlock()

		ch <- sc

		// If the transition is to StateClosed, this is the terminal state.
		// We can safely close the channel now because:
		// 1. No more states will be appended (StateClosed is final).
		// 2. The StateClosed notification was just successfully sent.
		if sc.To == StateClosed {
			l.mutex.Lock()
			close(ch)
			l.removeListener(listener)
			l.mutex.Unlock()
			return
		}
	}
}

func (l *lifeCycle) removeListener(listener *stateListener) {
	for i, lis := range l.listeners {
		if lis == listener {
			copy(l.listeners[i:], l.listeners[i+1:])
			l.listeners[len(l.listeners)-1] = nil // Allow GC to reclaim the listener
			l.listeners = l.listeners[:len(l.listeners)-1]
			break
		}
	}
}

func (l *lifeCycle) notifyStateChange(channel chan *StateChanged) {
	if channel == nil {
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// If the connection/channel is already closed, close the channel immediately.
	if l.state == StateClosed {
		close(channel)
		return
	}

	// Prevent duplicate registration of the same channel.
	for _, lis := range l.listeners {
		if lis.ch == channel {
			return
		}
	}

	listener := &stateListener{
		ch: channel,
	}
	l.listeners = append(l.listeners, listener)
}
