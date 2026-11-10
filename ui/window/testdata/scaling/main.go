// Exercise native fractional rendering, scrolling and resize acknowledgements.
package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/window"
)

type frame struct {
	metric unit.Metric
	step   int
	offset float32
}
type probe struct {
	root   *el.RootWidget
	frames chan frame
	step   int
	offset float32
}

func (p *probe) Layout(gtx core.C) core.D {
	d := p.root.Layout(gtx)
	select {
	case p.frames <- frame{metric: gtx.Metric, step: p.step, offset: p.offset}:
	default:
	}
	return d
}
func fail(v any) { fmt.Println("FAIL:", v); os.Exit(1) }
func main() {
	time.AfterFunc(30*time.Second, func() { fail("resize/scroll timeout") })
	dir, err := os.MkdirTemp("", "keel-scaling-")
	if err != nil {
		fail(err)
	}
	state, err := window.LoadSizeState(filepath.Join(dir, "size.json"))
	if err != nil {
		fail(err)
	}
	p := &probe{frames: make(chan frame, 16)}
	p.root = el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.ScrollTo("content", float32(p.step*28))
		p.offset, _, _ = cx.ScrollState("content")
		body := el.Div().ID("content").H(el.Dp(300)).ScrollY()
		for i := 0; i < 80; i++ {
			body.Child(el.Div().NoShrink().P(8).Child(kit.Button(fmt.Sprintf("Row %d · 中文", i), nil).Render(cx)))
		}
		return body
	}))
	w := window.Open(window.Options{Title: "Keel scaling regression", Width: 801, Height: 603, Frameless: true, SizeState: state, Content: p})
	go func() {
		first := <-p.frames
		<-time.After(200 * time.Millisecond)
		fmt.Printf("initial scale: dp=%g sp=%g\n", first.metric.PxPerDp, first.metric.PxPerSp)
		for step := 1; step <= 20; step++ {
			next := step
			core.Update(func() { p.step = next })
			size := image.Pt(801+(step%3)*20, 603+(step%2)*20)
			if err := w.Resize(size.X, size.Y); err != nil {
				fail(err)
			}
			for {
				f := <-p.frames
				// Window root has standard padding; Size is the full client area.
				width, height := w.Size()
				if f.step == step && width > 0 && height > 0 && (os.Getenv("KEEL_RESIZE_REQUIRED") != "1" || image.Pt(width, height) == size) {
					if step == 20 && f.offset <= 0 {
						fail("scroll position did not advance")
					}
					fmt.Printf("step=%d requested=%v actual=%v dp=%g\n", step, size, image.Pt(width, height), f.metric.PxPerDp)
					break
				}
			}
		}
		// Save outside the UI lock, then load it just as the next process would.
		if err := state.Save(); err != nil {
			fail(err)
		}
		restored, err := window.LoadSizeState(filepath.Join(dir, "size.json"))
		if err != nil {
			fail(err)
		}
		width, height := w.Size()
		ready := make(chan image.Point, 4)
		var second *window.Window
		opened := make(chan struct{})
		core.Update(func() {
			second = window.Open(window.Options{Title: "Keel restored size", Frameless: true, SizeState: restored, Content: el.Embed(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Restored") })), OnResize: func(x, y int) {
				select {
				case ready <- image.Pt(x, y):
				default:
				}
			}})
			close(opened)
		})
		<-opened
		if got := <-ready; got != image.Pt(width, height) {
			fail(fmt.Sprintf("restored %v, want %dx%d", got, width, height))
		}
		second.Close()
		fmt.Println("PASS: native fractional frames, 20 resizes and scroll positions, persisted size")
		w.Close()
	}()
	window.MainWithOptions(window.MainOptions{OnQuit: func() { os.RemoveAll(dir) }})
}
