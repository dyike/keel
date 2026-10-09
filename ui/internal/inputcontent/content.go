// Package inputcontent keeps editable text and atomic references together.
// Offsets are UTF-8 bytes; all token boundaries must also be grapheme boundaries.
package inputcontent

import (
	"cmp"
	"errors"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dyike/keel/third_party/typesetting/segmenter"
)

var (
	ErrToken    = errors.New("inputcontent: invalid token")
	ErrBoundary = errors.New("inputcontent: invalid grapheme boundary")
	ErrOverlap  = errors.New("inputcontent: overlapping tokens")
	ErrText     = errors.New("inputcontent: text mismatch or invalid UTF-8")
)

type Range struct{ Start, End int }
type Token struct{ ID, Text, Label string }
type Span struct {
	Range Range
	Token Token
}

// Content is immutable. Accessors return copies, including when no edit occurs.
type Content struct {
	text  string
	spans []Span
}

func (c Content) Text() string   { return c.text }
func (c Content) Tokens() []Span { return slices.Clone(c.spans) }
func (t Token) Display() string {
	if t.Label != "" {
		return t.Label
	}
	return t.Text
}
func (t Token) Validate() error {
	if strings.TrimSpace(t.ID) == "" || t.Text == "" || !singleLine(t.Text) || t.Label != "" && !singleLine(t.Label) {
		return ErrToken
	}
	return nil
}
func singleLine(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return false
		}
	}
	return true
}
func boundaries(s string) map[int]bool {
	out := map[int]bool{0: true}
	var seg segmenter.Segmenter
	seg.Init([]rune(s))
	it := seg.GraphemeIterator()
	at := 0
	for it.Next() {
		for _, r := range it.Grapheme().Text {
			at += utf8.RuneLen(r)
		}
		out[at] = true
	}
	return out
}
func New(text string, spans ...Span) (Content, error) {
	if !utf8.ValidString(text) {
		return Content{}, ErrText
	}
	owned := slices.Clone(spans)
	slices.SortStableFunc(owned, func(a, b Span) int { return cmp.Compare(a.Range.Start, b.Range.Start) })
	bs := boundaries(text)
	last := 0
	for _, s := range owned {
		if err := s.Token.Validate(); err != nil {
			return Content{}, err
		}
		r := s.Range
		if r.Start < 0 || r.End <= r.Start || r.End > len(text) || !bs[r.Start] || !bs[r.End] {
			return Content{}, ErrBoundary
		}
		if r.Start < last {
			return Content{}, ErrOverlap
		}
		if text[r.Start:r.End] != s.Token.Text {
			return Content{}, ErrText
		}
		last = r.End
	}
	return Content{text, owned}, nil
}

// WithToken annotates existing text without editing it. Plain text that merely
// matches a token's Text never becomes a reference implicitly.
func (c Content) WithToken(s Span) (Content, error) { return New(c.text, append(c.Tokens(), s)...) }

// Snap keeps carets outside token interiors. Negative/positive bias chooses the
// beginning/end; zero chooses the nearest side, with ties going to the end.
func (c Content) Snap(at, bias int) int {
	at = max(0, min(len(c.text), at))
	for _, s := range c.spans {
		r := s.Range
		if at > r.Start && at < r.End {
			if bias < 0 || bias == 0 && at-r.Start < r.End-at {
				return r.Start
			}
			return r.End
		}
	}
	return at
}

// Selection expands intersected tokens, preserving anchor/focus direction.
func (c Content) Selection(start, end int) (int, int) {
	if start == end {
		at := c.Snap(start, 0)
		return at, at
	}
	reverse := start > end
	lo, hi := min(start, end), max(start, end)
	lo, hi = max(0, min(len(c.text), lo)), max(0, min(len(c.text), hi))
	for _, s := range c.spans {
		if lo < s.Range.End && hi > s.Range.Start {
			lo = min(lo, s.Range.Start)
			hi = max(hi, s.Range.End)
		}
	}
	if reverse {
		return hi, lo
	}
	return lo, hi
}

