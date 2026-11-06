//go:build darwin && !ios

package theme

import (
	"bytes"
	"image"
	"runtime"
	"testing"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/text"
	"golang.org/x/image/math/fixed"
)

func TestNativeFontReclamationPreservesMasks(t *testing.T) {
	for _, tc := range []struct {
		family font.Typeface
		weight font.Weight
		style  font.Style
		sample string
	}{
		{"system-ui", font.Normal, font.Regular, "office AV fi"},
		{"system-ui", font.Bold, font.Italic, "Variable weight"},
		{"PingFang SC", font.Normal, font.Regular, "中文选择复制"},
		{"Menlo, Apple Color Emoji", font.Normal, font.Regular, "A🙂中"},
		{"Arial", font.Normal, font.Regular, "مرحبا بالعالم"},
	} {
		t.Run(string(tc.family)+tc.sample, func(t *testing.T) {
			sh := text.NewShaper(text.WithPlatformFonts())
			params := text.Parameters{Font: font.Font{Typeface: tc.family, Weight: tc.weight, Style: tc.style}, PxPerEm: fixed.I(30), MaxWidth: 1200}
			sh.LayoutString(params, tc.sample)
			var glyphs []text.Glyph
			for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
				glyphs = append(glyphs, g)
			}
			type raster struct {
				bounds image.Rectangle
				pixels []byte
				ok     bool
			}
			render := func(g text.Glyph, dark bool) raster {
				native, _, gid, ok := sh.BorrowGlyphNativeFont(g.ID)
				if !ok {
					t.Fatal("old glyph lost native font")
				}
				b, p, ok := coreTextNativeColorGlyph(native, gid, .25, 0)
				if !ok {
					b, p, ok = coreTextNativeMask(native, gid, .25, 0, dark)
				}
				return raster{b, p, ok}
			}
			for _, dark := range []bool{false, true} {
				want := make([]raster, len(glyphs))
				for i, g := range glyphs {
					want[i] = render(g, dark)
				}
				// Evict the original paragraph and cycle more sized faces than the font
				// cache can hold. Retained glyphs must remain independently paintable.
				for i := 0; i < 1200; i++ {
					params.PxPerEm = fixed.Int26_6(700 + i)
					sh.LayoutString(params, tc.sample)
				}
				for i, g := range glyphs {
					got := render(g, dark)
					if got.bounds != want[i].bounds || got.ok != want[i].ok || !bytes.Equal(got.pixels, want[i].pixels) {
						t.Fatalf("glyph %d pixels changed, dark=%v", i, dark)
					}
				}
			}
			runtime.KeepAlive(sh)
		})
	}
}
