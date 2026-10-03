package el

import (
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestLayerArrowFlipsAndReceivesPress(t *testing.T) {
	for _, side := range []Side{Top, Bottom, Left, Right} {
		open := true
		h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
			if open {
				cx.Overlay("panel", Anchored("anchor", Div().Size(Dp(80)).Rounded(8).Name("Panel")).Placement(side, Center).Arrow(true).OnDismiss(func() { open = false }))
			}
			return Div().Child(Div().ID("anchor").Absolute().Left(160).Top(120).Size(Dp(20)))
		})))
		b := nodeBounds(h, "Panel")
		x, y := float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2)
		switch side {
		case Top:
			y = float32(b.Max.Y + 3)
		case Bottom:
			y = float32(b.Min.Y - 3)
		case Left:
			x = float32(b.Max.X + 3)
		case Right:
			x = float32(b.Min.X - 3)
		}
		h.Click(x, y)
		h.Frame()
		if !open {
			t.Fatal("arrow press dismissed", side)
		}
		h.Click(1, 1)
		h.Frame()
		if open {
			t.Fatal("outside press failed", side)
		}
	}
	anchor := image.Rect(100, 280, 120, 300)
	pos, side := layerPlacement(anchor, image.Pt(80, 60), image.Pt(400, 300), Bottom, Center, 10)
	if side != Top || pos.Y != 210 {
		t.Fatal("flip", pos, side)
	}
	a := newLayerArrow(image.Rectangle{Min: pos, Max: pos.Add(image.Pt(80, 60))}, anchor, side, Center, 6, 8)
	if a.points[1].Y != 276 {
		t.Fatal("arrow did not flip", a)
	}
	for _, align := range []Align{Start, Center, End} {
		a := newLayerArrow(image.Rect(0, 0, 20, 40), image.Rect(-100, 0, -80, 10), Bottom, align, 6, 20)
		if a.points[0].X < 0 || a.points[2].X > 20 {
			t.Fatal("narrow panel arrow overflow", a)
		}
	}
}
