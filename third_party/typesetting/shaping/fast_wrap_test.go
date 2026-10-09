package shaping

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/fontscan"
	"golang.org/x/image/math/fixed"
)

// Keel patch test: WrapParagraphF's fast path for paragraphs that fit on
// one line returns what the full algorithm returns.
func TestWrapParagraphFastPathMatchesFullWrap(t *testing.T) {
	fm := fontscan.NewFontMap(nil)
	if err := fm.UseSystemFonts(t.TempDir()); err != nil {
		t.Skip("no system fonts:", err)
	}
	fm.SetQuery(fontscan.Query{Families: []string{"Menlo", "monospace"}})
	pieces := []string{"a", "Z", "0", "{", " ", "  ", "\t", "终", "端", "مر", "حبا", "é", "ﬁ", "👍", "​", "-", "שלום"}
	r := rand.New(rand.NewSource(1))
	var (
		seg    Segmenter
		shaper HarfbuzzShaper
		fast   LineWrapper
		full   LineWrapper
	)
	for i := 0; i < 3000; i++ {
		var s string
		for n := 1 + r.Intn(12); n > 0; n-- {
			s += pieces[r.Intn(len(pieces))]
		}
		text := []rune(s)
		dir := di.DirectionLTR
		if i%5 == 0 {
			dir = di.DirectionRTL
		}
		var outs []Output
		for _, in := range seg.Split(Input{Text: text, RunEnd: len(text), Direction: dir, Size: fixed.I(16)}, fm) {
			outs = append(outs, shaper.Shape(in))
		}
		if len(outs) < 2 {
			continue // upstream's single-run path, which never trims
		}
		for _, c := range []struct {
			cnf   WrapConfig
			width fixed.Int26_6
		}{{WrapConfig{}, fixed.I(1 << 20)}, {WrapConfig{DisableTrailingWhitespaceTrim: true}, fixed.I(1 << 20)}, {WrapConfig{TruncateAfterLines: 2}, fixed.I(1 << 20)}, {WrapConfig{}, fixed.I(40)}} {
			cnf, width := c.cnf, c.width
			got, gotTrunc := fast.WrapParagraphF(cnf, width, text, NewSliceIterator(deepCopyLines([]Line{outs})[0]))
			got = deepCopyLines(got)
			full.Prepare(cnf, text, NewSliceIterator(deepCopyLines([]Line{outs})[0]))
			var want []Line
			wantTrunc := 0
			for done := false; !done; {
				var line WrappedLine
				line, done = full.WrapNextLineF(width)
				if line.Line != nil {
					want = append(want, deepCopyLines([]Line{line.Line})...)
				}
				wantTrunc = line.Truncated
			}
			if gotTrunc != wantTrunc || !reflect.DeepEqual(got, want) {
				t.Fatalf("%q %+v: fast path and full wrapping differ:\n%+v\n%+v", s, cnf, got, want)
			}
		}
	}
}

func deepCopyLines(lines []Line) []Line {
	out := make([]Line, len(lines))
	for i, l := range lines {
		out[i] = append(Line(nil), l...)
		for j := range out[i] {
			out[i][j].Glyphs = append([]Glyph(nil), out[i][j].Glyphs...)
		}
	}
	return out
}
