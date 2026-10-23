package el

import (
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/theme"
)

// Ambient indicators share wall-clock ticks instead of drawing every display
// frame, and a sooner request from a transition still wins.
func TestAnimatingAmbientRequestsAlignedTick(t *testing.T) {
	transition := false
	root := Embed(viewFunc(func(cx *Context) Element {
		cx.AnimatingAmbient()
		cx.AnimatingAmbient()
		if transition {
			cx.Animating()
		}
		return Div().Size(Dp(10))
	}))
	var ops op.Ops
	var router input.Router
	frame := func(now time.Time) (time.Time, bool) {
		router.WakeupTime()
		ops.Reset()
		root.Layout(layout.Context{Ops: &ops, Now: now, Source: router.Source(),
			Constraints: layout.Exact(image.Pt(100, 100)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}})
		router.Frame(&ops)
		return router.PeekWakeup()
	}
	step := AmbientFrameInterval
	base := time.Unix(0, 30000*int64(step))
	for _, offset := range []time.Duration{0, time.Millisecond, step - time.Nanosecond} {
		at, ok := frame(base.Add(offset))
		if !ok || !at.Equal(base.Add(step)) {
			t.Fatalf("offset %v: wakeup %v %v, want %v", offset, at, ok, base.Add(step))
		}
	}
	transition = true
	if at, ok := frame(base.Add(time.Millisecond)); !ok || !at.IsZero() {
		t.Fatalf("transition frame should be immediate, got %v %v", at, ok)
	}
	transition = false
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(false)
	if at, ok := frame(base); ok {
		t.Fatalf("reduced motion requested a frame at %v", at)
	}
}
