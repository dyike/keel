package loop

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackgroundUpdatesWithoutFrames(t *testing.T) {
	SetBackgroundUpdates(true)
	t.Cleanup(func() { SetBackgroundUpdates(false) })
	done := make(chan struct{})
	var got []int
	Post(func() {
		got = append(got, 1)
		Post(func() { got = append(got, 2); close(done) })
	})
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("windowless update stalled")
	}
	Lock()
	defer Unlock()
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("updates = %v", got)
	}
}

func TestBackgroundUpdatesShareFrameLock(t *testing.T) {
	SetBackgroundUpdates(true)
	t.Cleanup(func() { SetBackgroundUpdates(false) })
	Lock()
	var calls atomic.Int32
	done := make(chan struct{})
	Post(func() { calls.Add(1); close(done) })
	if calls.Load() != 0 {
		t.Fatal("callback ran through frame lock")
	}
	Unlock()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("update did not resume")
	}
	// A frame and the worker may race to Drain, but callbacks execute once.
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() { defer wg.Done(); Lock(); Drain(); Unlock() }()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatal("callback repeated")
	}
}

func TestDefaultUpdatesWaitForFrame(t *testing.T) {
	SetBackgroundUpdates(false)
	Lock()
	defer Unlock()
	var calls int
	Post(func() { calls++ })
	if calls != 0 {
		t.Fatal("default update became synchronous")
	}
	Drain()
	if calls != 1 {
		t.Fatal("frame did not drain update")
	}
}

func TestPendingReportsQueuedUpdates(t *testing.T) {
	SetBackgroundUpdates(false)
	Lock()
	defer Unlock()
	Drain()
	if Pending() {
		t.Fatal("empty queue reported pending")
	}
	Post(func() {})
	if !Pending() {
		t.Fatal("queued update not reported")
	}
	Drain()
	if Pending() {
		t.Fatal("drained queue reported pending")
	}
}
