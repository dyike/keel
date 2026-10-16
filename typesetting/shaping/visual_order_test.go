package shaping

import (
	"github.com/dyike/keel/typesetting/di"
	"golang.org/x/image/math/fixed"
	"testing"
)

func TestPlatformVisualOrderAcrossWrappingPaths(t *testing.T) {
	for _, tc := range []struct {
		name            string
		width, maxLines int
	}{{"fitting", 100, 0}, {"wrapped", 15, 0}, {"truncated", 15, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			paragraph := []rune("a b c d")
			runs := make([]Output, len(paragraph))
			for i := range runs {
				runs[i] = Output{Direction: di.DirectionLTR, Size: fixed.I(10), Advance: fixed.I(10), Runes: Range{Offset: i, Count: 1}, Glyphs: []Glyph{{ClusterIndex: i, RuneCount: 1, GlyphCount: 1, Advance: fixed.I(10)}}}
			}
			calls, truncation := 0, false
			config := WrapConfig{Direction: di.DirectionLTR, TruncateAfterLines: tc.maxLines, BreakPolicy: Always, DisableTrailingWhitespaceTrim: true, Truncator: Output{Direction: di.DirectionLTR, Size: fixed.I(10), Advance: fixed.I(5), Glyphs: []Glyph{{RuneCount: 1, GlyphCount: 1, Advance: fixed.I(5)}}}}
			config.ResolveVisualOrder = func(line Line, hasTruncator bool) {
				calls++
				truncation = truncation || hasTruncator
				for i := range line {
					line[i].VisualIndex = int32(len(line) - i - 1)
				}
			}
			var wrapper LineWrapper
			lines, truncated := wrapper.WrapParagraph(config, tc.width, paragraph, NewSliceIterator(runs))
			if calls == 0 {
				t.Fatal("platform ordering was bypassed")
			}
			if tc.maxLines > 0 && (!truncation || truncated == 0) {
				t.Fatal("truncator was not distinguished from source runs")
			}
			for _, line := range lines {
				for i, run := range line {
					if run.VisualIndex != int32(len(line)-i-1) {
						t.Fatal("wrapper overwrote platform ordering")
					}
					if i > 0 && run.Runes.Offset < line[i-1].Runes.Offset {
						t.Fatal("visual ordering changed logical text order")
					}
				}
			}
		})
	}
}
