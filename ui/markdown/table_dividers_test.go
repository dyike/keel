package markdown

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

// Toggle an already cached table, checking the painted separators, text
// selection/copy and unchanged geometry at both pixel densities.
func TestTableDividersToggleCachedTable(t *testing.T) {
	old := theme.Current()
	defer theme.Apply(old)
	for _, dark := range []bool{false, true} {
		for _, scale := range []int{1, 2} {
			t.Run(fmt.Sprintf("dark=%v/scale=%d", dark, scale), func(t *testing.T) {
				p := theme.Light()
				if dark {
					p = theme.Dark()
				}
				p.Frameless, p.Border = true, theme.RGB(0x7b3cf9)
				theme.Apply(p)
				gpu, err := headless.NewWindow(400*scale, 300*scale)
				if err != nil {
					t.Fatal(err)
				}
				defer gpu.Release()
				d := New("| 名称 | 数值 |\n|:--|--:|\n| 甲乙 | 一二 |\n| 丙丁 | 三四 |")
				root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
					return el.Div().Bg(theme.Bg).Child(d.Render(cx))
				}))
				var size image.Point
				var pixels *image.RGBA
				h := uitest.NewFunc(func(gtx core.C) {
					gtx.Metric.PxPerDp, gtx.Metric.PxPerSp = float32(scale), float32(scale)
					gtx.Constraints = layout.Constraints{Max: image.Pt(400*scale, 300*scale)}
					size = root.Layout(gtx).Size
					if err := gpu.Frame(gtx.Ops); err != nil {
						t.Fatal(err)
					}
					pixels = image.NewRGBA(image.Rect(0, 0, 400*scale, 300*scale))
					if err := gpu.Screenshot(pixels); err != nil {
						t.Fatal(err)
					}
				})
				count := func() int {
					n := 0
					for y := 0; y < size.Y; y++ {
						for x := 0; x < size.X; x++ {
							if color.NRGBAModel.Convert(pixels.At(x, y)).(color.NRGBA) == p.Border {
								n++
							}
						}
					}
					return n
				}
				if count() == 0 {
					t.Fatal("default table has no dividers")
				}
				originalSize, originalText, parses := size, d.RenderedText(), d.parses
				x0, y0 := textPoint(t, h, d, "甲乙", 1)
				x1, y1 := textPoint(t, h, d, "三四", 1)
				h.Drag(x0, y0, x1, y1)
				for _, on := range []bool{false, true, false} {
					d.TableDividers(on)
					h.Frame()
					if (count() > 0) != on {
						t.Fatalf("painted dividers=%d, enabled=%v", count(), on)
					}
					if size != originalSize || d.RenderedText() != originalText || d.parses != parses {
						t.Fatal("toggling dividers changed geometry, text or parsing")
					}
					assertSelectionCopy(t, h, d, "乙\t一二\n丙丁\t三")
				}
				// The appearance applies to a growing table without needing to
				// rebuild the Doc or losing its selected prefix.
				d.SetStreaming(true)
				d.Append("\n| 戊己 | 五六 |")
				h.Frame()
				if count() != 0 {
					t.Fatal("streamed table restored hidden dividers")
				}
				d.TableDividers(true)
				h.Frame()
				if count() == 0 {
					t.Fatal("streamed table did not restore dividers")
				}
			})
		}
	}
}
