//go:build darwin && !ios

// SPDX-License-Identifier: Unlicense OR MIT
package text

import (
	"runtime"
	"testing"

	giofont "github.com/dyike/keel/gio/font"
	"golang.org/x/image/math/fixed"
)

func nativeMemoryGlyph(sh *Shaper, family string, size fixed.Int26_6) Glyph {
	sh.LayoutString(Parameters{Font: giofont.Font{Typeface: giofont.Typeface(family)}, PxPerEm: size, MaxWidth: 10000}, "A中🙂")
	g, _ := sh.NextGlyph()
	return g
}

func TestNativeFontEvictionKeepsGlyphIdentity(t *testing.T) {
	for _, family := range []string{"Menlo", "system-ui", "PingFang SC", "Apple Color Emoji"} {
		t.Run(family, func(t *testing.T) {
			sh := NewShaper(WithPlatformFonts())
			n := sh.shaper.native.(*ctShaper)
			first := nativeMemoryGlyph(sh, family, fixed.I(30))
			original, _, _, ok := sh.BorrowGlyphNativeFont(first.ID)
			if !ok {
				t.Fatal("missing native font")
			}
			ct.retain(original)
			defer ct.release(original)
			for i := 0; i < maxNativeFonts+64; i++ {
				nativeMemoryGlyph(sh, family, fixed.Int26_6(700+i))
			}
			if n.fontCount > maxNativeFonts {
				t.Fatalf("%d active native fonts", n.fontCount)
			}
			_, idx, _ := splitGlyphID(first.ID)
			if n.faces[idx].font != 0 && !n.faces[idx].pinned {
				t.Fatal("reconstructable cold font not evicted")
			}
			if family != "Apple Color Emoji" && n.faces[idx].pinned {
				t.Fatal("ordinary font could not be reconstructed")
			}
			restored, _, _, ok := sh.BorrowGlyphNativeFont(first.ID)
			if !ok || !ct.equal(original, restored) {
				t.Fatal("font identity changed after eviction")
			}
			sh.layoutCache = layoutCache{} // force shaping rather than a cached ID
			again := nativeMemoryGlyph(sh, family, fixed.I(30))
			if first.ID != again.ID {
				t.Fatal("glyph ID changed after eviction")
			}
			runtime.KeepAlive(sh)
		})
	}
}

func TestNativeFontLegacyBorrowRemainsValid(t *testing.T) {
	sh := NewShaper(WithPlatformFonts())
	n := sh.shaper.native.(*ctShaper)
	first := nativeMemoryGlyph(sh, "system-ui", fixed.I(30))
	original, _, _, _ := sh.GlyphNativeFont(first.ID)
	for i := 0; i < maxNativeFonts+64; i++ {
		nativeMemoryGlyph(sh, "system-ui", fixed.Int26_6(700+i))
	}
	_, idx, _ := splitGlyphID(first.ID)
	if n.faces[idx].font != original || !n.faces[idx].pinned {
		t.Fatal("legacy handle invalidated")
	}
	if ct.fontGetAscent(original) <= 0 {
		t.Fatal("legacy handle unusable")
	}
	runtime.KeepAlive(sh)
}

func TestNativeFontRecreatedLegacyHandleDropsDescriptor(t *testing.T) {
	sh := NewShaper(WithPlatformFonts())
	n := sh.shaper.native.(*ctShaper)
	first := nativeMemoryGlyph(sh, "Menlo", fixed.I(30))
	for i := 0; i < maxNativeFonts+64; i++ {
		nativeMemoryGlyph(sh, "Menlo", fixed.Int26_6(700+i))
	}
	_, idx, _ := splitGlyphID(first.ID)
	f := n.faces[idx]
	if f.font != 0 || f.descriptor == 0 {
		t.Fatal("test font was not evicted")
	}
	native, _, _, ok := sh.GlyphNativeFont(first.ID)
	if !ok || !f.pinned || f.font != native || f.descriptor != 0 {
		t.Fatal("permanent handle retained a redundant descriptor")
	}
	if ct.fontGetAscent(native) <= 0 {
		t.Fatal("recreated legacy handle unusable")
	}
	runtime.KeepAlive(sh)
}
