package shaping

import (
	"math/rand"
	"testing"

	"github.com/go-text/typesetting/bidi"
)

// Keel patch test: text simpleLTR accepts is one left-to-right run for the
// bidi algorithm, and text with one of the excluded classes is not always.
func TestSimpleLTRMatchesBidi(t *testing.T) {
	var p bidi.Paragraph
	runs := func(text []rune) int {
		out := p.Segment(text, bidi.LeftToRight)
		return out.NumRuns()
	}
	singleRun := func(text []rune) bool {
		out := p.Segment(text, bidi.LeftToRight)
		return out.NumRuns() == 1 && out.Run(0).IsLeftToRight() && out.Run(0).End == len(text)
	}
	// Every rune, between strong left-to-right letters and next to digits.
	for r := rune(0); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		text := []rune{'a', r, '1', 'b'}
		if simpleLTR(text) && !singleRun(text) {
			t.Fatalf("%U: simpleLTR, but bidi gives %d runs", r, runs(text))
		}
	}
	// Random mixes of accepted runes.
	var accepted []rune
	for r := rune(0); r < 0x3100; r++ {
		if simpleLTR([]rune{r}) {
			accepted = append(accepted, r)
		}
	}
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		text := make([]rune, 1+rng.Intn(16))
		for j := range text {
			text[j] = accepted[rng.Intn(len(accepted))]
		}
		if !singleRun(text) {
			t.Fatalf("%q: simpleLTR, but bidi gives %d runs", string(text), runs(text))
		}
	}
	for _, s := range []string{"aשb", "a١b", "a‫b", "a⁧b", "a\nb", "aمb"} {
		if simpleLTR([]rune(s)) {
			t.Errorf("%q: simpleLTR", s)
		}
	}
}
