// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"
	"time"

	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/unit"
)

// A GPU/offscreen test cannot reproduce Core Animation retaining the presented
// drawables. Drive a real frameless window through repeated size changes.
func TestMetalResizeWindow(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 for native Metal resize regression")
	}
	runMetalHelper(t, "GIO_METAL_RESIZE_HELPER", 30*time.Second)
}

func metalResizeWindowHelper() {
	fail := func(msg string) { fmt.Fprintln(os.Stderr, msg); os.Exit(1) }
	frames := make(chan image.Point, 64)
	w := new(Window)
	w.Option(Title("Metal resize regression"), Decorated(false), Size(unit.Dp(860), unit.Dp(600)), TopMost(true))
	go func() {
		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case DestroyEvent:
				if e.Err != nil {
					fail(e.Err.Error())
				}
				os.Exit(0)
			case FrameEvent:
				ops.Reset()
				paint.Fill(&ops, color.NRGBA{R: 252, G: 252, B: 251, A: 255})
				paint.FillShape(&ops, color.NRGBA{R: 242, G: 242, B: 241, A: 255}, clip.UniformRRect(image.Rect(30, 80, e.Size.X-30, e.Size.Y-30), 20).Op(&ops))
				e.Frame(&ops)
				select {
				case frames <- image.Pt(int(e.Metric.PxToDp(e.Size.X)), int(e.Metric.PxToDp(e.Size.Y))):
				default:
				}
			}
		}
	}()
	go func() {
		<-frames
		w.Perform(system.ActionRaise)
		time.Sleep(100 * time.Millisecond)
		var slowest time.Duration
		for i := range 60 {
			step := i
			if step >= 30 {
				step = 59 - step
			}
			want := image.Pt(860+step*18, 600+step*4)
			for len(frames) > 0 {
				<-frames
			}
			start := time.Now()
			w.Option(Size(unit.Dp(want.X), unit.Dp(want.Y)))
			deadline := time.NewTimer(600 * time.Millisecond)
		wait:
			for {
				select {
				case got := <-frames:
					if got == want {
						break wait
					}
				case <-deadline.C:
					exitIfOccluded(w, fmt.Sprintf("during resize %d", i))
					fail(fmt.Sprintf("resize %d to %v stalled waiting for a presented frame", i, want))
				}
			}
			deadline.Stop()
			slowest = max(slowest, time.Since(start))
			time.Sleep(16 * time.Millisecond)
		}
		fmt.Printf("60 native resizes: slowest %v\n", slowest)
		w.Perform(system.ActionClose)
	}()
	Main()
}
