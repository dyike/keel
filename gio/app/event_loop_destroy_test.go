// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"testing"
	"time"
)

func TestEventLoopIgnoresWakeupAfterDestroy(t *testing.T) {
	w := new(Window)
	loop := newEventLoop(&callbacks{w}, func() {})
	consumer := make(chan struct{})
	go func() {
		defer close(consumer)
		if _, ok := loop.Event().(DestroyEvent); !ok {
			t.Error("expected terminal destroy event")
		}
	}()
	loop.deliverEvent(DestroyEvent{})
	<-consumer
	// A native dispatch_async wakeup queued before close can run after the
	// only consumer has returned. It used to block the platform's main thread.
	w.processEvent(wakeupEvent{})
	finished := make(chan struct{})
	go func() { loop.FlushEvents(); close(finished) }()
	select {
	case <-finished:
	case <-time.After(time.Second):
		go func() { <-loop.events; <-loop.events }()
		t.Fatal("late wakeup blocked after DestroyEvent")
	}
}
