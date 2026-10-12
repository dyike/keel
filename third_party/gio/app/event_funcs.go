// SPDX-License-Identifier: Unlicense OR MIT

package app

import "sync"

// eventFuncs transfers work to the event-loop thread. Wake under the mutex so
// shutdown cannot destroy the wakeup pipe while another thread is writing it.
type eventFuncs struct {
	mu      sync.Mutex
	pending []func()
	closed  bool
}

func (q *eventFuncs) add(f, wake func()) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	q.pending = append(q.pending, f)
	wake()
}

func (q *eventFuncs) drain() {
	q.mu.Lock()
	pending := q.pending
	q.pending = nil
	q.mu.Unlock()
	for _, f := range pending {
		f()
	}
}

func (q *eventFuncs) close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.pending = nil
}
