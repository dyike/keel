package loop

import (
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const reclaimQuiet = 2 * time.Second
const reclaimCooldown = 30 * time.Second
const reclaimMinimum = 32 << 20

// Idle reclamation is opt-in because Go's heap belongs to the whole process.
// It does not request frames or modify GOGC/GOMEMLIMIT.
var reclaimEnabled atomic.Bool
var memoryIdle = idleReclaimer{reclaim: debug.FreeOSMemory}

type idleReclaimer struct {
	mu         sync.Mutex
	enabled    bool
	last, next time.Time
	timer      *time.Timer
	generation uint64
	reclaim    func()
	readStats  func(*runtime.MemStats)
}

func SetIdleMemoryReclaim(enabled bool) {
	memoryIdle.mu.Lock()
	defer memoryIdle.mu.Unlock()
	reclaimEnabled.Store(enabled)
	memoryIdle.enabled = enabled
	memoryIdle.generation++
	if memoryIdle.timer != nil {
		memoryIdle.timer.Stop()
		memoryIdle.timer = nil
	}
	if enabled {
		memoryIdle.last = time.Now()
		memoryIdle.scheduleLocked(reclaimQuiet)
	}
}

// Track every window and UI update, not just the window that opted in.
func memoryActivity() {
	if !reclaimEnabled.Load() {
		return
	}
	memoryIdle.mu.Lock()
	defer memoryIdle.mu.Unlock()
	if !memoryIdle.enabled {
		return
	}
	memoryIdle.last = time.Now()
	if memoryIdle.timer == nil {
		memoryIdle.scheduleLocked(reclaimQuiet)
	}
}

func reclaimDelay(now, last, next time.Time) time.Duration {
	due := last.Add(reclaimQuiet)
	if next.After(due) {
		due = next
	}
	if now.Before(due) {
		return due.Sub(now)
	}
	return 0
}

func reclaimable(idle, released uint64) bool {
	return idle >= released && idle-released >= reclaimMinimum
}

func (r *idleReclaimer) scheduleLocked(delay time.Duration) {
	generation := r.generation
	r.timer = time.AfterFunc(delay, func() { r.fire(generation) })
}

func (r *idleReclaimer) fire(generation uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled || generation != r.generation {
		return
	}
	r.timer = nil
	now := time.Now()
	if delay := reclaimDelay(now, r.last, r.next); delay > 0 {
		r.scheduleLocked(delay)
		return
	}
	// TryLock avoids collecting while a Keel UI callback or layout is running.
	// Do not block behind active frames, or wake the UI to perform cleanup.
	if !frame.TryLock() {
		r.last = now
		r.scheduleLocked(reclaimQuiet)
		return
	}
	defer frame.Unlock()
	var stats runtime.MemStats
	if r.readStats != nil {
		r.readStats(&stats)
	} else {
		runtime.ReadMemStats(&stats)
	}
	if !reclaimable(stats.HeapIdle, stats.HeapReleased) {
		return
	}
	r.reclaim()
	r.next = time.Now().Add(reclaimCooldown)
	// Leave the timer stopped until subsequent UI activity. Static windows do
	// not accumulate periodic GC work or maintenance frames.
}
