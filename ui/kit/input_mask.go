package kit

import (
	"strings"
	"unicode"

	"github.com/dyike/keel/ui/el"
)

type inputMaskPart struct {
	literal rune
	kind    rune
}
type inputMask struct {
	parts     []inputMaskPart
	number    bool
	separator rune
	fraction  int
}
type inputMaskResult struct {
	text, raw string
	origins   []int
	editable  []bool
	complete  bool
}

// Mask formats single-line input using # for a required ASCII digit, 9 for an
// optional digit, A for a Unicode letter and * for a letter or digit. Backslash
// quotes the next rune. Other runes are literals. Empty removes the mask.
// Invalid trailing escapes leave the existing configuration unchanged.
func (v *InputView) Mask(pattern string) *InputView {
	if v.multiline {
		return v
	}
	if pattern == "" {
		v.mask = nil
		return v
	}
	m := &inputMask{}
	escaped := false
	for _, r := range pattern {
		if escaped {
			m.parts = append(m.parts, inputMaskPart{literal: r})
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		p := inputMaskPart{literal: r}
		if strings.ContainsRune("#9A*", r) {
			p = inputMaskPart{kind: r}
		}
		m.parts = append(m.parts, p)
	}
	if escaped {
		return v
	}
	v.mask = m
	v.SetValue(v.value)
	return v
}

// NumberMask groups ASCII integer digits in threes and limits decimal places
// without floating-point conversion or rounding. Zero separator disables
// grouping; fraction -1 allows arbitrary precision, 0 permits integers only.
// Invalid separators or precision leave the existing configuration unchanged.
func (v *InputView) NumberMask(separator rune, fraction int) *InputView {
	if v.multiline || fraction < -1 || separator < 0 || separator > unicode.MaxRune || separator >= 0xD800 && separator <= 0xDFFF || separator != 0 && (unicode.IsDigit(separator) || strings.ContainsRune(".-+\r\n", separator)) {
		return v
	}
	v.mask = &inputMask{number: true, separator: separator, fraction: fraction}
	v.SetValue(v.value)
	return v
}

// UnmaskedValue returns accepted slot characters, or the ungrouped decimal text.
// With no mask it returns Value unchanged. Value and callbacks use display text.
func (v *InputView) UnmaskedValue() string {
	if v.mask == nil {
		return v.value
	}
	return v.formatMask(v.value).raw
}

// MaskComplete reports whether all required slots are filled. Number masks
// require at least one digit; a trailing decimal point is an incomplete draft.
// It is a formatting check, not domain validation such as a valid calendar date.
func (v *InputView) MaskComplete() bool {
	return v.mask == nil || v.formatMask(v.value).complete
}

func maskAccept(kind, r rune) bool {
	switch kind {
	case '#', '9':
		return r >= '0' && r <= '9'
	case 'A':
		return unicode.IsLetter(r)
	case '*':
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	return false
}
func (v *InputView) maskAllowed(r rune) bool {
	return v.filter == "" || strings.ContainsRune(v.filter, r)
}
func (v *InputView) formatMask(text string) inputMaskResult {
	if v.mask.number {
		return v.formatNumberMask(text)
	}
	in := []rune(text)
	var out, raw []rune
	var pending []struct {
		r      rune
		origin int
	}
	res := inputMaskResult{}
	pos := 0
	emit := func(r rune, origin int, editable bool) {
		out = append(out, r)
		res.origins = append(res.origins, origin)
		res.editable = append(res.editable, editable)
	}
	for index := 0; index < len(v.mask.parts); index++ {
		part := v.mask.parts[index]
		if part.kind == 0 {
			end := index
			for end < len(v.mask.parts) && v.mask.parts[end].kind == 0 {
				end++
			}
			// Consume a literal run only when it is present in full. Otherwise
			// a raw value beginning with a country-code digit can lose a slot.
			matched := pos+end-index <= len(in)
			for j := index; j < end && matched; j++ {
				matched = in[pos+j-index] == v.mask.parts[j].literal
			}
			for j := index; j < end; j++ {
				origin := -1
				if matched {
					origin = pos + j - index
				}
				pending = append(pending, struct {
					r      rune
					origin int
				}{v.mask.parts[j].literal, origin})
			}
			if matched {
				pos += end - index
			}
			index = end - 1
			continue
		}
		if v.maxLen > 0 && len(raw) >= v.maxLen {
			continue
		}
		if part.kind == '9' && pos < len(in) && !maskAccept(part.kind, in[pos]) {
			continue
		}
		for pos < len(in) && (!maskAccept(part.kind, in[pos]) || !v.maskAllowed(in[pos])) {
			pos++
		}
		if pos == len(in) {
			continue
		}
		for _, p := range pending {
			origin := p.origin
			if origin < 0 {
				origin = pos
			}
			emit(p.r, origin, false)
		}
		pending = nil
		emit(in[pos], pos, true)
		raw = append(raw, in[pos])
		pos++
		// Completeness is computed below so an earlier missing slot stays missing.
	}
	if len(raw) > 0 && len(pending) > 0 {
		// Suffix literals are emitted only when no unfilled slot precedes them.
		filled, slots := 0, 0
		for _, b := range res.editable {
			if b {
				filled++
			}
		}
		for _, p := range v.mask.parts {
			if p.kind != 0 {
				slots++
			}
		}
		if filled == slots {
			for _, p := range pending {
				origin := p.origin
				if origin < 0 {
					origin = max(0, len(in)-1)
				}
				emit(p.r, origin, false)
			}
		}
	}
	// Replay slot assignment against output: literals cannot satisfy a slot.
	res.complete = true
	j := 0
	for _, p := range v.mask.parts {
		if p.kind == 0 {
			continue
		}
		if j < len(raw) && maskAccept(p.kind, raw[j]) {
			j++
			continue
		}
		if p.kind != '9' {
			res.complete = false
		}
	}
	res.text, res.raw = string(out), string(raw)
	return res
}

func (v *InputView) formatNumberMask(text string) inputMaskResult {
	var raw []rune
	var origins []int
	dot, decimals := -1, 0
	for i, r := range []rune(text) {
		if r == '.' && v.mask.fraction == 0 {
			break
		}
		if v.maxLen > 0 && len(raw) >= v.maxLen {
			break
		}
		if !v.maskAllowed(r) {
			continue
		}
		if r == '-' && len(raw) == 0 {
			raw = append(raw, r)
			origins = append(origins, i)
			continue
		}
		if r == '.' && dot < 0 && v.mask.fraction != 0 {
			dot = len(raw)
			raw = append(raw, r)
			origins = append(origins, i)
			continue
		}
		if r < '0' || r > '9' {
			continue
		}
		if dot >= 0 {
			if v.mask.fraction >= 0 && decimals >= v.mask.fraction {
				continue
			}
			decimals++
		}
		raw = append(raw, r)
		origins = append(origins, i)
	}
	end := len(raw)
	if dot >= 0 {
		end = dot
	}
	first := 0
	if len(raw) > 0 && raw[0] == '-' {
		first = 1
	}
	res := inputMaskResult{raw: string(raw)}
	var out []rune
	for i, r := range raw {
		if v.mask.separator != 0 && i > first && i < end && (end-i)%3 == 0 {
			out = append(out, v.mask.separator)
			res.origins = append(res.origins, origins[i])
			res.editable = append(res.editable, false)
		}
		out = append(out, r)
		res.origins = append(res.origins, origins[i])
		res.editable = append(res.editable, true)
	}
	res.text = string(out)
	res.complete = len(raw) > first && raw[len(raw)-1] != '.' && (end > first || decimals > 0)
	return res
}

func (r inputMaskResult) caret(pos int) int {
	n := 0
	for _, origin := range r.origins {
		if origin < pos {
			n++
		}
	}
	return n
}
func (v *InputView) transformMask(before, after el.InputEdit) el.InputEdit {
	res := v.formatMask(after.Text)
	// Removing only an inserted literal must still advance deletion. Use the
	// previous caret to distinguish Backspace from Delete and remove one slot.
	if res.text == before.Text && before.Start == before.End && after.Start == after.End {
		old, next := []rune(before.Text), []rune(after.Text)
		if len(next) < len(old) {
			start := 0
			for start < len(next) && next[start] == old[start] {
				start++
			}
			end := len(old)
			tail := len(next)
			for end > start && tail > start && old[end-1] == next[tail-1] {
				end--
				tail--
			}
			previous := v.formatMask(before.Text)
			literal := len(previous.editable) == len(old)
			for i := start; i < end && literal; i++ {
				literal = !previous.editable[i]
			}
			if literal {
				index, delta := end, 1
				if before.Start > after.Start {
					index, delta = start-1, -1
				}
				for index >= 0 && index < len(old) && !previous.editable[index] {
					index += delta
				}
				if index >= 0 && index < len(old) {
					if index >= end {
						index -= end - start
					}
					next = append(next[:index], next[index+1:]...)
					if index < after.Start {
						after.Start--
						after.End--
					}
					after.Text = string(next)
					res = v.formatMask(after.Text)
				}
			}
		}
	}
	return el.InputEdit{Text: res.text, Start: res.caret(after.Start), End: res.caret(after.End)}
}
