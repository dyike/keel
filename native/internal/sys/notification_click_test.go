package sys

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNotificationClickReplacementRollbackAndConsumption(t *testing.T) {
	r := notificationClicks{}
	got := 0
	r.register("a", func() { got = 1 })(nil)
	failed := r.register("a", func() { got = 2 })
	failed(errors.New("denied"))
	fn := r.take("a")
	if fn == nil {
		t.Fatal("previous handler lost")
	}
	fn()
	if got != 1 || r.take("a") != nil {
		t.Fatal("rollback or one-shot consumption")
	}
	old := r.register("a", func() { got = 3 })
	r.register("a", func() { got = 4 })(nil)
	old(errors.New("late failure"))
	fn = r.take("a")
	fn()
	if got != 4 {
		t.Fatal("old completion overwrote replacement")
	}
	r.register("a", func() { got = 5 })(nil)
	r.register("a", nil)(nil)
	if r.take("a") != nil {
		t.Fatal("plain replacement kept click handler")
	}
	failed = r.register("b", func() {})
	failed(errors.New("unsupported"))
	if len(r.entries) != 0 {
		t.Fatal("failed first post leaked handler")
	}
}

func TestNotificationClickConcurrentResponse(t *testing.T) {
	r := notificationClicks{}
	var count atomic.Int32
	r.register("same", func() { count.Add(1) })(nil)
	var workers sync.WaitGroup
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if fn := r.take("same"); fn != nil {
				fn()
			}
		}()
	}
	workers.Wait()
	if count.Load() != 1 {
		t.Fatal("duplicate response invoked handler", count.Load())
	}
}
