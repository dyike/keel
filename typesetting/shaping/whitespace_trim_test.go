package shaping

import (
	"testing"

	"github.com/dyike/keel/typesetting/di"
	"golang.org/x/image/math/fixed"
)

func TestTrimWhitespaceDoesNotDiscardMissingOutlineMetrics(t *testing.T) {
	for _, dir := range []di.Direction{di.DirectionLTR, di.DirectionRTL, di.DirectionTTB} {
		for _, value := range []string{"具", "。", "A", " ", "\t", "\u3000", " \u0301"} {
			rs := []rune(value)
			advance := fixed.I(16)
			g := Glyph{Advance: advance, XAdvance: advance, YAdvance: advance, RuneCount: len(rs), GlyphCount: 1}
			line := Line{{Direction: dir, Advance: advance, Glyphs: []Glyph{g}}}
			trimmed := trimTrailingWhitespace(dir, line, rs)
			want := fixed.Int26_6(0)
			if value == " " || value == "\t" || value == "\u3000" {
				want = advance
			}
			if trimmed != want || line[0].Glyphs[0].Advance != advance-want {
				t.Fatalf("%q dir=%v trimmed=%v glyph=%v", value, dir, trimmed, line[0].Glyphs[0].Advance)
			}
		}
	}
}
