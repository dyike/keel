package el

import (
	"gioui.org/op"
	"github.com/dyike/keel/ui/theme"
	"runtime"
	"strconv"
	"time"
)

type timerKey struct {
	scope string
	site  uintptr
	index int
}
type timerCall struct {
	scope string
	site  uintptr
}
type viewTimer struct {
	due      time.Time
	duration time.Duration
	frame    uint64
	owner    string
	fn       func()
	fired    bool
}

// Scope binds timer declarations to an element ID in this root. Use stable IDs
// for repeated views. The element must be present and visible in the returned tree.
func (cx *Context) Scope(id string) *Context {
	return &Context{root: cx.root, scope: cx.scope + strconv.Itoa(len(id)) + ":" + id, owner: id}
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

// After declares a one-shot timer. Declare it on every Render while it is alive;
// omitting it, or removing its Scope owner, cancels it. A fired declaration does
// not rearm until omitted for a frame. Call sites identify timers within a scope.
func (cx *Context) After(d time.Duration, fn func()) {
	r := cx.root
	if r.timers == nil {
		r.timers = map[timerKey]*viewTimer{}
	}
	pc, _, _, _ := runtime.Caller(1)
	call := timerCall{cx.scope, pc}
	index := r.timerCalls[call]
	r.timerCalls[call]++
	k := timerKey{cx.scope, pc, index}
	timer := r.timers[k]
	if timer == nil || timer.duration != d {
		timer = &viewTimer{due: cx.Now().Add(d), duration: d}
		r.timers[k] = timer
	}
	timer.frame = r.timerEpoch
	timer.owner = cx.owner
	timer.fn = fn
}
func (r *RootWidget) beginTimers() { r.timerEpoch++; r.timerCalls = map[timerCall]int{} }
func (r *RootWidget) finishTimers(tree *Node) {
	gtx := r.e.gtx
	if !gtx.Enabled() {
		return
	}
	ids := map[string]bool{}
	var visit func(*Node)
	visit = func(n *Node) {
		if n.style.hidden {
			return
		}
		if n.id != "" {
			ids[n.id] = true
		}
		for _, c := range n.children {
			visit(c.node())
		}
	}
	visit(tree)
	// Collect before invoking callbacks so callbacks cannot mutate this iteration.
	var due []func()
	for k, timer := range r.timers {
		if timer.frame != r.timerEpoch || timer.owner != "" && !ids[timer.owner] {
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
		fn()
	}
	if len(due) > 0 {
		gtx.Execute(op.InvalidateCmd{})
	}
}
