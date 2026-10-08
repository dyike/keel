package loop

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestReclaimPolicy(t *testing.T) {
	now := time.Now()
	if reclaimDelay(now, now, time.Time{}) != reclaimQuiet {
		t.Fatal("active UI not deferred")
	}
	if reclaimDelay(now, now.Add(-time.Minute), now.Add(reclaimCooldown)) != reclaimCooldown {
		t.Fatal("cooldown ignored")
	}
	if reclaimDelay(now, now.Add(-time.Minute), time.Time{}) != 0 {
		t.Fatal("idle UI not eligible")
	}
	for _, tc := range []struct {
		idle, released uint64
		want           bool
	}{{0, 1, false}, {reclaimMinimum - 1, 0, false}, {reclaimMinimum, 0, true}, {64 << 20, 40 << 20, false}} {
		if reclaimable(tc.idle, tc.released) != tc.want {
			t.Fatalf("incorrect page threshold: %+v", tc)
		}
	}
}

func TestIdleReclaimerRunsOnceAndHonorsActivity(t *testing.T) {
	calls := 0
	r := idleReclaimer{enabled: true, last: time.Now().Add(-time.Minute), reclaim: func() { calls++ }, readStats: func(s *runtime.MemStats) { s.HeapIdle = 64 << 20 }}
	defer func() {
		if r.timer != nil {
			r.timer.Stop()
		}
	}()
	// A callback/layout in any Keel window prevents reclamation.
	frame.Lock()
	r.fire(0)
	frame.Unlock()
	if calls != 0 || r.timer == nil {
		t.Fatal("reclaimed during a UI callback")
	}
	r.timer.Stop()
	r.timer = nil
	r.last = time.Now().Add(-time.Minute)
	r.fire(0)
	if calls != 1 || r.timer != nil || !r.next.After(time.Now()) {
		t.Fatal("idle collection did not stop or establish cooldown")
	}
	r.fire(0)
	if calls != 1 || r.timer == nil {
		t.Fatal("cooldown did not defer repeat collection")
	}
	r.timer.Stop()
	r.timer = nil
	r.enabled = false
	r.fire(0)
	if calls != 1 || r.timer != nil {
		t.Fatal("disabled collector ran")
	}
	r.enabled = true
	r.generation = 1
	r.fire(0)
	if calls != 1 || r.timer != nil {
		t.Fatal("cancelled timer ran")
	}
}

func TestIdleReclaimerSkipsSmallHeap(t *testing.T) {
	r := idleReclaimer{enabled: true, last: time.Now().Add(-time.Minute), reclaim: func() { t.Fatal("small heap collected") }, readStats: func(s *runtime.MemStats) { s.HeapIdle = 64 << 20; s.HeapReleased = 60 << 20 }}
	r.fire(0)
	if r.timer != nil {
		r.timer.Stop()
		t.Fatal("small static heap scheduled periodic work")
	}
}

func TestIdleMemoryConcurrentConfiguration(t *testing.T) {
	defer SetIdleMemoryReclaim(false)
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				SetIdleMemoryReclaim(i%2 == 0)
			}
		}()
	}
	wg.Wait()
	SetIdleMemoryReclaim(false)
	memoryIdle.mu.Lock()
	defer memoryIdle.mu.Unlock()
	if memoryIdle.enabled || reclaimEnabled.Load() || memoryIdle.timer != nil {
		t.Fatal("disabled configuration left work scheduled")
	}
}
