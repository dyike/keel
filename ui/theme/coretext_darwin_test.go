//go:build darwin && !ios

package theme

import (
	"image"
	"image/color"
	"testing"
)

const ctTestFont = "/System/Library/Fonts/Helvetica.ttc"

func TestCoreTextMask(t *testing.T) {
	// Helvetica's 'H' is glyph 43; any outline glyph serves.
	var sums [2]int
	for i, dark := range []bool{true, false} {
		b, a, ok := coreTextMask(ctTestFont, 0, 28, 43, 0.25, 0, dark)
		if !ok || a == nil || b.Dx() < 10 || b.Dy() < 15 || b.Min.Y >= 0 || len(a) != b.Dx()*b.Dy() {
			t.Fatalf("dark=%v: bounds %v, %d pixels, ok=%v", dark, b, len(a), ok)
		}
		for _, v := range a {
			sums[i] += int(v)
		}
	}
	// Linear alpha for dark ink over light is larger than for light over
	// dark at the same CoreText coverage: that is the gamma it reproduces.
	if sums[0] <= sums[1] {
		t.Fatalf("dark ink alpha %d, light ink alpha %d", sums[0], sums[1])
	}
	if _, a, ok := coreTextMask(ctTestFont, 0, 28, 3, 0, 0, true); !ok || a != nil {
		t.Fatal("a space should be blank") // glyph 3 is the space
	}
	if _, _, ok := coreTextMask("/nonexistent.ttf", 0, 28, 43, 0, 0, true); ok {
		t.Fatal("a missing file must fall back to outlines")
	}
}

// A page whose image may already be on the GPU must not change in place;
// one written this frame may, as it is uploaded once at the end of the frame.
func TestCoreTextPageCopyOnWrite(t *testing.T) {
	ct.Lock()
	defer ct.Unlock()
	saved := ct.pages
	defer func() { ct.pages = saved }()
	p := &ctPage{img: image.NewRGBA(image.Rect(0, 0, ctPageSize, ctPageSize))}
	ct.pages = []*ctPage{p}
	mask := []byte{255, 128, 0, 64}
	r, _ := p.place(image.Pt(2, 2))
	ctWrite(p, r, mask)
	if p.img.RGBAAt(0, 0) != (color.RGBA{255, 255, 255, 255}) || p.img.RGBAAt(1, 0).A != 128 {
		t.Fatal("mask not stored as premultiplied white")
	}
	// Publish the page in this frame, then write again: same image.
	p.opValid, p.opFrame = true, ct.frame
	first := p.img
	r2, _ := p.place(image.Pt(2, 2))
	ctWrite(p, r2, mask)
	if p.img != first || !p.opValid {
		t.Fatal("a page published this frame was copied")
	}
	// A later frame: the next write copies, leaving the old image intact.
	ct.frame++
	r3, _ := p.place(image.Pt(2, 2))
	ctWrite(p, r3, mask)
	if p.img == first || p.opValid {
		t.Fatal("a page published earlier was changed in place")
	}
	if first.RGBAAt(r3.Min.X, r3.Min.Y).A != 0 || p.img.RGBAAt(r3.Min.X, r3.Min.Y).A != 255 {
		t.Fatal("copy-on-write lost or leaked pixels")
	}
}

func TestCoreTextSmallPagesKeepBudgetAndEvictSafely(t *testing.T) {
	ct.Lock()
	defer ct.Unlock()
	pages, glyphs, frame := ct.pages, ct.glyphs, ct.frame
	defer func() { ct.pages, ct.glyphs, ct.frame = pages, glyphs, frame }()
	ct.pages, ct.glyphs, ct.frame = nil, map[ctGlyphKey]*ctGlyph{}, 1
	if ctPageSize*ctPageSize*4 > 256<<10 {
		t.Fatal("one fallback tint exceeds 256 KiB")
	}
	if ctMaxPages*ctPageSize*ctPageSize*4 != 32<<20 {
		t.Fatal("changed total atlas budget")
	}
	for i := range ctMaxPages {
		key := ctGlyphKey{phaseX: uint8(i)}
		page, r, ok := ctPlace(image.Pt(ctPageSize-1, ctPageSize-1), key)
		if !ok || page != i || r.Dx() != ctPageSize-1 {
			t.Fatal("could not fill atlas", i, page, r)
		}
		ct.glyphs[key] = &ctGlyph{page: page}
		ct.pages[page].used = ct.frame
	}
	if _, _, ok := ctPlace(image.Pt(4, 4), ctGlyphKey{phaseY: 1}); ok {
		t.Fatal("evicted a page used by the current frame")
	}
	old := ct.pages[0].img
	ct.frame++
	page, _, ok := ctPlace(image.Pt(4, 4), ctGlyphKey{phaseY: 1})
	if !ok || page != 0 || ct.pages[0].img == old || ct.glyphs[ctGlyphKey{}] != nil {
		t.Fatal("eviction retained glyphs or mutated a published page")
	}
}
