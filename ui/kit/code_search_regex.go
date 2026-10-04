package kit

import (
	"regexp"
	"unicode/utf8"
)

// Regular expressions search the whole document, so a pattern can span
// lines ("\n", "\s+", "[^;]*"). (?m) keeps ^ and $ per line and . still
// stops at a line break, as in a per-line search.

// searchRegexp compiles the active query, or nil when it is invalid.
func (v *CodeEditorView) searchRegexp() *regexp.Regexp {
	s := v.search
	expr := "(?m)" + s.query
	if !s.matchCase {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil
	}
	return re
}

// regexMatch is one match: its range and its submatch byte offsets in text.
type regexMatch struct {
	r   codeRange
	sub []int
}

// regexMatches finds up to limit non-empty matches (limit < 0: all) in the
// document text, honoring whole word. truncated reports a cut list.
func (v *CodeEditorView) regexMatches(re *regexp.Regexp, text string, limit int) (out []regexMatch, truncated bool) {
	n := -1
	if limit >= 0 {
		n = limit + 1
	}
	var locs [][]int
	for _, loc := range re.FindAllStringSubmatchIndex(text, n) {
		if loc[0] == loc[1] {
			continue
		}
		if v.search.wholeWord && !wordBoundsAt(text, loc[0], loc[1]) {
			continue
		}
		locs = append(locs, loc)
	}
	if limit >= 0 && len(locs) > limit {
		locs, truncated = locs[:limit], true
	}
	// Byte offsets to positions in one pass over the text.
	w := textWalker{text: text}
	for _, loc := range locs {
		from := w.to(loc[0])
		to := w.to(loc[1])
		out = append(out, regexMatch{codeRange{from, to}, loc})
	}
	return out, truncated
}

// wordBoundsAt reports whether text[a:b] is not part of a longer identifier.
func wordBoundsAt(text string, a, b int) bool {
	if a > 0 {
		if r, _ := utf8.DecodeLastRuneInString(text[:a]); isIdent(r) {
			return false
		}
	}
	if b < len(text) {
		if r, _ := utf8.DecodeRuneInString(text[b:]); isIdent(r) {
			return false
		}
	}
	return true
}

// textWalker turns increasing byte offsets into line and rune column.
type textWalker struct {
	text string
	at   int
	pos  codePos
}

func (w *textWalker) to(offset int) codePos {
	for w.at < offset {
		r, size := utf8.DecodeRuneInString(w.text[w.at:])
		w.at += size
		if r == '\n' {
			w.pos = codePos{w.pos.line + 1, 0}
		} else {
			w.pos.col++
		}
	}
	return w.pos
}

// regexReplacement expands the replacement template for the match at r,
// with $1 and named groups taken in the context of the whole document.
func (v *CodeEditorView) regexReplacement(re *regexp.Regexp, text string, m regexMatch) string {
	return string(re.ExpandString(nil, v.search.replacement, text, m.sub))
}
