package el

import (
	"image"
	"image/color"
	"strconv"
	"strings"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestNumericLabelAtlasPixels(t *testing.T) {
	w, err := headless.NewWindow(400, 160)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	var atlas theme.GlyphAtlas
	defer atlas.Release()
	var value, mode string
	var align Align
	var width float32
	view := ViewFunc(func(*Context) Element {
		col := color.NRGBA{R: 230, G: 190, B: 160, A: 255}
		if mode == "translucent" {
			col.A = 128
		}
		box := Div().W(Dp(width)).H(Dp(50)).P(3).TextColor(col).Child(Text(value).TextSize(12).TextAlign(align))
		switch mode {
		case "disabled":
			box.Disabled(true)
		case "opacity":
			box.Opacity(0.5)
		case "scroll":
			box.ScrollY()
		case "decorated":
			box.Decorate(func(gtx core.C, draw func()) {
				tr := op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(1.25, 1.25)).Offset(f32.Pt(0.25, 0.5))).Push(gtx.Ops)
				draw()
				tr.Pop()
			})
		}
		return Div().Bg(color.NRGBA{R: 20, G: 30, B: 40, A: 255}).Child(box)
	})
	roots := [2]*RootWidget{Root(view), Root(view)}
	roots[1].SetTextAtlas(&atlas)
	var ops op.Ops
	for _, value = range []string{"R14·12345", "R3·98765", "中文123", "12🙂34", "one1\ntwo2", "مرحبا12", "AV fi123", "Hello", ""} {
		for _, scale := range []float32{1, 1.5, 2} {
			for _, width = range []float32{35, 190} {
				for _, align = range []Align{Start, Center, End} {
					for _, mode = range []string{"plain", "disabled", "translucent", "opacity", "scroll", "decorated"} {
						var imgs [2]*image.RGBA
						for i, root := range roots {
							ops.Reset()
							root.Layout(core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale}, Constraints: layout.Exact(image.Pt(400, 160))})
							if err := w.Frame(&ops); err != nil {
								t.Fatal(err)
							}
							imgs[i] = image.NewRGBA(image.Rect(0, 0, 400, 160))
							if err := w.Screenshot(imgs[i]); err != nil {
								t.Fatal(err)
							}
						}
						maxDelta, sumDelta, ink := 0, 0, 0
						for i, want := range imgs[0].Pix {
							if want != []byte{20, 30, 40, 255}[i%4] {
								ink++
							}
							d := int(imgs[1].Pix[i]) - int(want)
							if d < 0 {
								d = -d
							}
							maxDelta = max(maxDelta, d)
							sumDelta += d
						}
						if maxDelta > 24 || sumDelta > max(ink, 1) {
							t.Fatalf("%q scale=%v width=%v align=%v mode=%s max=%d mean=%v", value, scale, width, align, mode, maxDelta, float64(sumDelta)/float64(max(ink, 1)))
						}
					}
				}
			}
		}
	}
	if atlas.Stats().RasterDraws == 0 {
		t.Fatal("image cache was never used")
	}
	roots[1].SetTextAtlas(nil)
	if roots[1].e.textAtlas != nil || len(roots[1].e.atlasLabels) != 0 || roots[1].e.atlasGlyphs != nil {
		t.Fatal("disabling the atlas retained prepared frame data")
	}
}

func TestNumericLabelAtlasFrameBudget(t *testing.T) {
	var atlas theme.GlyphAtlas
	defer atlas.Release()
	count, value := 530, strings.Repeat("12", 32)
	root := Root(ViewFunc(func(*Context) Element {
		box := Div()
		for range count {
			box.Child(Text(value).TextSize(6).H(Dp(8)))
		}
		return box
	}))
	root.SetTextAtlas(&atlas)
	var ops op.Ops
	frame := func() {
		ops.Reset()
		root.Layout(core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Exact(image.Pt(400, 18000))})
	}
	frame()
	if len(root.e.atlasGlyphs) > maxAtlasLabelGlyphs || len(root.e.atlasLabels) >= count || len(root.e.atlasLabels) == 0 {
		t.Fatalf("glyph budget not enforced: glyphs=%d labels=%d", len(root.e.atlasGlyphs), len(root.e.atlasLabels))
	}
	count, value = 2100, "12"
	frame()
	if len(root.e.atlasLabels) != 2048 {
		t.Fatalf("label budget not enforced: %d", len(root.e.atlasLabels))
	}
	count = 0
	frame()
	if len(root.e.atlasLabels) != 0 || len(root.e.atlasGlyphs) != 0 {
		t.Fatal("labels from the previous frame were retained")
	}
}

func TestNumericLabelAtlasPreservesClicksAndSemantics(t *testing.T) {
	var atlas theme.GlyphAtlas
	defer atlas.Release()
	count := 0
	root := Root(ViewFunc(func(*Context) Element {
		return Div().W(Dp(100)).H(Dp(40)).OnClick(func() { count++ }).Child(Text("R" + strconv.Itoa(count) + "·123"))
	}))
	root.SetTextAtlas(&atlas)
	h := uitest.New(root)
	h.Click(20, 12)
	if count != 1 {
		t.Fatalf("got %d clicks", count)
	}
	visible := false
	for _, node := range h.Router.AppendSemantics(nil) {
		walk(node, func(n input.SemanticNode) {
			visible = visible || n.Desc.Label == "R1·123"
		})
	}
	if !visible {
		t.Fatal("updated label absent from accessibility tree")
	}
}
