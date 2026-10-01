package kit

import (
	"math"
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// NumberInputView edits a number with − and + buttons. Typing may pass
// through out-of-range text; Enter or leaving the field clamps it to the
// range and rounds it to the step. The arrow keys move the caret, as in any
// text field, so use the buttons to step.
type NumberInputView struct {
	name              string // accessible name from a Form row when label is empty
	label, text, err  string
	value, step       float64
	lo, hi            float64
	decimals          int
	disabled, focused bool
	onChange          func(float64)
}

func NumberInput(label string) *NumberInputView {
	v := &NumberInputView{label: label, step: 1, lo: math.Inf(-1), hi: math.Inf(1), decimals: -1}
	v.text = v.format(0)
	return v
}

// Range limits the value to [min, max].
func (v *NumberInputView) Range(min, max float64) *NumberInputView {
	v.lo, v.hi = math.Min(min, max), math.Max(min, max)
	v.SetValue(v.value)
	return v
}

// Step sets how far − and + move, 1 by default.
func (v *NumberInputView) Step(s float64) *NumberInputView {
	if s > 0 && !math.IsInf(s, 0) {
		v.step = s
	}
	return v
}

// Decimals fixes how many decimals are shown; -1 (default) shows as many as needed.
func (v *NumberInputView) Decimals(n int) *NumberInputView {
	v.decimals = n
	v.text = v.format(v.value)
	return v
}
func (v *NumberInputView) OnChange(fn func(float64)) *NumberInputView { v.onChange = fn; return v }
func (v *NumberInputView) Value() float64                             { return v.value }
func (v *NumberInputView) SetValue(x float64)                         { v.value = v.clamp(x); v.text = v.format(v.value) }
func (v *NumberInputView) SetDisabled(on bool)                        { v.disabled = on }
func (v *NumberInputView) SetError(msg string)                        { v.err = msg }
func (v *NumberInputView) Error() string                              { return v.err }
func (v *NumberInputView) FocusID() string                            { return autoID("number", v) + "/text" }

func (v *NumberInputView) clamp(x float64) float64 {
	if math.IsNaN(x) {
		x = 0
	}
	return math.Min(v.hi, math.Max(v.lo, x))
}

func (v *NumberInputView) format(x float64) string {
	return strconv.FormatFloat(x, 'f', v.decimals, 64)
}

func (v *NumberInputView) set(x float64) {
	x = v.clamp(x)
	v.err = ""
	v.text = v.format(x)
	if x == v.value {
		return
	}
	v.value = x
	if v.onChange != nil {
		v.onChange(x)
	}
}

// commit turns the typed text into the value, or restores the old one.
func (v *NumberInputView) commit() {
	if x, err := strconv.ParseFloat(strings.TrimSpace(v.text), 64); err == nil {
		v.set(x)
	} else {
		v.text = v.format(v.value)
	}
}

func (v *NumberInputView) Render(cx *el.Context) el.Element {
	id := autoID("number", v)
	focused := cx.FocusWithin(id)
	if v.focused && !focused {
		v.commit()
	}
	v.focused = focused
	text := locale.Current()
	step := func(d float64) func() {
		return func() {
			v.commit()
			v.set(v.value + d)
		}
	}
	minus := Button("", step(-v.step)).Name(text.Name(text.Decrease, v.a11y())).Icon(IconMinus).Variant(ButtonGhost).Size(28)
	plus := Button("", step(v.step)).Name(text.Name(text.Increase, v.a11y())).Icon(IconPlus).Variant(ButtonGhost).Size(28)
	minus.SetDisabled(v.disabled || v.value <= v.lo)
	plus.SetDisabled(v.disabled || v.value >= v.hi)
	border := theme.Border
	switch {
	case v.err != "":
		border = theme.Danger
	case focused && !v.disabled:
		border = theme.Primary
	}
	field := el.Input().ID(v.FocusID()).Name(v.a11y()).Bind(&v.text).Filter("0123456789.-").
		Border(0, theme.Border).Bg(theme.Surface).P(0).Grow().MinW(el.Dp(40)).
		OnSubmit(func(string) { v.commit() })
	box := el.Div().ID(id).WFull().Row().Items(el.Center).Gap(4).Px(4).Py(4).Rounded(6).Border(1, border).Bg(theme.Surface).
		Disabled(v.disabled).Child(minus.Render(cx), field, plus.Render(cx))
	if v.disabled {
		box.Bg(theme.Subtle)
		field.Bg(theme.Subtle)
	}
	return labelled(v.label, box, v.err)
}

func (v *NumberInputView) setName(s string) { v.name = s }
func (v *NumberInputView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
