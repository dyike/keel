package theme

import (
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/gpu/headless"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"testing"
)

func TestGlyphPainterPixels(t *testing.T) {
	w, err := headless.NewWindow(400, 130)
	if err != nil {
		t.Skipf("GPU unavailable: %v", err)
	}
	defer w.Release()
	shGo := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	var ops op.Ops
	for _, sh := range []*text.Shaper{shGo, Material.Shaper} {
		for _, value := range []string{"abcdefABCDEF0123456789", "中文终端Hello123", "AV fi á 12", "12🙂34", "مرحبا12", ""} {
			for _, scale := range []float32{1, 1.5, 2} {
				params := text.Parameters{PxPerEm: fixed.Int26_6(14 * scale * 64), MaxWidth: 4000}
				params.Font.Typeface = Material.Face
				sh.LayoutString(params, value)
				var gs []text.Glyph
				for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
					gs = append(gs, g)
				}
				for _, size := range []int{1, 2, 4, 8} {
					for _, alpha := range []uint8{255, 128} {
						col := color.NRGBA{R: 230, G: 190, B: 160, A: alpha}
						imgs := [2]*image.RGBA{}
						for v := range imgs {
							ops.Reset()
							paint.Fill(&ops, color.NRGBA{R: 20, G: 30, B: 40, A: 255})
							tr := op.Affine(f32.AffineId().Offset(f32.Pt(10.25, 50.5))).Push(&ops)
							if v == 0 {
								paint.ColorOp{Color: col}.Add(&ops)
								c := clip.Outline{Path: sh.Shape(gs)}.Op().Push(&ops)
								paint.PaintOp{}.Add(&ops)
								c.Pop()
								sh.Bitmaps(gs).Add(&ops)
							} else {
								p := GlyphPainter{FragmentSize: size}
								p.Paint(&ops, sh, params, gs, col)
							}
							tr.Pop()
							if err := w.Frame(&ops); err != nil {
								t.Fatal(err)
							}
							imgs[v] = image.NewRGBA(image.Rect(0, 0, 400, 130))
							if err := w.Screenshot(imgs[v]); err != nil {
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
						if maxDelta > 24 || sumDelta > max(1, ink) {
							t.Fatalf("%q scale=%v fragment=%d alpha=%d max=%d mean=%v", value, scale, size, alpha, maxDelta, float64(sumDelta)/float64(max(1, ink)))
						}
					}
				}
			}
		}
	}
}

func TestGlyphPainterAnchorCacheBounded(t *testing.T) {
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	var p GlyphPainter
	var ops op.Ops
	for size := 1; size <= 140; size++ {
		params := text.Parameters{PxPerEm: fixed.I(size), MaxWidth: 4000}
		sh.LayoutString(params, "12")
		var gs []text.Glyph
		for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
			gs = append(gs, g)
		}
		ops.Reset()
		p.Paint(&ops, sh, params, gs, color.NRGBA{A: 255})
		if len(p.blanks) > 128 {
			t.Fatal("anchor cache exceeds bound", len(p.blanks))
		}
	}
}

func TestGlyphPainterFragmentCacheBounded(t *testing.T) {
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	params := text.Parameters{PxPerEm: fixed.I(12), MaxWidth: 4000}
	sh.LayoutString(params, "12")
	var gs []text.Glyph
	for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
		gs = append(gs, g)
	}
	var p GlyphPainter
	var ops op.Ops
	for i := 0; i < 4200; i++ {
		// Different spacing is valid for a caller-positioned terminal run.
		gs[1].X = gs[0].X + fixed.I(32+i)
		ops.Reset()
		p.Paint(&ops, sh, params, gs, color.NRGBA{A: 255})
		if len(p.fragments) > 4096 {
			t.Fatal("fragment cache exceeds bound")
		}
	}
	if len(p.fragments) != 4096 {
		t.Fatal("eviction was not exercised", len(p.fragments))
	}
}
