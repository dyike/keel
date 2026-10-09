// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

import (
	stdcontext "context"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"testing"
	"time"

	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

func TestMain(m *testing.M) {
	if os.Getenv("GIO_METAL_IDLE_HELPER") == "1" {
		metalIdleWindowHelper()
		return
	}
	os.Exit(m.Run())
}

func TestMetalIdleWindow(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 for native Metal window")
	}
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Env = append(os.Environ(), "GIO_METAL_IDLE_HELPER=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	} else {
		t.Log(string(out))
	}
}

func metalIdleWindowHelper() {
	fail := func(msg string) { fmt.Fprintln(os.Stderr, msg); os.Exit(1) }
	frames := make(chan struct{}, 32)
	w := new(Window)
	w.Option(Title("Metal idle regression"), Size(unit.Dp(320), unit.Dp(240)))
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
				paint.FillShape(&ops, color.NRGBA{R: 230, G: 80, A: 255}, clip.UniformRRect(image.Rect(20, 20, 160, 120), 12).Op(&ops))
				e.Frame(&ops)
				select {
				case frames <- struct{}{}:
				default:
				}
			}
		}
	}()
	go func() {
		<-frames
		w.Perform(system.ActionRaise)
		w.Run(func() {
			objc.ID(objc.GetClass("NSApplication")).Send(objc.RegisterName("sharedApplication")).Send(objc.RegisterName("activateIgnoringOtherApps:"), true)
		})
		var size func(uintptr, uintptr) struct{ W, H float64 }
		addr, err := purego.Dlsym(purego.RTLD_DEFAULT, "objc_msgSend")
		if err != nil {
			fail(err.Error())
		}
		purego.RegisterFunc(&size, addr)
		sel := uintptr(objc.RegisterName("drawableSize"))
		for i := 0; i < 3; i++ {
			time.Sleep(3 * time.Second)
			w.Run(func() {
				c, ok := w.ctx.(*mtlContext)
				if !ok {
					fail("not a Metal context")
				}
				sz := size(uintptr(c.layer), sel)
				if !c.idle.trimmed || w.gpu == nil || sz.W != 1 || sz.H != 1 {
					fail(fmt.Sprintf("idle resources retained: %v trimmed=%v gpu=%v", sz, c.idle.trimmed, w.gpu != nil))
				}
			})
			for len(frames) > 0 {
				<-frames
			}
			start := time.Now()
			w.Invalidate()
			<-frames
			w.Run(func() {
				c := w.ctx.(*mtlContext)
				sz := size(uintptr(c.layer), sel)
				if c.idle.trimmed || w.gpu == nil || sz.W < 320 || sz.H < 240 {
					fail(fmt.Sprintf("drawable not restored: %v", sz))
				}
			})
			fmt.Printf("resume %d: %v\n", i, time.Since(start))
		}
		w.Perform(system.ActionClose) // cancels the newly armed trim timer
	}()
	Main()
}
