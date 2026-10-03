package kit

import (
	"image/color"
	"math"
	"math/big"
	"strconv"
	"strings"

	"gioui.org/io/key"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// NumberStepAction identifies the requested direction for StepBy.
type NumberStepAction int

const (
	NumberStepActionIncrement NumberStepAction = iota
	NumberStepActionDecrement
)

// NumberStepEvent requests an application-managed step. Value is the normalized
// draft (or the committed value for invalid text). Count is positive: one for
// arrows/buttons, ten for PageUp/PageDown. Action supplies the direction.
type NumberStepEvent struct {
	Value  float64
	Action NumberStepAction
	Count  int64
}

// NumberInputView edits a number with − and + buttons. Typing may pass
// through out-of-range text; Enter or leaving the field clamps it to the
// range and applies the configured decimal precision. ↑ ↓ step like the buttons; PageUp and
// PageDown move ten steps.
type NumberInputView struct {
	name              string // accessible name from a Form row when label is empty
	label, text, err  string
	value, step       float64
	lo, hi            float64
	decimals          int
	disabled, focused bool
	onChange          func(float64)
	stepBy            func(float64, NumberStepAction) float64
	prefix, suffix    el.View
	onStep            func(NumberStepEvent)
	height            float32
	plain             bool
	separator         rune
}

func NumberInput(label string) *NumberInputView {
	v := &NumberInputView{label: label, step: 1, lo: math.Inf(-1), hi: math.Inf(1), decimals: -1}
	v.text = v.format(0)
	return v
}

// Size sets the minimum field height in dp and scales its typography and
// step buttons together. Recommended heights are 28, 36 and 48. Zero restores
// theme.ControlHeight and inherited typography; invalid values are ignored.
func (v *NumberInputView) Size(dp float32) *NumberInputView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// Appearance disables/enables the default background, border, rounding and
// padding. It preserves the configured minimum height and keyboard behavior.
func (v *NumberInputView) Appearance(on bool) *NumberInputView { v.plain = !on; return v }

// Range limits the value to [min, max].
func (v *NumberInputView) Range(min, max float64) *NumberInputView {
	if math.IsNaN(min) || math.IsNaN(max) || math.IsInf(min, 0) && max == min {
		return v
	}
	v.lo, v.hi = math.Min(min, max), math.Max(min, max)
	v.SetValue(v.value)
	return v
}

// Step sets how far − and + move, 1 by default.
func (v *NumberInputView) Step(s float64) *NumberInputView {
	if s > 0 && !math.IsInf(s, 0) {
		v.step = s
		v.stepBy = nil
		v.onStep = nil
	}
	return v
}

// StepBy calculates a positive step from the normalized draft and direction.
// It runs once per action; PageUp/PageDown multiply that step by ten.
// Invalid results cancel the action without committing the draft. nil restores
// the last fixed Step. The callback should not mutate this NumberInput.
func (v *NumberInputView) StepBy(fn func(float64, NumberStepAction) float64) *NumberInputView {
	v.stepBy = fn
	v.onStep = nil
	return v
}

// OnStep replaces built-in stepping with an application callback. It does not
// commit the draft or emit OnChange; the callback may use SetValue. nil restores
// the previous fixed/dynamic strategy. Step and StepBy leave this mode.
// Range boundaries and disabled state still prevent outward actions.
func (v *NumberInputView) OnStep(fn func(NumberStepEvent)) *NumberInputView {
	v.onStep = fn
	return v
}

// Prefix places content between the decrement button and text. nil removes it.
func (v *NumberInputView) Prefix(content el.View) *NumberInputView { v.prefix = content; return v }

// Suffix places content between the text and increment button. nil removes it.
func (v *NumberInputView) Suffix(content el.View) *NumberInputView { v.suffix = content; return v }

// Decimals rounds values to 0–15 decimal places; -1 (default) keeps precision.
// Exact range endpoints take precedence when they need more decimal places.
func (v *NumberInputView) Decimals(n int) *NumberInputView {
	if n < -1 || n > 15 {
		return v
	}
	v.decimals = n
	v.SetValue(v.value)
	return v
}
func (v *NumberInputView) OnChange(fn func(float64)) *NumberInputView { v.onChange = fn; return v }
func (v *NumberInputView) Value() float64                             { return v.value }
func (v *NumberInputView) SetValue(x float64) {
	if !finiteNumber(x) {
		return
	}
	v.value = v.normalize(x)
	v.text = v.format(v.value)
}
func (v *NumberInputView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.focused = false
		v.text = v.format(v.value)
	}
}
func (v *NumberInputView) SetError(msg string) { v.err = msg }
func (v *NumberInputView) Error() string       { return v.err }
func (v *NumberInputView) FocusID() string     { return autoID("number", v) + "/text" }

func finiteNumber(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func (v *NumberInputView) normalize(x float64) float64 {
	x = math.Min(v.hi, math.Max(v.lo, x))
	if v.decimals >= 0 {
		x, _ = strconv.ParseFloat(strconv.FormatFloat(x, 'f', v.decimals, 64), 64)
	}
	x = math.Min(v.hi, math.Max(v.lo, x))
	if x == 0 {
		return 0
	} // Avoid displaying negative zero after rounding.
	return x
}

func (v *NumberInputView) format(x float64) string {
	s := strconv.FormatFloat(x, 'f', v.decimals, 64)
	rounded, _ := strconv.ParseFloat(s, 64)
	if rounded != x {
		return groupNumberText(strconv.FormatFloat(x, 'f', -1, 64), v.separator)
	}
	return groupNumberText(s, v.separator)
}

func (v *NumberInputView) set(x float64) {
	if !finiteNumber(x) {
		v.text = v.format(v.value)
		return
	}
	x = v.normalize(x)
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

// normalizeNumberText converts full-width numeric characters for editing and parsing.
func normalizeNumberText(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= '０' && r <= '９':
			return '0' + r - '０'
		case r == '＋':
			return '+'
		case r == '－':
			return '-'
		case r == '．' || r == '。':
			return '.'
		default:
			return r
		}
	}, strings.TrimSpace(s))
}

