package markdown

import (
	"image"
	"unicode"

	"github.com/dyike/keel/third_party/typesetting/segmenter"
)

// character chooses the glyph under the pointer, not the nearest caret. The
// right half of a word's last letter must still select that word.
func (s *documentSelection) character(pt image.Point) int {
	for _, part := range s.parts {
		if !pt.In(part.visible) {
			continue
		}
		local := pt.Sub(part.rect.Min)
		for _, p := range part.r.rt.pieces {
			if !local.In(p.rect) {
				continue
			}
			pos := part.start + p.start
			for _, g := range p.glyphs {
				if local.X-p.rect.Min.X < g.x+g.adv {
					return pos
				}
				pos += g.runes
			}
		}
	}
	return s.hit(pt)
}

func (s *documentSelection) unitRange(i, mode int) [2]int {
	for _, p := range s.parts {
		if i < p.start || i > p.end {
			continue
		}
		if p.start == p.end {
			return [2]int{p.start, p.end}
		}
		if mode == 3 {
			if !p.r.code {
				return [2]int{p.start, p.end}
			}
			// A code block's logical paragraphs are its source lines, not wraps.
			i = min(i, p.end-1)
			lo, hi := i, i
			for lo > p.start && s.text[lo-1] != '\n' {
				lo--
			}
			for hi < p.end && s.text[hi] != '\n' {
				hi++
			}
			return [2]int{lo, hi}
		}
		// A typeset formula is one selectable unit and copies its TeX source.
		for _, piece := range p.r.rt.pieces {
			lo, hi := p.start+piece.start, p.start+piece.start+piece.runes
			if i >= lo && i < hi && (p.r.rt.runs[piece.run].object != nil || p.r.rt.runs[piece.run].math != nil || p.r.rt.runs[piece.run].image != nil) {
				return [2]int{lo, hi}
			}
		}
		text := s.text[p.start:p.end]
		if len(text) == 0 {
			return [2]int{p.start, p.start}
		}
		at := min(i-p.start, len(text)-1)
		// Unicode words preserve combining marks, apostrophes and identifiers.
		// Dot/colon are useful additional boundaries in code (context.Context).
		var seg segmenter.Segmenter
		seg.Init(text)
		words := seg.WordIterator()
		for words.Next() {
			word := words.Word()
			lo, hi := word.Offset, word.Offset+len(word.Text)
			if at < lo || at >= hi {
				continue
			}
			if text[at] == '.' || text[at] == ':' {
				break
			}
			for j := at - 1; j >= lo; j-- {
				if text[j] == '.' || text[j] == ':' {
					lo = j + 1
					break
				}
			}
			for j := at + 1; j < hi; j++ {
				if text[j] == '.' || text[j] == ':' {
					hi = j
					break
				}
			}
			return [2]int{p.start + lo, p.start + hi}
		}
		if unicode.IsSpace(text[at]) {
			lo, hi := at, at+1
			for lo > 0 && text[lo-1] != '\n' && unicode.IsSpace(text[lo-1]) {
				lo--
			}
			for hi < len(text) && text[hi] != '\n' && unicode.IsSpace(text[hi]) {
				hi++
			}
			return [2]int{p.start + lo, p.start + hi}
		}
		// Punctuation/emoji select one grapheme, preserving ZWJ sequences.
		graphemes := seg.GraphemeIterator()
		for graphemes.Next() {
			g := graphemes.Grapheme()
			if at >= g.Offset && at < g.Offset+len(g.Text) {
				return [2]int{p.start + g.Offset, p.start + g.Offset + len(g.Text)}
			}
		}
	}
	return [2]int{i, i}
}
