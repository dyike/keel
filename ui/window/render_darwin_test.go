//go:build darwin && !ios && !nometal

package window

import (
	"context"
	"fmt"
	"image"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	gioapp "github.com/dyike/keel/gio/app"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/ebitengine/purego"
)

func TestMain(m *testing.M) {
	if os.Getenv("KEEL_RESIZE_HELPER") == "1" {
		resizeWindowHelper()
		return
	}
	os.Exit(m.Run())
}

// Offscreen GPU pixels cannot test an old drawable's Core Animation gravity
// or a presentation made inside displayLayer's transaction.
func TestRealWindowsResize(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 for native resize regression")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Env = append(os.Environ(), "KEEL_RESIZE_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "PASS") {
		t.Fatalf("%v\n%s", err, out)
	}
	t.Log(string(out))
}

func resizeWindowHelper() {
	fail := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		os.Exit(1)
	}
	// NSWindow's struct argument plus display flag is only needed by this
	// native test. Bind once, without adding a production binding or callback.
	var setWindowFrame func(appkit.ID, appkit.SEL, appkit.Rect, bool)
	purego.RegisterFunc(&setWindowFrame, appkit.Sym("objc_msgSend"))
	// Keep the process alive while switching between ordinary and glass windows.
	keeper := Open(Options{Title: "resize keeper", Width: 180, Height: 100})
	go func() {
		for _, glass := range []bool{false, true} {
			var w *Window
			ready := make(chan struct{})
			var frames, synchronous atomic.Int32
			var scrollFrames atomic.Int32
			pendingScroll, scrollOffset := false, float32(0) // frame lock
			hooked := false                                  // window goroutine only
			content := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
				if pendingScroll {
					cx.ScrollTo("text", scrollOffset)
					pendingScroll = false
					scrollFrames.Add(1)
				}
				text := el.Div().ID("text").WFull().HFull().ScrollY()
				for i := range 100 {
					text.Child(el.Text(fmt.Sprintf("%03d Resize text: 中文 Aa 0123456789", i)))
				}
				return text
			}))
			opts := Options{Title: "resize regression", Width: 480, Height: 320, Frameless: true, TopMost: true}
			if glass {
				opts.Glass = &GlassOptions{Style: GlassFrosted}
			}
			opts.Content = core.Func(func(gtx core.C) core.D {
				<-ready
				if r, ok := w.glassRenderer.(*metalGlassRenderer); ok && !hooked {
					hooked = true
					run := r.run
					r.run = func(present func()) {
						run(func() {
							if !appkit.IsMainThread() || !appkit.SendBool(r.context.layer, "presentsWithTransaction") {
								fail("resize present must run on main inside a transaction")
							}
							present()
							synchronous.Add(1)
						})
					}
				}
				frames.Add(1)
				return content.Layout(gtx)
			})
			w = Open(opts)
			close(ready)
			<-w.shown
			// All probes run outside the frame lock. Use Gio Run for the same
			// event-loop rendezvous as presentation, not a dispatch_sync that
			// could wait behind displayLayer while it waits for a frame.
			onView := func(fn func(appkit.ID)) {
				loop.Lock()
				view := appkit.Retain(appkit.ID(w.nativeView))
				loop.Unlock()
				w.win.Run(func() {
					appkit.Pool(func() {
						defer appkit.Release(view)
						fn(view)
					})
				})
			}
			lastCheck, phase := "", "initial"
			check := func(size image.Point, idle bool) bool {
				ok := false
				onView(func(view appkit.ID) {
					layer := appkit.Send(view, "layer")
					redraw, placement := appkit.Send(view, "layerContentsRedrawPolicy"), appkit.Send(view, "layerContentsPlacement")
					if redraw != nsViewLayerContentsRedrawDuringViewResize || placement != nsViewLayerContentsPlacementTopLeft {
						lastCheck = fmt.Sprintf("AppKit redraw=%d placement=%d", redraw, placement)
						return // native setup may still be pending
					}
					gravity := appkit.GoString(appkit.Send(layer, "contentsGravity"))
					lastCheck = fmt.Sprintf("gravity=%s", gravity)
					want := "topLeft"
					if appkit.SendBool(layer, "contentsAreFlipped") {
						want = "bottomLeft"
					}
					if gravity != want {
						return // installation and its retained layer dispatch may still be pending
					}
					bounds := appkit.MsgRect(layer, appkit.Sel("bounds")).Size
					scale := appkit.MsgFloat(layer, appkit.Sel("contentsScale"))
					drawable := appkit.MsgSize(layer, appkit.Sel("drawableSize"))
					anchor := appkit.MsgPoint(layer, appkit.Sel("anchorPoint"))
					transformScale := func(axis string) float64 {
						value := appkit.Send(layer, "valueForKeyPath:", uintptr(appkit.String("transform.scale."+axis)))
						return appkit.MsgFloat(value, appkit.Sel("doubleValue"))
					}
					sx, sy := transformScale("x"), transformScale("y")
					if sx != 1 || sy != 1 {
						fail("glass=%v unexpected model-layer scale (%g,%g)", glass, sx, sy)
					}
					lastCheck = fmt.Sprintf("redraw=%d placement=%d gravity=%s anchor=%v transformScale=(%g,%g) bounds=%v scale=%v drawable=%v", redraw, placement, gravity, anchor, sx, sy, bounds, scale, drawable)
					wantSize := appkit.Size{Width: bounds.Width * scale, Height: bounds.Height * scale}
					if idle {
						wantSize = appkit.Size{Width: 1, Height: 1}
					}
					ok = bounds.Width == float64(size.X) && bounds.Height == float64(size.Y) && drawable == wantSize
				})
				return ok
			}
			wait := func(size image.Point, idle bool) {
				timeout := 600 * time.Millisecond
				if phase == "initial" {
					timeout = 5 * time.Second // async glass installation is not a resize
				}
				deadline := time.Now().Add(timeout)
				for !check(size, idle) {
					if time.Now().After(deadline) {
						fail("glass=%v phase=%s size=%v idle=%v stalled: %s", glass, phase, size, idle, lastCheck)
					}
					time.Sleep(time.Millisecond)
				}
			}
			settleFrames := func() {
				// isZoomed and the model bounds can reach their target before
				// AppKit finishes the animated resize. Wait for its terminal
				// frames before testing fixed-size scrolling or starting a
				// reverse transition. This only samples the test's frame count.
				deadline, quietSince := time.Now().Add(3*time.Second), time.Now()
				lastFrame := frames.Load()
				for time.Since(quietSince) < 250*time.Millisecond {
					if current := frames.Load(); current != lastFrame {
						lastFrame, quietSince = current, time.Now()
					}
					if time.Now().After(deadline) {
						fail("glass=%v %s never settled", glass, phase)
					}
					time.Sleep(5 * time.Millisecond)
				}
			}
			wait(image.Pt(480, 320), false)
			fmt.Printf("glass=%v native layer: %s\n", glass, lastCheck)
			var slowest time.Duration
			for i := range 60 {
				phase = fmt.Sprintf("resize %d", i)
				step := i
				if step >= 30 {
					step = 59 - step
				}
				size := image.Pt(500+step*12, 340+step*6)
				start := time.Now()
				w.win.Option(gioapp.Size(unit.Dp(size.X), unit.Dp(size.Y)))
				wait(size, false)
				slowest = max(slowest, time.Since(start))
				time.Sleep(16 * time.Millisecond)
			}
			if glass && synchronous.Load() < 59 {
				fail("only %d synchronous resize presents", synchronous.Load())
			}
			// Reproduce resizing around the window's center: both opposite
			// edges move, as with Option-dragging a native resize handle.
			var center appkit.Point
			onView(func(view appkit.ID) {
				frame := appkit.MsgRect(appkit.Send(view, "window"), appkit.Sel("frame"))
				center = appkit.Point{X: frame.Origin.X + frame.Size.Width/2, Y: frame.Origin.Y + frame.Size.Height/2}
			})
			var centerSlowest time.Duration
			for i := range 60 {
				phase = fmt.Sprintf("center resize %d", i)
				step := i
				if step >= 30 {
					step = 59 - step
				}
				size := image.Pt(500+step*12, 340+step*6)
				start := time.Now()
				onView(func(view appkit.ID) {
					window := appkit.Send(view, "window")
					frame := appkit.Rect{Origin: appkit.Point{X: center.X - float64(size.X)/2, Y: center.Y - float64(size.Y)/2}, Size: appkit.Size{Width: float64(size.X), Height: float64(size.Y)}}
					setWindowFrame(window, appkit.Sel("setFrame:display:"), frame, true)
					actual := appkit.MsgRect(window, appkit.Sel("frame"))
					if math.Abs(actual.Origin.X+actual.Size.Width/2-center.X) > 0.5 || math.Abs(actual.Origin.Y+actual.Size.Height/2-center.Y) > 0.5 {
						fail("window center moved during %s: %v", phase, actual)
					}
				})
				wait(size, false)
				centerSlowest = max(centerSlowest, time.Since(start))
				time.Sleep(16 * time.Millisecond)
			}
			if glass && synchronous.Load() < 117 {
				fail("only %d synchronous presents after both resize modes", synchronous.Load())
			}
			// ToggleMaximize uses AppKit's NSWindow zoom/restore, including
			// its native animation. Check each transition settles with a
			// matching full-resolution drawable and unchanged view policy.
			for i := range 4 {
				phase = fmt.Sprintf("native zoom %d", i)
				core.Update(w.ToggleMaximize)
				wantZoomed := i%2 == 0
				deadline := time.Now().Add(3 * time.Second)
				for {
					var zoomed bool
					var size image.Point
					onView(func(view appkit.ID) {
						zoomed = appkit.SendBool(appkit.Send(view, "window"), "isZoomed")
						bounds := appkit.MsgRect(view, appkit.Sel("bounds")).Size
						size = image.Pt(int(bounds.Width), int(bounds.Height))
					})
					if zoomed == wantZoomed && check(size, false) {
						break
					}
					if time.Now().After(deadline) {
						fail("glass=%v %s stalled (isZoomed=%v): %s", glass, phase, zoomed, lastCheck)
					}
					time.Sleep(time.Millisecond)
				}
				settleFrames()
			}
			w.win.Option(gioapp.Size(unit.Dp(500), unit.Dp(340)))
			wait(image.Pt(500, 340), false)
			settleFrames()
			beforeScroll := synchronous.Load()
			for i := range 60 {
				done := scrollFrames.Load()
				core.Update(func() {
					step := i
					if step >= 30 {
						step = 59 - step
					}
					scrollOffset, pendingScroll = float32(step*30), true
				})
				deadline := time.Now().Add(600 * time.Millisecond)
				for scrollFrames.Load() == done {
					if time.Now().After(deadline) {
						fail("glass=%v scroll %d stalled", glass, i)
					}
					time.Sleep(time.Millisecond)
				}
				time.Sleep(16 * time.Millisecond)
			}
			if synchronous.Load() != beforeScroll {
				fail("fixed-size scrolling used synchronous resize presentation")
			}
			// Stable-size frames must return to asynchronous presentation,
			// including recovery from the 1x1 idle drawable pool.
			phase = "stable"
			core.Update(func() {})
			stable := image.Pt(500, 340)
			time.Sleep(100 * time.Millisecond)
			wait(stable, false)
			if glass {
				onView(func(view appkit.ID) {
					if appkit.SendBool(appkit.Send(view, "layer"), "presentsWithTransaction") {
						fail("stable-size glass frame still uses transaction presentation")
					}
				})
			}
			time.Sleep(1500 * time.Millisecond)
			phase = "idle"
			wait(stable, true)
			idleFrames := frames.Load()
			cpuTime := func() time.Duration {
				var usage syscall.Rusage
				if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
					fail("getrusage: %v", err)
				}
				return time.Duration(usage.Utime.Nano() + usage.Stime.Nano())
			}
			cpuStart, idleStart := cpuTime(), time.Now()
			time.Sleep(2 * time.Second)
			idleCPU := 100 * float64(cpuTime()-cpuStart) / float64(time.Since(idleStart))
			if frames.Load() != idleFrames {
				fail("static window keeps drawing while idle")
			}
			before := synchronous.Load()
			phase = "resume"
			core.Update(func() {})
			wait(stable, false)
			time.Sleep(50 * time.Millisecond)
			if synchronous.Load() != before {
				fail("idle recovery was mistaken for a window resize")
			}
			fmt.Printf("glass=%v: 60 native resizes, slowest %v; 60 center resizes, slowest %v; 4 native zoom/restore transitions; 60 scroll steps, sync presents=%d, idle frames=0, idle process CPU=%.2f%%\n", glass, slowest, centerSlowest, synchronous.Load(), idleCPU)
			w.Close()
		}
		fmt.Println("PASS: resize gravity, full resolution, transaction present, stable-size recovery and idle")
		keeper.Close()
	}()
	Main()
}
