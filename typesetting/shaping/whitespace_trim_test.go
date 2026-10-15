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

// Glyphs without outline metrics report zero ink. Wrapping must still count
// the last one on a line unless its text is whitespace, or each wrapped CJK
// line takes one glyph more than fits.
func TestWrapCountsLastGlyphWithoutOutlineMetrics(t *testing.T) {
	const text = "接入电表或带计量的智能插座后这里会显示全屋"
	paragraph := []rune(text)
	advance := fixed.I(10)
	glyphs := make([]Glyph, len(paragraph))
	for i := range glyphs {
		glyphs[i] = Glyph{Advance: advance, XAdvance: advance, ClusterIndex: i, RuneCount: 1, GlyphCount: 1}
	}
	run := Output{Direction: di.DirectionLTR, Advance: advance * fixed.Int26_6(len(glyphs)), Glyphs: glyphs,
		Runes: Range{Count: len(paragraph)}, Size: fixed.I(10)}
	var wrapper LineWrapper
	lines, _ := wrapper.WrapParagraph(WrapConfig{}, 35, paragraph, NewSliceIterator([]Output{run}))
	if len(lines) != 7 {
		t.Fatalf("got %d lines, want 7", len(lines))
	}
	for i, line := range lines {
		var width fixed.Int26_6
		for _, run := range line {
			width += run.Advance
		}
		if width > fixed.I(35) {
			t.Fatalf("line %d is %v wide, over the 35 allowed", i, width)
		}
	}
}

func TestAdvanceSpaceAwareTrimsOnlyWhitespace(t *testing.T) {
	advance := fixed.I(10)
	for value, want := range map[string]fixed.Int26_6{"具": 2 * advance, "A": 2 * advance, " ": advance, "　": advance} {
		paragraph := []rune("x" + value)
		run := Output{Direction: di.DirectionLTR, Advance: 2 * advance, Glyphs: []Glyph{
			{Advance: advance, Width: advance, ClusterIndex: 0, RuneCount: 1, GlyphCount: 1},
			{Advance: advance, ClusterIndex: 1, RuneCount: 1, GlyphCount: 1},
		}}
		if got := run.advanceSpaceAware(di.DirectionLTR, paragraph); got != want {
			t.Fatalf("%q: advanceSpaceAware = %v, want %v", value, got, want)
		}
	}
}
