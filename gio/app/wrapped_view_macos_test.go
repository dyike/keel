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
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/unit"
	"github.com/ebitengine/purego/objc"
)

func TestMacOSWrappedInputView(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 for native wrapped-view regression")
	}
	runMetalHelper(t, "GIO_WRAPPED_VIEW_HELPER", 15*time.Second)
}

type wrappedViewRect struct{ X, Y, Width, Height float64 }

// Replacing contentView must preserve the original input view used by window
// delegates. The close notification used to cast the wrapper to GioView.
func wrappedInputViewHelper() {
	fail := func(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
	frames := make(chan image.Point, 32)
	w := new(Window)
	w.Option(Title("Wrapped input view regression"), Size(unit.Dp(320), unit.Dp(240)), TopMost(true))
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
				paint.Fill(&ops, color.NRGBA{R: 30, G: 40, B: 50, A: 255})
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
		w.Run(func() {
			view := objc.ID(uintptr(w.driver.(*window).view))
			win := view.Send(objc.RegisterName("window"))
			bounds := objc.Send[wrappedViewRect](view, objc.RegisterName("bounds"))
			view.Send(objc.RegisterName("retain"))
			root := objc.ID(objc.GetClass("NSView")).Send(objc.RegisterName("alloc")).Send(objc.RegisterName("init"))
			win.Send(objc.RegisterName("setContentView:"), root)
			root.Send(objc.RegisterName("addSubview:"), view)
			view.Send(objc.RegisterName("setFrame:"), bounds)
			view.Send(objc.RegisterName("setAutoresizingMask:"), uintptr(18))
			win.Send(objc.RegisterName("makeFirstResponder:"), view)
			if win.Send(objc.RegisterName("firstResponder")) != view || win.Send(objc.RegisterName("contentView")) == view {
				fail("input view identity lost after wrapping")
			}
			root.Send(objc.RegisterName("release"))
			view.Send(objc.RegisterName("release"))
		})
		for len(frames) > 0 {
			<-frames
		}
		w.Option(Size(unit.Dp(400), unit.Dp(260)))
		deadline := time.NewTimer(2 * time.Second)
		defer deadline.Stop()
		for {
			select {
			case size := <-frames:
				if size == image.Pt(400, 260) {
					fmt.Println("PASS: wrapped GioView retains focus, resizes and closes through native delegates")
					w.Perform(system.ActionClose)
					return
				}
			case <-deadline.C:
				exitIfOccluded(w, "during wrapped-view resize")
				fail("wrapped input view stopped drawing")
			}
		}
	}()
	Main()
}