// Replace makes one atomic edit. Partial token selections expand to whole tokens;
// insertion inside a token is rejected. A plain replacement removes references
// in the replaced interval. References after it shift by the byte delta.
func (c Content) Replace(r Range, text string, token *Token) (Content, Range, error) {
	original := r
	bs := boundaries(c.text)
	if r.Start < 0 || r.End < r.Start || r.End > len(c.text) || !bs[r.Start] || !bs[r.End] {
		return c, r, ErrBoundary
	}
	if r.Start == r.End && c.Snap(r.Start, 1) != r.Start {
		return c, r, ErrBoundary
	}
	if !utf8.ValidString(text) {
		return c, r, ErrText
	}
	if token != nil {
		if err := token.Validate(); err != nil {
			return c, r, err
		}
		if text != token.Text {
			return c, r, ErrText
		}
	}
	r.Start, r.End = c.Selection(r.Start, r.End)
	nextText := c.text[:r.Start] + text + c.text[r.End:]
	delta := len(text) - (r.End - r.Start)
	spans := make([]Span, 0, len(c.spans)+1)
	for _, s := range c.spans {
		if s.Range.End <= r.Start {
			spans = append(spans, s)
		} else if s.Range.Start >= r.End {
			s.Range.Start += delta
			s.Range.End += delta
			spans = append(spans, s)
		}
	}
	if token != nil {
		spans = append(spans, Span{Range{r.Start, r.Start + len(text)}, *token})
	}
	// Combining text inserted next to a token can merge graphemes across its edge.
	// Reject rather than retaining a reference with invalid boundaries.
	next, err := New(nextText, spans...)
	if err != nil {
		return c, original, err
	}
	caret := r.Start + len(text)
	return next, Range{caret, caret}, nil
}

// Presentation replaces reference text with display labels and provides the
// corresponding source/display byte ranges for hit testing and clipboard maps.
type Presentation struct {
	Text  string
	Spans []DisplaySpan
}
type DisplaySpan struct {
	Source, Display Range
	Token           Token
}

func (c Content) Presentation() Presentation {
	return c.present(func(_ int, token Token) string { return token.Display() })
}

// LayoutPresentation supplies editor-only replacements for atomic references,
// for example a single private-use rune shaped as an inline object. It never
// changes the stored text or token metadata. Map edits with SourceRange before
// passing them to Session.ReplaceSource; clipboard text stays source-based.
func LayoutPresentation(c Content, replacements []string) (Presentation, error) {
	if len(replacements) != len(c.spans) {
		return Presentation{}, ErrToken
	}
	for _, label := range replacements {
		if label == "" || !singleLine(label) {
			return Presentation{}, ErrToken
		}
	}
	return c.present(func(i int, _ Token) string { return replacements[i] }), nil
}

func (c Content) present(label func(int, Token) string) Presentation {
	var out strings.Builder
	result := Presentation{}
	at := 0
	for i, s := range c.spans {
		out.WriteString(c.text[at:s.Range.Start])
		start := out.Len()
		out.WriteString(label(i, s.Token))
		result.Spans = append(result.Spans, DisplaySpan{s.Range, Range{start, out.Len()}, s.Token})
		at = s.Range.End
	}
	out.WriteString(c.text[at:])
	result.Text = out.String()
	return result
}
func (p Presentation) SourceOffset(display, bias int) int {
	display = max(0, min(len(p.Text), display))
	delta := 0
	for _, s := range p.Spans {
		if display < s.Display.Start {
			break
		}
		if display <= s.Display.End {
			if display == s.Display.Start {
				return s.Source.Start
			}
			if display == s.Display.End {
				return s.Source.End
			}
			if bias < 0 || bias == 0 && display-s.Display.Start < s.Display.End-display {
				return s.Source.Start
			}
			return s.Source.End
		}
		delta += (s.Source.End - s.Source.Start) - (s.Display.End - s.Display.Start)
	}
	return display + delta
}

// DisplayOffset maps a source caret to the visible label, snapping token
// interiors using the same bias as SourceOffset.
func (p Presentation) DisplayOffset(source, bias int) int {
	sourceLen := len(p.Text)
	for _, s := range p.Spans {
		sourceLen += s.Source.End - s.Source.Start - (s.Display.End - s.Display.Start)
	}
	source = max(0, min(sourceLen, source))
	delta := 0
	for _, s := range p.Spans {
		if source < s.Source.Start {
			break
		}
		if source <= s.Source.End {
			if source == s.Source.Start {
				return s.Display.Start
			}
			if source == s.Source.End {
				return s.Display.End
			}
			if bias < 0 || bias == 0 && source-s.Source.Start < s.Source.End-source {
				return s.Display.Start
			}
			return s.Display.End
		}
		delta += s.Display.End - s.Display.Start - (s.Source.End - s.Source.Start)
	}
	return source + delta
}

// Equal compares both text and reference metadata.
func Equal(a, b Content) bool { return a.text == b.text && slices.Equal(a.spans, b.spans) }
