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
	a.masks = map[atlasMaskKey]*atlasMask{{}: m}
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

func TestGlyphAtlasBatchesAndDeferredColors(t *testing.T) {
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	var a GlyphAtlas
	defer a.Release()
	params := text.Parameters{PxPerEm: fixed.I(13), MaxWidth: 4000}
	// More than one batch of distinct glyphs, plus duplicate requests before
	// Commit. Both deferred and already-rasterized glyphs must receive colors.
	var value []rune
	for r := rune('!'); r <= '~'; r++ {
		value = append(value, r)
	}
	sh.LayoutString(params, string(value))
	var gs []text.Glyph
	for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
		gs = append(gs, g)
	}
	a.BeginFrame(sh)
	colors := []color.NRGBA{{R: 255, A: 255}, {G: 255, A: 255}}
	// Prepare glyphs separately to avoid whole-run overlap fallback.
	for _, col := range colors {
		for _, g := range gs {
			a.Prepare(params, []text.Glyph{g}, col)
		}
	}
	a.Commit()
	if a.failed {
		t.Skip("offscreen renderer unavailable")
	}
	if a.newMasks <= glyphMaskBatch || a.maskBatches != uint64((a.newMasks+glyphMaskBatch-1)/glyphMaskBatch) {
		t.Fatalf("masks=%d batches=%d", a.newMasks, a.maskBatches)
	}
	if a.window != nil || a.readback != nil {
		t.Fatal("initial frame retained scratch resources")
	}
	if len(a.pending) != 0 {
		t.Fatal("commit retained pending masks")
	}
	for _, m := range a.masks {
		if m == nil || len(m.alpha) == 0 {
			continue
		}
		expectedColors := 2
		if atlasTintSupported {
			expectedColors = 1
		}
		if !m.ready || m.wantedCount != 0 || len(m.colors) != expectedColors {
			t.Fatalf("incomplete colors: ready=%v wanted=%d colors=%d", m.ready, m.wantedCount, len(m.colors))
		}
	}
	// Compare masks on both sides of the batch boundary with isolated
	// rasterization, so tile placement and readback offsets cannot mix glyphs.
	for _, i := range []int{0, 63, 64, len(gs) - 1} {
		var single GlyphAtlas
		single.BeginFrame(sh)
		single.Prepare(params, gs[i:i+1], colors[0])
		single.Commit()
		for key, want := range single.masks {
			got := a.masks[key]
			if want == nil || len(want.alpha) == 0 {
				continue
			}
			if got == nil || got.bounds != want.bounds || len(got.alpha) != len(want.alpha) {
				t.Fatalf("batch tile %d has wrong bounds or length", i)
			}
			// Integer translation inside the larger surface can change GPU
			// coverage rounding by one alpha unit, but not move the glyph.
			for j, w := range want.alpha {
				d := int(got.alpha[j]) - int(w)
				if d < 0 {
					d = -d
				}
				if d > 1 {
					t.Fatalf("batch tile %d pixel %d alpha difference %d", i, j, d)
				}
			}
		}
		single.Release()
	}
	batches := a.maskBatches
	a.BeginFrame(sh)
	for _, g := range gs {
		a.Prepare(params, []text.Glyph{g}, colors[0])
	}
	a.Commit()
	if a.maskBatches != batches {
		t.Fatal("warm cache submitted new raster work")
	}
	pages := len(a.pages)
	a.ReleaseScratch()
	if a.window != nil || a.readback != nil || len(a.pages) != pages {
		t.Fatal("scratch release removed cache or retained renderer")
	}
	a.BeginFrame(sh)
	for _, g := range gs {
		a.Prepare(params, []text.Glyph{g}, colors[0])
	}
	a.Commit()
	if a.window != nil || a.maskBatches != batches {
		t.Fatal("cached glyphs recreated scratch renderer")
	}
	params.PxPerEm = fixed.I(17)
	sh.LayoutString(params, "9")
	g, ok := sh.NextGlyph()
	if !ok {
		t.Fatal("no glyph")
	}
	a.BeginFrame(sh)
	a.Prepare(params, []text.Glyph{g}, colors[0])
	a.Commit()
	if a.window == nil || a.maskBatches != batches+1 {
		t.Fatal("new glyph did not recreate scratch renderer")
	}

}

func TestGlyphAtlasTrimsOnlyTransparentPixels(t *testing.T) {
	for _, tc := range []struct {
		name string
		ink  image.Rectangle
	}{
		{"empty", image.Rectangle{}},
		{"full", image.Rect(0, 0, 9, 7)},
		{"interior", image.Rect(2, 1, 7, 5)},
		{"edge", image.Rect(0, 3, 4, 7)},
		{"single", image.Rect(8, 6, 9, 7)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bounds := image.Rect(-4, -12, 5, -5)
			m := atlasMask{bounds: bounds, alpha: make([]byte, 63)}
			for y := tc.ink.Min.Y; y < tc.ink.Max.Y; y++ {
				for x := tc.ink.Min.X; x < tc.ink.Max.X; x++ {
					m.alpha[y*9+x] = byte(1 + (x+y*9)%255)
				}
			}
			before := append([]byte(nil), m.alpha...)
			released := m.trimTransparentBorder()
			if released != len(before)-len(m.alpha) {
				t.Fatal("incorrect memory accounting")
			}
			if cap(m.alpha) != len(m.alpha) {
				t.Fatal("old oversized backing storage retained")
			}
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					want := before[(y-bounds.Min.Y)*9+x-bounds.Min.X]
					var got byte
					if image.Pt(x, y).In(m.bounds) {
						got = m.alpha[(y-m.bounds.Min.Y)*m.bounds.Dx()+x-m.bounds.Min.X]
					}
					if got != want {
						t.Fatalf("coverage changed at (%d,%d): %d -> %d", x, y, want, got)
					}
				}
			}
			if !tc.ink.Empty() && m.bounds != tc.ink.Add(bounds.Min) {
				t.Fatalf("bounds=%v", m.bounds)
			}
		})
	}
}

func TestGlyphAtlasSharesColors(t *testing.T) {
	if !atlasTintSupported {
		t.Skip("Gio build has no tinted image extension")
	}
	var a GlyphAtlas
	defer a.Release()
	a.frame = 1
	m := &atlasMask{ready: true, bounds: image.Rect(0, 0, 2, 1), alpha: []byte{255, 128}, colors: make(map[color.NRGBA]atlasLocation)}
	a.prepareColor(m, color.NRGBA{R: 255, A: 255})
	a.Commit()
	first := a.pages[0].pixels
	for i := 0; i < 256; i++ {
		a.frame++
		a.prepareColor(m, color.NRGBA{R: uint8(i), G: uint8(255 - i), A: 255})
		a.Commit()
	}
	if len(m.colors) != 1 || len(a.pages) != 1 {
		t.Fatal("colors duplicate mask storage")
	}
	if a.pages[0].pixels != first {
		t.Fatal("new colors copied an immutable page")
	}
	if a.pages[0].x != 2 {
		t.Fatal("new colors consumed atlas space")
	}
}
