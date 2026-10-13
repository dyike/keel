package el

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestDefaultInputTokenCentersInkWithLineLeading(t *testing.T) {
	for _, label := range []string{"@AKX Monorepo项目学习", "@review_gy", "@学习项目"} {
		for _, scale := range []int{1, 2} {
			for _, leading := range []float32{1, 1.6, 2} {
				var doc InputDocument
				content, _ := NewInputContent("ref", InputTokenSpan{Range: InputRange{End: 3}, Token: InputToken{ID: "ref", Text: "ref", Label: label}})
				doc.SetContent(content)
				p := theme.Current()
				p.Subtle = theme.RGB(0xdceafa)
				p.Text = theme.RGB(0x111111)
				root := Root(ViewFunc(func(cx *Context) Element {
					return cx.Themed(p, ViewFunc(func(cx *Context) Element {
						return Div().P(12).Bg(theme.RGB(0xffffff)).Child(TextArea().ID("token").Document(&doc).TextSize(14).LineHeight(leading).P(0).Border(0, color.NRGBA{}).Bg(color.NRGBA{}).AutoGrow(1, 1))
					}))
				}))
				width, height := 400*scale, 150*scale
				gpu, err := headless.NewWindow(width, height)
				if err != nil {
					t.Fatal(err)
				}
				pixels := image.NewRGBA(image.Rect(0, 0, width, height))
				h := uitest.NewFunc(func(g core.C) {
					g.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					g.Constraints.Max = image.Pt(width, height)
					root.Layout(g)
					if err := gpu.Frame(g.Ops); err != nil {
						t.Fatal(err)
					}
					if err := gpu.Screenshot(pixels); err != nil {
						t.Fatal(err)
					}
				})
				h.Frame()
				gpu.Release()
				var pill, ink image.Rectangle
				for y := 0; y < height; y++ {
					for x := 0; x < width; x++ {
						pixel := color.NRGBAModel.Convert(pixels.At(x, y)).(color.NRGBA)
						if pixel == p.Subtle {
							pill = pill.Union(image.Rect(x, y, x+1, y+1))
						}
						if pixel.R < 80 && pixel.G < 80 && pixel.B < 80 {
							ink = ink.Union(image.Rect(x, y, x+1, y+1))
						}
					}
				}
				top, bottom := ink.Min.Y-pill.Min.Y, pill.Max.Y-ink.Max.Y
				difference := top - bottom
				if difference < 0 {
					difference = -difference
				}
				if pill.Empty() || ink.Empty() || top < 1 || bottom < 1 || difference > scale+1 {
					t.Fatalf("label=%q scale=%d leading=%v: uneven padding top=%d bottom=%d pill=%v ink=%v", label, scale, leading, top, bottom, pill, ink)
				}
				if doc.Content().Text() != "ref" {
					t.Fatal("painting changed token source")
				}
			}
		}
	}
}

func BenchmarkDefaultInputToken(b *testing.B) {
	var doc InputDocument
	content, _ := NewInputContent("before ref after", InputTokenSpan{Range: InputRange{Start: 7, End: 10}, Token: InputToken{ID: "ref", Text: "ref", Label: "@AKX Monorepo项目学习"}})
	doc.SetContent(content)
	root := Root(ViewFunc(func(cx *Context) Element {
		return TextArea().Document(&doc).TextSize(14).LineHeight(1.6).W(Dp(380)).AutoGrow(1, 1)
	}))
	h := uitest.New(root)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		h.Frame()
	}
}
