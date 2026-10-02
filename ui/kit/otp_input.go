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
	disabled, masked  bool
	groups            int
	size              float32
	onChange          func(string)
	onComplete        func(string)
}

// OtpInput creates a field for a code of length digits (6 if length < 1).
func OtpInput(label string, length int) *OtpInputView {
	if length < 1 {
		length = 6
	}
	return &OtpInputView{label: label, length: length, groups: min(2, length), size: 48}
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

// Masked controls whether digits are concealed in both cells and semantics.
func (v *OtpInputView) Masked(on bool) *OtpInputView { v.masked = on; return v }

// Groups splits the digits into count groups (clamped to 1..length).
// Extra digits go in the first groups. Changing groups preserves the value.
func (v *OtpInputView) Groups(count int) *OtpInputView {
	v.groups = min(max(count, 1), v.length)
	return v
}

// Size sets cell height in dp; width and spacing scale with it. Invalid sizes
// are ignored. Cells shrink horizontally to fit the available width.
func (v *OtpInputView) Size(dp float32) *OtpInputView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.size = dp
	}
	return v
}
func (v *OtpInputView) SetDisabled(on bool) { v.disabled = on }
func (v *OtpInputView) SetError(msg string) { v.err = msg }
func (v *OtpInputView) Error() string       { return v.err }
func (v *OtpInputView) FocusID() string     { return autoID("otp", v) + "/text" }

func (v *OtpInputView) Render(cx *el.Context) el.Element {
	id := autoID("otp", v)
	focused := cx.FocusWithin(id)
	box, gap := v.size*5/6, theme.SpaceSm*v.size/48
	groupGap := gap * 2
	digits := []rune(v.value)
	width := float32(v.length)*box + float32(v.length-v.groups)*gap + float32(v.groups-1)*groupGap
	row := el.Div().WFull().Row()
	index := 0
	for g := 0; g < v.groups; g++ {
		count := v.length / v.groups
		if g < v.length%v.groups {
			count++
		}

		for j := 0; j < count; j++ {
			i := index
			index++
			if i > 0 {
				space := gap
				if j == 0 {
					space = groupGap
				}
				row.Child(el.Div().W(el.Dp(space)).NoShrink())
			}
			ch := ""
			if i < len(digits) {
				ch = string(digits[i])
				if v.masked {
					ch = "•"
				}
			}
			border := theme.Border
			switch {
			case v.err != "":
				border = theme.Danger
			case focused && !v.disabled && i == min(len(digits), v.length-1):
				border = theme.Primary
			}
			cell := el.Div().W(el.Dp(0)).Grow().MaxW(el.Dp(box)).H(el.Dp(v.size)).Rounded(theme.RadiusMd).Border(1, border).Bg(theme.Surface).Center().
				Child(el.Text(ch).TextSize(theme.TextXl * v.size / 48).Bold())
			if v.disabled {
				cell.Bg(theme.Subtle).TextColor(theme.Muted)
			}
			row.Child(cell)
		}
	}
	// One real text box lies invisibly over the cells: it owns focus, editing,
	// Backspace and paste; the cells only display its digits.
	clear := color.NRGBA{}
	field := el.Input().ID(v.FocusID()).Name(v.a11y()).Bind(&v.value).Filter("0123456789").MaxLen(v.length).
		Absolute().Top(0).Left(0).Right(0).H(el.Dp(v.size)).Border(0, clear).Bg(clear).TextColor(clear).P(0).
		OnChange(func(s string) {
			v.err = ""
			if v.onChange != nil {
				v.onChange(s)
			}
			if len([]rune(s)) == v.length && v.onComplete != nil {
				v.onComplete(s)
			}
		})
	if v.masked {
		field.Password()
	}
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
