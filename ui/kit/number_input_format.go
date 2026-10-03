package kit

import (
	"github.com/dyike/keel/ui/el"
	"strings"
	"unicode/utf8"
)

// ThousandsSeparator groups integer digits while editing and displaying values.
// Supported separators are comma, space, apostrophe and narrow/nonbreaking space.
// Zero disables grouping. Numeric Value, range and step semantics are unchanged.
func (v *NumberInputView) ThousandsSeparator(separator rune) *NumberInputView {
	switch separator {
	case 0, ',', ' ', '\'', '\u00a0', '\u202f':
	default:
		return v
	}
	v.separator = separator
	v.text = v.format(v.value)
	return v
}
func (v *NumberInputView) separatorText() string {
	if v.separator == 0 {
		return ""
	}
	return string(v.separator)
}
func (v *NumberInputView) parseText(s string) string {
	s = normalizeNumberText(s)
	if v.separator != 0 {
		s = strings.ReplaceAll(s, string(v.separator), "")
	}
	return s
}
func groupNumberText(s string, separator rune) string {
	if separator == 0 {
		return s
	}
	sign := ""
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		sign, s = s[:1], s[1:]
	}
	integer, fraction, hasDot := strings.Cut(s, ".")
	var b strings.Builder
	b.WriteString(sign)
	for i, r := range integer {
		if i > 0 && (len(integer)-i)%3 == 0 {
			b.WriteRune(separator)
		}
		b.WriteRune(r)
	}
	if hasDot {
		b.WriteByte('.')
		b.WriteString(fraction)
	}
	return b.String()
}
func (v *NumberInputView) transformEdit(edit el.InputEdit) el.InputEdit {
	original := []rune(edit.Text)
	raw := v.parseText(edit.Text)
	// Keep incomplete signs/decimal drafts, but reject misplaced signs and dots.
	dots := 0
	for i, r := range raw {
		if r == '.' {
			dots++
			if dots <= 1 {
				continue
			}
		}
		if r >= '0' && r <= '9' || i == 0 && (r == '+' || r == '-') {
			continue
		}
		return el.InputEdit{Text: v.text, Start: min(edit.Start, utf8.RuneCountInString(v.text)), End: min(edit.End, utf8.RuneCountInString(v.text))}
	}
	result := groupNumberText(raw, v.separator)
	// Map both selection ends through removed and inserted separators.
	mapEnd := func(pos int) int {
		pos = min(max(0, pos), len(original))
		count := 0
		for _, r := range original[:pos] {
			if v.separator == 0 || r != v.separator {
				count++
			}
		}
		if count == 0 {
			return 0
		}
		seen := 0
		for i, r := range []rune(result) {
			if v.separator == 0 || r != v.separator {
				seen++
			}
			if seen == count {
				return i + 1
			}
		}
		return utf8.RuneCountInString(result)
	}
	return el.InputEdit{Text: result, Start: mapEnd(edit.Start), End: mapEnd(edit.End)}
}
