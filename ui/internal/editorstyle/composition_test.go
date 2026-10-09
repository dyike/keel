package editorstyle

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/third_party/gio/widget"
	"github.com/dyike/keel/third_party/gio/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestCompositionUnderlineClipsAndClears(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, single := range []bool{false, true} {
			size := image.Pt(160*scale, 55*scale)
			gpu, err := headless.NewWindow(size.X, size.Y)
			if err != nil {
				t.Fatal(err)
			}
			func() {
				defer gpu.Release()
				var ed widget.Editor
				var caret Caret
				ed.SingleLine = single
				ed.SetText("中文输入法 composing words overflowing the viewport 第二行测试")
				ed.SetCaret(ed.Len(), ed.Len()) // Both horizontal and wrapped vertical scroll.
				active := key.Range{Start: 0, End: ed.Len()}
				var bounds image.Rectangle
				var imagePixels *image.RGBA
				h := uitest.NewFunc(func(g core.C) {
					g.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					g.Constraints = layout.Exact(size)
					paint.Fill(g.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
					for {
						_, ok := ed.Update(g)
						if !ok {
							break
						}
					}
					style := material.Editor(theme.Material, &ed, "")
					style.Color = color.NRGBA{A: 255}
					caret.Layout(g, style, theme.Material.Shaper)
					bounds = Composition(g, &ed, active, color.NRGBA{R: 255, A: 255})
					if err := gpu.Frame(g.Ops); err != nil {
						t.Fatal(err)
					}
					imagePixels = image.NewRGBA(image.Rectangle{Max: size})
					if err := gpu.Screenshot(imagePixels); err != nil {
						t.Fatal(err)
					}
				})
				h.Frame()
				viewport := image.Rectangle{Max: size}
				if bounds.Empty() || !bounds.In(viewport) {
					t.Fatalf("scale %d single=%v invalid candidate bounds: %v", scale, single, bounds)
				}
				rows := map[int]int{}
				for y := 0; y < size.Y; y++ {
					for x := 0; x < size.X; x++ {
						pixel := imagePixels.RGBAAt(x, y)
						if pixel.R > 240 && pixel.G < 15 && pixel.B < 15 {
							rows[y]++
							if !image.Pt(x, y).In(bounds) {
								t.Fatal("underline outside candidate bounds", x, y, bounds)
							}
						}
					}
				}
				if len(rows) == 0 {
					t.Fatal("composition has no underline pixels")
				}
				if single && len(rows) != scale {
					t.Fatalf("underline thickness %d want %d", len(rows), scale)
				}
				for _, r := range []key.Range{{Start: -1, End: -1}, {Start: ed.Len() + 1, End: ed.Len() + 20}, {Start: 0, End: 0}} {
					active = r
					h.Frame()
					if !bounds.Empty() {
						t.Fatal("inactive range retains candidate bounds", r, bounds)
					}
					for y := 0; y < size.Y; y++ {
						for x := 0; x < size.X; x++ {
							p := imagePixels.RGBAAt(x, y)
							if p.R > 240 && p.G < 15 && p.B < 15 {
								t.Fatal("inactive range retains underline", r)
							}
						}
					}
				}
			}()
		}
	}
}
