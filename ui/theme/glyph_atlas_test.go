package theme

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/gpu/headless"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"
)

func TestGlyphAtlasPixels(t *testing.T) {
	w, err := headless.NewWindow(600, 150)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	sh := Material.Shaper
	var ops op.Ops
	p := GlyphAtlas{SubpixelPhases: 4}
	defer p.Release()
	for _, value := range []string{"abc0123456789", "中文终端Hello123", "AV fi á 12", "12🙂34", "مرحبا12", " "} {
		for _, scale := range []float64{1, 1.5, 2} {
			params := text.Parameters{Font: font.Font{Typeface: Material.Face}, PxPerEm: fixed.Int26_6(13 * scale * 64), MaxWidth: 4000}
			sh.LayoutString(params, value)
			var gs []text.Glyph
			for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
				gs = append(gs, g)
			}
			for _, col := range []color.NRGBA{{R: 230, G: 190, B: 160, A: 255}, {R: 42, G: 146, B: 102, A: 255}, {R: 230, G: 190, B: 160, A: 128}} {
				var imgs [2]*image.RGBA
				for i := range imgs {
					ops.Reset()
					paint.Fill(&ops, color.NRGBA{R: 36, G: 39, B: 42, A: 255})
					tr := op.Affine(f32.AffineId().Offset(f32.Pt(10, 70))).Push(&ops)
					if i == 0 {
						v := GlyphPainter{FragmentSize: 1, SubpixelPhases: 4}
						v.Paint(&ops, sh, params, gs, col)
					} else {
						p.BeginFrame(sh)
						p.Prepare(params, gs, col)
						p.Commit()
						p.Paint(&ops, params, gs, col)
					}
					tr.Pop()
					if err := w.Frame(&ops); err != nil {
						t.Fatal(err)
					}
					imgs[i] = image.NewRGBA(image.Rect(0, 0, 600, 150))
					if err := w.Screenshot(imgs[i]); err != nil {
						t.Fatal(err)
					}
				}
				maxDelta, sumDelta, ink := 0, 0, 0
				for i, want := range imgs[0].Pix {
					if want != []byte{36, 39, 42, 255}[i%4] {
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
					t.Fatalf("%q scale=%v color=%v max=%d mean=%v", value, scale, col, maxDelta, float64(sumDelta)/float64(max(ink, 1)))
				}
			}
		}
	}
	if p.maskBytes == 0 {
		t.Fatal("raster cache was not used")
	}
	p.Release()
	if p.maskBytes != 0 || p.window != nil || len(p.pages) != 0 {
		t.Fatal("release did not clear resources")
	}
}

func TestGlyphAtlasPagesImmutableAndBounded(t *testing.T) {
	var a GlyphAtlas
	a.frame = 1
	for i := 0; i < glyphAtlasPages; i++ {
		loc, ok := a.allocate(image.Pt(glyphAtlasSide, glyphAtlasSide))
		if !ok {
			t.Fatal("page allocation failed")
		}
		a.pages[loc.page].pixels.Pix[0] = byte(i + 1)
	}
	a.Commit()
	old := a.pages[0].pixels
	before := append([]byte(nil), old.Pix...)
	if _, ok := a.allocate(image.Pt(1, 1)); ok {
		t.Fatal("recycled a page still referenced by the current frame")
	}
	m := &atlasMask{colors: map[color.NRGBA]atlasLocation{
		{R: 1}: {page: 0}, {R: 2}: {page: 1},
	}}
	a.masks = map[glyphFragmentKey]*atlasMask{{}: m}
	a.frame++
	loc, ok := a.allocate(image.Pt(1, 1))
	if !ok || loc.page != 0 {
		t.Fatal("oldest unused page was not recycled")
	}
	if len(a.pages) != glyphAtlasPages {
		t.Fatal("page count exceeds limit")
	}
	if _, ok := m.colors[color.NRGBA{R: 1}]; ok {
		t.Fatal("stale glyph location survived page recycling")
	}
	if _, ok := m.colors[color.NRGBA{R: 2}]; !ok {
		t.Fatal("unrelated glyph location lost")
	}
	a.pages[0].pixels.Pix[0] = 100
	if !bytes.Equal(old.Pix, before) {
		t.Fatal("changed pixels held by a previous ImageOp")
	}
	a.Release()
}

func TestGlyphAtlasMaskLimitsAndShaperReset(t *testing.T) {
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	var a GlyphAtlas
	defer a.Release()
	params := text.Parameters{PxPerEm: fixed.I(13), MaxWidth: 4000}
	sh.LayoutString(params, "12")
	var gs []text.Glyph
	for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
		gs = append(gs, g)
	}
	a.BeginFrame(sh)
	a.maskBytes = glyphMaskLimit
	a.Prepare(params, gs, color.NRGBA{A: 255})
	a.Commit()
	if a.maskBytes != glyphMaskLimit || len(a.pages) != 0 {
		t.Fatal("mask budget exceeded")
	}
	other := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	a.BeginFrame(other)
	if a.maskBytes != 0 || len(a.masks) != 0 || len(a.pages) != 0 {
		t.Fatal("old font cache retained")
	}
}
