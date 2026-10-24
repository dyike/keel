// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && nometal

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

	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/unit"
)

func TestMain(m *testing.M) {
	if os.Getenv("GIO_GL_WINDOW_HELPER") == "1" {
		glWindowHelper()
		return
	}
	os.Exit(m.Run())
}

// TestOpenGLWindow opens a nometal window, which draws through an
// NSOpenGLContext, and resizes it. Keel patch: it covers the OpenGL window
// context under either binding (go test -tags nometal, and with purego).
func TestOpenGLWindow(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 for the native OpenGL window")
	}
	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0])
	cmd.Env = append(os.Environ(), "GIO_GL_WINDOW_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	t.Log(string(out))
}

func glWindowHelper() {
	fail := func(msg string) { fmt.Fprintln(os.Stderr, msg); os.Exit(1) }
	frames := make(chan image.Point, 32)
	w := new(Window)
	w.Option(Title("OpenGL window regression"), Size(unit.Dp(320), unit.Dp(240)), TopMost(true))
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
				paint.Fill(&ops, color.NRGBA{R: 200, G: 40, B: 50, A: 255})
				e.Frame(&ops)
				select {
				case frames <- image.Pt(int(e.Metric.PxToDp(e.Size.X)), int(e.Metric.PxToDp(e.Size.Y))):
				default:
				}
			}
		}
	}()
	go func() {
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		select {
		case <-frames:
		case <-deadline.C:
			fail("no first frame")
		}
		w.Option(Size(unit.Dp(400), unit.Dp(260)))
		for {
			select {
			case size := <-frames:
				if size == image.Pt(400, 260) {
					fmt.Println("PASS: OpenGL window draws and resizes")
					w.Perform(system.ActionClose)
					return
				}
			case <-deadline.C:
				fail("OpenGL window stopped drawing")
			}
		}
	}()
	Main()
}