// commit turns the typed text into the value, or restores the old one.
func (v *NumberInputView) commit() {
	if x, err := strconv.ParseFloat(v.parseText(v.text), 64); err == nil {
		v.set(x)
	} else {
		v.text = v.format(v.value)
	}
}

func (v *NumberInputView) draftValue() float64 {
	if x, err := strconv.ParseFloat(v.parseText(v.text), 64); err == nil && finiteNumber(x) {
		return v.normalize(x)
	}
	return v.value
}

// move commits the draft and the increment as one user change. Decimal
// arithmetic avoids accumulating binary rounding noise for steps such as 0.1.
func (v *NumberInputView) move(count int64) {
	if v.disabled || count == 0 {
		return
	}
	base := v.draftValue()
	if v.onStep != nil {
		if count > 0 && base >= v.hi || count < 0 && base <= v.lo {
			return
		}
		action := NumberStepActionIncrement
		amount := count
		if count < 0 {
			action = NumberStepActionDecrement
			amount = -count
		}
		v.onStep(NumberStepEvent{Value: base, Action: action, Count: amount})
		return
	}
	step := v.step
	if v.stepBy != nil {
		direction := NumberStepActionIncrement
		if count < 0 {
			direction = NumberStepActionDecrement
		}
		step = v.stepBy(base, direction)
		if !finiteNumber(step) || step <= 0 || v.disabled {
			return
		}
	}
	decimal := func(x float64) *big.Rat {
		r, _ := new(big.Rat).SetString(strconv.FormatFloat(x, 'f', -1, 64))
		return r
	}
	delta := new(big.Rat).Mul(decimal(step), new(big.Rat).SetInt64(count))
	x, _ := new(big.Rat).Add(decimal(base), delta).Float64()
	if math.IsInf(x, 1) {
		x = math.Min(v.hi, math.MaxFloat64)
	}
	if math.IsInf(x, -1) {
		x = math.Max(v.lo, -math.MaxFloat64)
	}
	v.set(x)
}

type numberBlurKey struct{ id string }

func (v *NumberInputView) Render(cx *el.Context) el.Element {
	id := autoID("number", v)
	if v.focused && !cx.Enabled(id) {
		v.focused = false
		v.text = v.format(v.value)
	}
	focused := !v.disabled && cx.FocusWithin(id)
	if v.focused && !focused {
		cx.AfterEnabled(id, numberBlurKey{id}, 0, func() { v.commit(); v.focused = false })
	} else {
		v.focused = focused
	}
	height := float32(theme.ControlHeight)
	if v.height > 0 {
		height = v.height
	}
	ratio := height / float32(theme.ControlHeight)
	buttonHeight := max(1, height-8*ratio)
	padding := 3 * ratio
	text := locale.Current()
	step := func(count int64) func() { return func() { v.move(count) } }
	minus := Button("", step(-1)).ID(id + "/decrease").Name(text.Name(text.Decrease, v.a11y())).Icon(IconMinus).Variant(ButtonGhost).Size(buttonHeight)
	plus := Button("", step(1)).ID(id + "/increase").Name(text.Name(text.Increase, v.a11y())).Icon(IconPlus).Variant(ButtonGhost).Size(buttonHeight)
	draft := v.draftValue()
	minus.SetDisabled(v.disabled || draft <= v.lo)
	plus.SetDisabled(v.disabled || draft >= v.hi)
	field := fieldText(el.Input().ID(v.FocusID()).Name(v.a11y()).Bind(&v.text).Filter("0123456789.+-０１２３４５６７８９＋－．。" + v.separatorText()).Transform(v.transformEdit)).MinW(el.Dp(40)).
		OnSubmit(func(string) { v.commit() }).
		OnKey(func(e el.KeyEvent) bool {
			steps := map[key.Name]int64{key.NameUpArrow: 1, key.NameDownArrow: -1, key.NamePageUp: 10, key.NamePageDown: -10}
			count, ok := steps[key.Name(e.Name)]
			if !ok || e.Modifiers != 0 {
				return false
			}
			if e.State == el.KeyPress {
				v.move(count)
			}
			return true
		})
	// Keep the text and stable step-button identities while changing appearance.
	box := fieldFrame(id, focused, v.err != "", v.disabled, false).FocusOnPress(v.FocusID()).Gap(theme.SpaceXs * ratio).P(padding).MinH(el.Dp(height)).
		Child(minus.Render(cx))
	if v.height > 0 {
		box.TextSize(float32(theme.BodySize) * ratio)
	}
	if v.plain {
		box.Bg(color.NRGBA{}).Border(0, color.NRGBA{}).Rounded(0).P(0)
	}
	if v.prefix != nil {
		box.Child(el.Div().ID(id + "/prefix").Child(v.prefix.Render(cx)))
	}
	box.Child(field)
	if v.suffix != nil {
		box.Child(el.Div().ID(id + "/suffix").Child(v.suffix.Render(cx)))
	}
	box.Child(plus.Render(cx))
	return labelled(v.label, box, v.err)
}

func (v *NumberInputView) setName(s string) { v.name = s }
func (v *NumberInputView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}

func (v *NumberInputView) commitForm() {
	if !v.disabled {
		v.commit()
	}
}
