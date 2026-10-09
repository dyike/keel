package el

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
	"time"
)

func TestScrollingBarsExpireWithoutInterceptingContent(t *testing.T) {
	for _, horizontal := range []bool{false, true} {
		for _, scale := range []int{1, 2} {
			now := time.Now()
			clicks := 0
			var cx *Context
			root := Root(viewFunc(func(c *Context) Element {
				cx = c
				child := Div().W(Dp(100)).H(Dp(400)).OnClick(func() { clicks++ })
				box := Div().ID("box").W(Dp(100)).H(Dp(100)).Scrollbars(ScrollbarScrolling)
				if horizontal {
					box.ScrollX()
					child.W(Dp(400)).H(Dp(100))
				} else {
					box.ScrollY()
				}
				return Div().Items(Start).Child(box.Child(child))
			}))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Now = now
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints.Max = image.Pt(200*scale, 200*scale)
				root.Layout(gtx)
			})
			x, y := float32(95*scale), float32(70*scale)
			if horizontal {
				x, y = y, x
			}
			h.Click(x, y)
			if clicks != 1 {
				t.Fatal("initially hidden bar intercepted click")
			}
			if horizontal {
				cx.ScrollToX("box", 40)
			} else {
				cx.ScrollTo("box", 40)
			}
			h.Frame()
			h.Frame()
			h.Click(x, y)
			if clicks != 1 {
				t.Fatal("visible bar failed to capture click")
			}
			now = now.Add(ScrollbarLinger + time.Millisecond)
			h.Frame()
			h.Frame()
			h.Click(x, y)
			if clicks != 2 {
				t.Fatal("expired bar retained hit area", horizontal, scale, clicks)
			}
		}
	}
}

func TestHoverBarsCaptureAndReleaseOutsideViewport(t *testing.T) {
	var cx *Context
	clicks := 0
	root := Root(viewFunc(func(c *Context) Element {
		cx = c
		return Div().Items(Start).Child(Div().ID("box").ScrollY().Scrollbars(ScrollbarHover).W(Dp(100)).H(Dp(100)).Child(Div().H(Dp(400)).W(Dp(100)).OnClick(func() { clicks++ })))
	}))
	h := uitest.New(root)
	move := func(x, y float32) {
		h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, y)})
		h.Frame()
		h.Frame()
	}
	move(150, 150)
	move(50, 50)
	h.Drag(95, 12, 150, 62)
	h.Frame()
	offset, _, _ := cx.ScrollState("box")
	if offset < 190 || clicks != 0 {
		t.Fatal("hover bar lost drag outside viewport", offset, clicks)
	}
	move(150, 150)
	// One press on a hidden edge reaches content before hover can expose a bar.
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(95, 70)}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(95, 70)})
	h.Frame()
	h.Frame()
	if clicks != 1 {
		t.Fatal("hidden hover bar intercepted content", clicks)
	}
}
