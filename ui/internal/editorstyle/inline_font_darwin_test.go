//go:build darwin && !ios

package editorstyle

import (
	"testing"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/font/gofont"
	"github.com/dyike/keel/gio/text"
	"golang.org/x/image/math/fixed"
)

// A synthetic font must not silently disable native shaping for the entire
// editor. Keep equal rune mappings with different advances alive together:
// CoreText must not alias their font identities, even after another layout.
func TestInlineFontNativeIdentity(t *testing.T) {
	var shapers []*text.Shaper
	for _, advance := range []uint16{3750, 1250, 3750} {
		face, err := InlineFont("Keel Inline", InlineFontMetrics{1000, 800, 200}, []InlineGlyph{{0xe000, advance}, {0xf0000, advance}})
		if err != nil {
			t.Fatal(err)
		}
		shapers = append(shapers, text.NewShaper(text.WithPlatformFonts(), text.WithCollection(append([]font.FontFace{face}, gofont.Collection()...))))
	}
	for repeat := 0; repeat < 3; repeat++ {
		for i, sh := range shapers {
			for _, scale := range []int{1, 2} {
				sh.LayoutString(text.Parameters{Font: font.Font{Typeface: "Keel Inline, Go"}, PxPerEm: fixed.I(16 * scale), MaxWidth: 500}, "\ue000\U000f0000")
				for j := 0; j < 2; j++ {
					g, ok := sh.NextGlyph()
					if !ok {
						t.Fatal("missing inline glyph")
					}
					_, _, _, native := sh.BorrowGlyphNativeFont(g.ID)
					want := 60 * scale
					if i == 1 {
						want = 20 * scale
					}
					if !native || g.Advance != fixed.I(want) {
						t.Fatalf("shaper %d, scale %d: native=%v advance=%v, want %d", i, scale, native, g.Advance, want)
					}
				}
			}
		}
	}
}

func TestInlineFontNativeEditorGeometry(t *testing.T) {
	testInlineFontEditorGeometry(t, true)
}
