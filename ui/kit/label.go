package kit

import (
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// LabelView displays one wrapping text flow with optional secondary text.
type LabelView struct {
	text, secondary, match string
	masked, prefix         bool
	highlight              *color.NRGBA
	style                  func(*el.TextEl)
}

func Label(text string) *LabelView                    { return &LabelView{text: text} }
func (v *LabelView) SetText(text string)              { v.text = text }
func (v *LabelView) Secondary(text string) *LabelView { v.secondary = text; return v }
func (v *LabelView) Masked(on bool) *LabelView        { v.masked = on; return v }

// Highlights colors all non-overlapping, case-sensitive exact matches. Empty clears.
func (v *LabelView) Highlights(text string) *LabelView { v.match = text; v.prefix = false; return v }

// HighlightPrefix colors a match only at the start of the primary text.
func (v *LabelView) HighlightPrefix(text string) *LabelView {
	v.match = text
	v.prefix = true
	return v
}
func (v *LabelView) HighlightColor(c color.NRGBA) *LabelView { v.highlight = &c; return v }

// Style configures typography, alignment, sizing and field association on each fresh TextEl.
func (v *LabelView) Style(fn func(*el.TextEl)) *LabelView { v.style = fn; return v }

func (v *LabelView) content() (string, []el.TextRange) {
	text := v.text
	var ranges []el.TextRange
	if v.masked {
		text = strings.Repeat("•", utf8.RuneCountInString(text))
	} else if v.match != "" {
		c := theme.PrimaryText
		if v.highlight != nil {
			c = *v.highlight
		}
		for offset := 0; offset < len(text); {
			i := strings.Index(text[offset:], v.match)
			if i < 0 || (v.prefix && offset+i != 0) {
				break
			}
			start := offset + i
			a := utf8.RuneCountInString(text[:start])
			ranges = append(ranges, el.TextRange{Start: a, End: a + utf8.RuneCountInString(v.match), Color: c})
			if v.prefix {
				break
			}
			offset = start + len(v.match)
		}
	}
	if v.secondary != "" {
		start := utf8.RuneCountInString(text) + 1
		text += " " + v.secondary
		ranges = append(ranges, el.TextRange{Start: start, End: utf8.RuneCountInString(text), Color: theme.Muted})
	}
	return text, ranges
}
func (v *LabelView) Render(cx *el.Context) el.Element {
	text, ranges := v.content()
	label := el.Text(text).Ranges(ranges...)
	if v.style != nil {
		v.style(label)
	}
	return label
}
