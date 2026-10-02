package kit

import (
	"image/color"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// OtpInputView enters a one-time code of a fixed number of digits, one box
// per digit. Typing moves to the next box, Backspace goes back, and pasting a
// whole code fills every box. OnComplete runs when the last digit is entered.
type OtpInputView struct {
	name              string // accessible name from a Form row when label is empty
	label, value, err string
	length            int
	disabled          bool
	onChange          func(string)
	onComplete        func(string)
}

// OtpInput creates a field for a code of length digits (6 if length < 1).
func OtpInput(label string, length int) *OtpInputView {
	if length < 1 {
		length = 6
	}
	return &OtpInputView{label: label, length: length}
}
func (v *OtpInputView) OnChange(fn func(string)) *OtpInputView   { v.onChange = fn; return v }
func (v *OtpInputView) OnComplete(fn func(string)) *OtpInputView { v.onComplete = fn; return v }
func (v *OtpInputView) Value() string                            { return v.value }

// SetValue sets the digits without calling OnChange; extra or non-digit
// characters are dropped.
func (v *OtpInputView) SetValue(s string) {
	out := make([]rune, 0, v.length)
	for _, r := range s {
		if r >= '0' && r <= '9' && len(out) < v.length {
			out = append(out, r)
		}
	}
	v.value = string(out)
}
func (v *OtpInputView) SetDisabled(on bool) { v.disabled = on }
func (v *OtpInputView) SetError(msg string) { v.err = msg }
func (v *OtpInputView) Error() string       { return v.err }
func (v *OtpInputView) FocusID() string     { return autoID("otp", v) + "/text" }

func (v *OtpInputView) Render(cx *el.Context) el.Element {
	id := autoID("otp", v)
	focused := cx.FocusWithin(id)
	const box, gap = 40, 8
	digits := []rune(v.value)
	width := float32(v.length*box + (v.length-1)*gap)
	row := el.Div().WFull().Row().Gap(gap)
	for i := 0; i < v.length; i++ {
		ch := ""
		if i < len(digits) {
			ch = string(digits[i])
		}
		border := theme.Border
		switch {
		case v.err != "":
			border = theme.Danger
		case focused && !v.disabled && i == min(len(digits), v.length-1):
			border = theme.Primary
		}
		cell := el.Div().W(el.Dp(0)).Grow().MaxW(el.Dp(box)).H(el.Dp(48)).Rounded(6).Border(1, border).Bg(theme.Surface).Center().
			Child(el.Text(ch).TextSize(20).Bold())
		if v.disabled {
			cell.Bg(theme.Subtle).TextColor(theme.Muted)
		}
		row.Child(cell)
	}
	// One real text box lies invisibly over the cells: it owns focus, editing,
	// Backspace and paste; the cells only display its digits.
	clear := color.NRGBA{}
	field := el.Input().ID(v.FocusID()).Name(v.a11y()).Bind(&v.value).Filter("0123456789").MaxLen(v.length).
		Absolute().Top(0).Left(0).Right(0).H(el.Dp(48)).Border(0, clear).Bg(clear).TextColor(clear).P(0).
		OnChange(func(s string) {
			v.err = ""
			if v.onChange != nil {
				v.onChange(s)
			}
			if len([]rune(s)) == v.length && v.onComplete != nil {
				v.onComplete(s)
			}
		})
	wrap := el.Div().ID(id).W(el.Dp(width)).MaxW(el.Full).Disabled(v.disabled).Items(el.Start).Child(row, field)
	return labelled(v.label, wrap, v.err)
}

func (v *OtpInputView) setName(s string) { v.name = s }
func (v *OtpInputView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
