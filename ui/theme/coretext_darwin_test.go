//go:build darwin && !ios

package theme

import (
	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/text"
	gotext "github.com/dyike/keel/typesetting/font"
	ot "github.com/dyike/keel/typesetting/font/opentype"
	"github.com/dyike/keel/typesetting/font/opentype/tables"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"os"
	"runtime"
	"testing"
	"weak"
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
	p := &ctPage{img: image.NewAlpha(image.Rect(0, 0, ctPageSize, ctPageSize))}
	ct.pages = []*ctPage{p}
	mask := []byte{255, 128, 0, 64}
	r, _ := p.place(image.Pt(2, 2))
	ctWrite(p, r, mask)
	if p.img.AlphaAt(0, 0) != (color.Alpha{255}) || p.img.AlphaAt(1, 0).A != 128 {
		t.Fatal("mask not stored as linear coverage")
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
	if first.AlphaAt(r3.Min.X, r3.Min.Y).A != 0 || p.img.AlphaAt(r3.Min.X, r3.Min.Y).A != 255 {
		t.Fatal("copy-on-write lost or leaked pixels")
	}
}

func TestCoreTextSmallPagesKeepBudgetAndEvictSafely(t *testing.T) {
	ct.Lock()
	defer ct.Unlock()
	pages, glyphs, frame, extra := ct.pages, ct.glyphs, ct.frame, ct.extra
	defer func() { ct.pages, ct.glyphs, ct.frame, ct.extra = pages, glyphs, frame, extra }()
	ct.pages, ct.glyphs, ct.frame, ct.extra = nil, map[ctGlyphKey]*ctGlyph{}, 1, nil
	if ctPageSize*ctPageSize*4 > 256<<10 {
		t.Fatal("one fallback tint exceeds 256 KiB")
	}
	if ctMaxPages*ctPageSize*ctPageSize != 8<<20 {
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

func TestCoreTextCollectionUsesFaceIdentity(t *testing.T) {
	const path = "/System/Library/PrivateFrameworks/FontServices.framework/Resources/Reserved/PingFangUI.ttc"
	if _, err := os.Stat(path); err != nil {
		t.Skip("reserved PingFang collection is unavailable")
	}
	// TTC face 20 is SC Medium; CoreText descriptor 20 is HK Thin on current
	// macOS. Test the actual native font, not just the requested family.
	if got := fontPostScriptName(path, 20); got != "PingFangSC-Medium" {
		t.Skipf("different system collection layout: %q", got)
	}
	for _, size := range []float64{14, 28} {
		f := ctFont(path, 20, size)
		if f == 0 {
			t.Fatal("failed to open the exact PingFang face")
		}
		actual := ctIdentity.name(uintptr(f))
		wanted := ctIdentity.stringNew(0, "PingFangSC-Medium", 0x08000100)
		equal := actual != 0 && ctIdentity.equal(actual, wanted)
		if actual != 0 {
			ctIdentity.release(actual)
		}
		ctIdentity.release(wanted)
		if !equal {
			t.Fatal("rendering selected another region or weight")
		}
	}
	if openCTFontFace(path, 999, 14) != 0 {
		t.Fatal("invalid TTC index selected a substitute")
	}
}

func TestCoreTextSFUsesExactPostScriptFace(t *testing.T) {
	const path = "/System/Library/Fonts/SFNS.ttf"
	if _, err := os.Stat(path); err != nil {
		t.Skip("SFNS is unavailable")
	}
	name := fontPostScriptName(path, 0)
	if name != ".SFNS-Regular" {
		t.Skipf("different system font: %q", name)
	}
	f := ctFont(path, 0, 28)
	if f == 0 {
		t.Fatal("SF regular could not be matched")
	}
	actual := ctIdentity.name(uintptr(f))
	wanted := ctIdentity.stringNew(0, name, 0x08000100)
	equal := actual != 0 && ctIdentity.equal(actual, wanted)
	if actual != 0 {
		ctIdentity.release(actual)
	}
	ctIdentity.release(wanted)
	if !equal {
		t.Fatal("SF rendered a different named instance")
	}
}

func TestCoreTextPingFangRegularAndSemiboldInstances(t *testing.T) {
	const path = "/System/Library/PrivateFrameworks/FontServices.framework/Resources/Reserved/PingFangUI.ttc"
	file, err := os.Open(path)
	if err != nil {
		t.Skip("variable PingFang unavailable")
	}
	defer file.Close()
	loaders, err := ot.NewLoaders(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"PingFangSC-Regular", "PingFangSC-Semibold"} {
		index, instance := -1, 0
		for i, ld := range loaders {
			raw, err := ld.RawTable(ot.MustNewTag("fvar"))
			if err != nil {
				continue
			}
			fv, _, err := tables.ParseFvar(raw)
			if err != nil {
				t.Fatal(err)
			}
			for j := range fv.Instances {
				if fontPostScriptName(path, i, j+1) == name {
					index, instance = i, j+1
					break
				}
			}
			if index >= 0 {
				break
			}
		}
		if index < 0 {
			t.Skipf("reserved fixture lacks named instance %s", name)
		}
		ft, err := gotext.NewFont(loaders[index])
		if err != nil {
			t.Fatal(err)
		}
		gid, ok := gotext.NewFace(ft).NominalGlyph('中')
		if !ok {
			t.Fatal("fixture lacks Chinese glyph")
		}
		f := ctFont(path, index, 28, instance)
		if f == 0 {
			t.Fatal("native font missing", name)
		}
		actual := ctIdentity.name(uintptr(f))
		wanted := ctIdentity.stringNew(0, name, 0x08000100)
		equal := actual != 0 && ctIdentity.equal(actual, wanted)
		if actual != 0 {
			ctIdentity.release(actual)
		}
		ctIdentity.release(wanted)
		if !equal {
			t.Fatal("wrong native weight", name)
		}
		b, mask, ok := coreTextMask(path, index, 28, uint16(gid), 0, 0, true, instance)
		if !ok || b.Empty() || len(mask) == 0 {
			t.Fatal("named glyph rendered blank", name)
		}
	}
}

// Installed MobileAsset fonts can legitimately be static TTC faces and take
// precedence over the variable UI collection. Preserve this integration check
// without requiring a named-instance index for a non-variable font.
func TestCoreTextSystemPingFangWeights(t *testing.T) {
	const path = "/System/Library/PrivateFrameworks/FontServices.framework/Resources/Reserved/PingFangUI.ttc"
	if _, err := os.Stat(path); err != nil {
		t.Skip("PingFang unavailable")
	}
	sh := new(text.Shaper)
	for _, tc := range []struct {
		weight font.Weight
		name   string
	}{{font.Normal, "PingFangSC-Regular"}, {font.SemiBold, "PingFangSC-Semibold"}} {
		sh.LayoutString(text.Parameters{Font: font.Font{Typeface: "PingFang SC", Weight: tc.weight}, PxPerEm: fixed.I(28), MaxWidth: 1000}, "中文")
		g, ok := sh.NextGlyph()
		if !ok {
			t.Fatal("no glyph")
		}
		file, index, instance, ppem, gid, ok := sh.GlyphFileInstance(g.ID)
		if !ok {
			t.Fatal("missing exact native face identity")
		}
		f := ctFont(file, index, float64(ppem)/64, instance)
		if f == 0 {
			t.Fatal("native font missing", tc.name)
		}
		actual := ctIdentity.name(uintptr(f))
		wanted := ctIdentity.stringNew(0, tc.name, 0x08000100)
		equal := actual != 0 && ctIdentity.equal(actual, wanted)
		if actual != 0 {
			ctIdentity.release(actual)
		}
		ctIdentity.release(wanted)
		if !equal {
			t.Fatal("wrong installed native weight", tc.name)
		}
		b, mask, ok := coreTextMask(file, index, 28, gid, 0, 0, true, instance)
		if !ok || b.Empty() || len(mask) == 0 {
			t.Fatal("installed glyph rendered blank", tc.name)
		}
	}
}

func TestPlatformPreparationKeepsGoTablesLazy(t *testing.T) {
	paths := []string{"/System/Library/Fonts/Menlo.ttc", "/System/Library/Fonts/Hiragino Sans GB.ttc"}
	set, err := PrepareFontFiles(func(f font.Font) bool { return true }, paths...)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.faces) < 2 {
		t.Fatal("required font faces unavailable")
	}
	for _, f := range set.faces {
		p := f.Face.(*parsedFace)
		if p.f != nil {
			t.Fatal("preparation parsed glyph/shaping tables eagerly")
		}
		if path, index := p.Source(); path == "" || index < 0 {
			t.Fatal("lost font source identity")
		}
	}
	sh := set.shaper
	sh.LayoutString(text.Parameters{Font: font.Font{Typeface: "Menlo, Hiragino Sans GB, Apple Symbols, Apple Color Emoji"}, PxPerEm: fixed.I(24), MaxWidth: 2000}, "Hello 中文 ⠋⠙⠹ 😀")
	var gs []text.Glyph
	for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
		gs = append(gs, g)
		if _, _, _, ok := sh.GlyphNativeFont(g.ID); !ok {
			t.Fatal("prepared fonts did not use native shaping")
		}
	}
	ct.Lock()
	defer ct.Unlock()
	fontsBefore := len(ctFonts)
	if ct.glyphs == nil {
		ct.glyphs = make(map[ctGlyphKey]*ctGlyph)
	}
	for _, g := range gs {
		gl := ctGlyphFor(sh, g, 1, 0, false)
		if gl.vector {
			t.Fatalf("native glyph unexpectedly used vector fallback: %+v", g)
		}
	}
	if len(ctFonts) != fontsBefore {
		t.Fatal("rasterizer reopened fonts already owned by the native shaper")
	}
	for _, f := range set.faces {
		if f.Face.(*parsedFace).f != nil {
			t.Fatal("rendering caused duplicate Go font parsing")
		}
	}
	// The Go-only shaper still obtains independently mutable Face scratch.
	a, b := set.faces[0].Face.Face(), set.faces[0].Face.Face()
	if a == nil || b == nil || a == b || a.Font != b.Font {
		t.Fatal("Go fallback did not share tables with independent Face state")
	}
}

func TestCoreTextExtraCacheBounded(t *testing.T) {
	ct.Lock()
	defer ct.Unlock()
	glyphs, extra := ct.glyphs, ct.extra
	defer func() { ct.glyphs, ct.extra = glyphs, extra }()
	ct.glyphs, ct.extra = make(map[ctGlyphKey]*ctGlyph), nil
	// Blank/fallback markers are bounded even though they occupy no atlas page.
	for i := 0; i < ctExtraEntries*3; i++ {
		ctRememberExtra(ctGlyphKey{id: text.GlyphID(i)}, &ctGlyph{blank: true}, 0)
	}
	if len(ct.glyphs) != ctExtraEntries || ct.extra.order.Len() != ctExtraEntries {
		t.Fatal("non-atlas metadata exceeded its bound")
	}
	// Pixel budget bounds color images independently of the entry count.
	for i := 0; i < 20; i++ {
		ctRememberExtra(ctGlyphKey{id: text.GlyphID(i + 10000)}, &ctGlyph{}, 256<<10)
	}
	if ct.extra.bytes > ctExtraBudget {
		t.Fatal("color glyph byte budget exceeded")
	}
	if ct.glyphs[ctGlyphKey{id: 10000}] != nil {
		t.Fatal("old color glyph was not evicted")
	}
}

func TestCoreTextCacheDoesNotRetainShaper(t *testing.T) {
	ct.Lock()
	defer ct.Unlock()
	glyphs, extra := ct.glyphs, ct.extra
	defer func() { ct.glyphs, ct.extra = glyphs, extra }()
	ct.glyphs, ct.extra = make(map[ctGlyphKey]*ctGlyph), nil
	put := func() weak.Pointer[text.Shaper] {
		sh := text.NewShaper(text.WithPlatformFonts())
		key := ctGlyphKey{sh: weak.Make(sh), id: 123}
		ctRememberExtra(key, &ctGlyph{blank: true}, 0)
		return key.sh
	}
	key := put()
	for i := 0; i < 5; i++ {
		runtime.GC()
		if key.Value() == nil {
			return
		}
	}
	t.Fatal("a cached glyph kept a closed window's shaper alive")
}

func TestLoadFontsKeepsNativeMetadataAndOwnsBytes(t *testing.T) {
	old, shaper := loaded, Material.Shaper
	defer func() { loaded, Material.Shaper = old, shaper }()
	loaded = nil
	data := append([]byte(nil), goregular.TTF...)
	if err := LoadFonts(data); err != nil {
		t.Fatal(err)
	}
	clear(data)
	if len(loaded) != 1 {
		t.Fatal("memory collection was lost")
	}
	p := loaded[0].Face.(*parsedFace)
	if p.f != nil {
		t.Fatal("LoadFonts eagerly parsed native font tables")
	}
	Material.Shaper.LayoutString(text.Parameters{Font: loaded[0].Font, PxPerEm: fixed.I(18), MaxWidth: 1000}, "Hello")
	g, ok := Material.Shaper.NextGlyph()
	if !ok {
		t.Fatal("memory font did not shape")
	}
	if _, _, _, ok := Material.Shaper.GlyphNativeFont(g.ID); !ok {
		t.Fatal("owned memory font did not use CoreText")
	}
	if p.f != nil {
		t.Fatal("native shaping loaded duplicate Go tables")
	}
	if err := LoadFonts([]byte("invalid")); err == nil {
		t.Fatal("invalid font was accepted")
	}
}
