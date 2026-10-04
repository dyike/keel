package el

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/font"
	"gioui.org/gpu/headless"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

// Text moved down to center its ink must not lose ink to its own box, which
// clips (it carries the text's semantics). With a fixed shift, fonts with a
// tight line box (Microsoft YaHei on Windows) lost the tails of g and y.
func TestTextShiftKeepsDescendersInBox(t *testing.T) {
	// The default faces, then the Go font alone: its line box is as tight
	// as YaHei's, so this reproduces the Windows clipping on any system.
	for _, face := range []font.Typeface{theme.Material.Face, "Go"} {
		t.Run(string(face), func(t *testing.T) {
			old := theme.Material.Face
			theme.Material.Face = face
			defer func() { theme.Material.Face = old }()
			for _, size := range []float32{12, 15, 20, 40} {
				got := redInk(t, Text("gyg").TextSize(size).TextColor(color.NRGBA{R: 0xff, A: 0xff}))
				// The same glyphs drawn by Gio in a tall, unclipped area.
				want := redInk(t, Widget(core.Func(func(gtx core.C) core.D {
					lb := material.Label(theme.Material, unit.Sp(size), "gyg")
					lb.Color = color.NRGBA{R: 0xff, A: 0xff}
					return lb.Layout(gtx)
				})).H(Dp(110)))
				if got < want*97/100 {
					t.Errorf("%vsp: %d ink pixels drawn, %d expected: the box clipped the descenders", size, got, want)
				}
			}
		})
	}
}

// redInk renders e on white and counts the pixels its red text tints.
func redInk(t *testing.T, e Element) int { return redInkScaled(t, e, 1) }

func redInkScaled(t *testing.T, e Element, scale int) int {
	t.Helper()
	gpu, err := headless.NewWindow(300*scale, 120*scale)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	root := Root(ViewFunc(func(cx *Context) Element {
		return Div().Bg(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}).Items(Start).Child(e)
	}))
	n := 0
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
		gtx.Constraints.Max = image.Pt(300*scale, 120*scale)
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rect(0, 0, 300*scale, 120*scale))
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		n = 0
		for y := 0; y < 120*scale; y++ {
			for x := 0; x < 300*scale; x++ {
				if p := img.RGBAAt(x, y); int(p.R)-int(p.G) > 40 {
					n++
				}
			}
		}
	})
	h.Frame()
	return n
}
