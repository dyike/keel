package el

import (
	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"time"
)

type viewTimer struct {
	due      time.Time
	duration time.Duration
	frame    uint64
	fn       func()
	fired    bool
}

// Now is the frame timestamp. Animations must derive their phase from it.
func (cx *Context) Now() time.Time { return cx.root.e.gtx.Now }

// Animating requests the next animation frame without spawning a goroutine.
func (cx *Context) Animating() {
	if cx.root.e.gtx.Enabled() && !ReducedMotion() {
		cx.root.e.gtx.Execute(op.InvalidateCmd{})
	}
}

// ReducedMotion returns the application override; platforms without a native
// preference bridge default to false. Set through theme.SetReducedMotion.
func ReducedMotion() bool { return theme.ReducedMotion }

// After declares a one-shot timer with a comparable key unique to this root.
// Declare the key on every Render while the timer is alive; omitting it cancels
// it. Changing d restarts the timer. A fired timer does not rearm until omitted
// for a frame. Do not declare timers in Cache builders, which may not run again.
func (cx *Context) After(key any, d time.Duration, fn func()) {
	r := cx.root
	if !r.e.gtx.Enabled() {
		return
	}
	if r.timers == nil {
		r.timers = map[any]*viewTimer{}
	}
	timer := r.timers[key]
	if timer == nil || timer.duration != d {
		timer = &viewTimer{due: cx.Now().Add(d), duration: d}
		r.timers[key] = timer
	}
	timer.frame = r.timerEpoch
	timer.fn = fn
}
func (r *RootWidget) beginTimers() { r.timerEpoch++ }
func (r *RootWidget) finishTimers() {
	gtx := r.e.gtx
	if !gtx.Enabled() {
		return
	}
	// Collect before invoking callbacks so callbacks cannot mutate this iteration.
	var due []func()
	for k, timer := range r.timers {
		if timer.frame != r.timerEpoch {
			delete(r.timers, k)
			continue
		}
		if timer.fired {
			continue
		}
		if !gtx.Now.Before(timer.due) {
			timer.fired = true
			if timer.fn != nil {
				due = append(due, timer.fn)
			}
		} else {
			gtx.Execute(op.InvalidateCmd{At: timer.due})
		}
	}
	for _, fn := range due {
		core.Call(gtx, fn)
	}
}
