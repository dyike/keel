package kit

import (
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// RadioGroupView picks one of several options. Tab enters the group at the
// chosen option; arrow keys move the choice, like native radio groups.
type RadioGroupView struct {
	name       string // accessible name from a Form row when label is empty
	label      string
	options    []string
	value      string
	horizontal bool
	disabled   bool
	onChange   func(string)
}

func RadioGroup(label string, options ...string) *RadioGroupView {
	return &RadioGroupView{label: label, options: options}
}
func (v *RadioGroupView) Horizontal() *RadioGroupView                    { v.horizontal = true; return v }
func (v *RadioGroupView) OnChange(fn func(value string)) *RadioGroupView { v.onChange = fn; return v }
func (v *RadioGroupView) Value() string                                  { return v.value }

// SetValue chooses an option ("" clears) without calling OnChange.
func (v *RadioGroupView) SetValue(s string)   { v.value = s }
func (v *RadioGroupView) SetDisabled(on bool) { v.disabled = on }
func (v *RadioGroupView) SetOptions(options ...string) {
	v.options = options
	if v.index(v.value) < 0 {
		v.value = ""
	}
}

func (v *RadioGroupView) index(s string) int {
	for i, o := range v.options {
		if o == s {
			return i
		}
	}
	return -1
}

func (v *RadioGroupView) choose(s string) {
	if v.value == s {
		return
	}
	v.value = s
	if v.onChange != nil {
		v.onChange(s)
	}
}

func (v *RadioGroupView) Render(cx *el.Context) el.Element {
	id := autoID("radio", v)
	group := el.Div().Role("radiogroup").Name(v.a11y()).Gap(8).Disabled(v.disabled)
	if v.horizontal {
		group.Row().Gap(16)
	}
	tabStop := max(v.index(v.value), 0)
	for i, o := range v.options {
		on := o == v.value
		// Filled circles, not thick borders: a stroked ring this small looks faceted.
		ring, fill := theme.Muted, theme.Surface
		switch {
		case v.disabled:
			ring, fill = theme.Border, theme.Subtle
		case on:
			ring = theme.Primary
		}
		inner := el.Div().Size(el.Dp(16)).Rounded(8).Bg(fill).Center()
		if on {
			inner.Child(el.Div().Size(el.Dp(8)).Rounded(4).Bg(ring))
		}
		dot := el.Div().Size(el.Dp(18)).NoShrink().Rounded(9).Bg(ring).Center().Child(inner)
		i, o := i, o
		row := check(id+"/"+strconv.Itoa(i), "radio", o, "", on, v.disabled, dot, func() {
			v.choose(o)
			cx.Focus(id + "/" + strconv.Itoa(i)) // it becomes the group's Tab stop
		}).
			Focusable(i == tabStop).
			OnKey(func(e el.KeyEvent) bool {
				d := 0
				switch key.Name(e.Name) {
				case key.NameDownArrow, key.NameRightArrow:
					d = 1
				case key.NameUpArrow, key.NameLeftArrow:
					d = -1
				default:
					return false
				}
				if e.State == el.KeyPress && len(v.options) > 0 {
					j := (i + d + len(v.options)) % len(v.options)
					v.choose(v.options[j])
					cx.Focus(id + "/" + strconv.Itoa(j))
				}
				return true
			})
		group.Child(row)
	}
	return labelled(v.label, group, "")
}

func (v *RadioGroupView) setName(s string) { v.name = s }
func (v *RadioGroupView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
