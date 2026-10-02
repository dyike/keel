package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// RadioGroupView picks one of several options. Tab enters the group at the
// chosen option; arrow keys move the choice, like native radio groups.
type RadioGroupView struct {
	name           string // accessible name from a Form row when label is empty
	label          string
	options        []string
	value          string
	horizontal     bool
	disabled       bool
	optionDisabled map[string]bool
	onChange       func(string)
}

func RadioGroup(label string, options ...string) *RadioGroupView {
	v := &RadioGroupView{label: label}
	v.SetOptions(options...)
	return v
}
func (v *RadioGroupView) Horizontal() *RadioGroupView                    { v.horizontal = true; return v }
func (v *RadioGroupView) OnChange(fn func(value string)) *RadioGroupView { v.onChange = fn; return v }
func (v *RadioGroupView) Value() string                                  { return v.value }

// SetValue chooses an option ("" clears) without calling OnChange.
func (v *RadioGroupView) SetValue(s string) {
	if s != "" && v.index(s) < 0 {
		s = ""
	}
	v.value = s
}
func (v *RadioGroupView) SetDisabled(on bool) { v.disabled = on }
func (v *RadioGroupView) SetOptions(options ...string) {
	v.options = nil
	seen := map[string]bool{}
	for _, o := range options {
		if o != "" && !seen[o] {
			v.options = append(v.options, o)
			seen[o] = true
		}
	}
	for o := range v.optionDisabled {
		if !seen[o] {
			delete(v.optionDisabled, o)
		}
	}
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
	if v.value == s || v.disabled || v.optionDisabled[s] || v.index(s) < 0 {
		return
	}
	v.value = s
	if v.onChange != nil {
		v.onChange(s)
	}
}

// Options returns a copy of the options, in keyboard navigation order.
func (v *RadioGroupView) Options() []string { return append([]string(nil), v.options...) }

// SetOptionDisabled prevents user selection while preserving a selected value.
func (v *RadioGroupView) SetOptionDisabled(value string, on bool) {
	if v.index(value) < 0 {
		return
	}
	if v.optionDisabled == nil {
		v.optionDisabled = map[string]bool{}
	}
	v.optionDisabled[value] = on
}
func (v *RadioGroupView) itemID(value string) string { return autoID("radio", v) + "/item/" + value }
func (v *RadioGroupView) tabStop() int {
	i := v.index(v.value)
	if i >= 0 && !v.optionDisabled[v.value] {
		return i
	}
	for i, o := range v.options {
		if !v.optionDisabled[o] {
			return i
		}
	}
	return -1
}
func (v *RadioGroupView) FocusID() string {
	if i := v.tabStop(); i >= 0 {
		return v.itemID(v.options[i])
	}
	return ""
}

// Item renders an option independently, sharing this group's selection and
// keyboard order. Render each option at most once per root/frame.
func (v *RadioGroupView) Item(value string) el.View {
	return el.ViewFunc(func(cx *el.Context) el.Element {
		if i := v.index(value); i >= 0 {
			return v.renderItem(cx, i)
		}
		return el.Div()
	})
}
func (v *RadioGroupView) renderItem(cx *el.Context, i int) el.Element {
	o := v.options[i]
	on := o == v.value
	disabled := v.disabled || v.optionDisabled[o]
	ring, fill := theme.Muted, theme.Surface
	if disabled {
		ring, fill = theme.Border, theme.Subtle
	} else if on {
		ring = theme.Primary
	}
	inner := el.Div().Size(el.Dp(16)).Rounded(theme.RadiusLg).Bg(fill).Center()
	if on {
		inner.Child(el.Div().Size(el.Dp(8)).Rounded(theme.RadiusSm).Bg(ring))
	}
	dot := el.Div().Size(el.Dp(18)).NoShrink().Rounded(theme.RadiusFull).Bg(ring).Center().Child(inner)
	return check(v.itemID(o), "radio", o, "", on, disabled, dot, func() { v.choose(o); cx.Focus(v.itemID(o)) }).Focusable(i == v.tabStop()).
		OnKey(func(e el.KeyEvent) bool {
			if e.Modifiers != 0 {
				return false
			}
			direction, start := 0, i
			switch key.Name(e.Name) {
			case key.NameDownArrow, key.NameRightArrow:
				direction = 1
			case key.NameUpArrow, key.NameLeftArrow:
				direction = -1
			case key.NameHome:
				direction, start = 1, -1
			case key.NameEnd:
				direction, start = -1, 0
			default:
				return false
			}
			if e.State == el.KeyPress {
				for n := 1; n <= len(v.options); n++ {
					j := (start + direction*n + len(v.options)) % len(v.options)
					option := v.options[j]
					if !v.optionDisabled[option] && cx.Enabled(v.itemID(option)) {
						v.choose(option)
						cx.Focus(v.itemID(option))
						break
					}
				}
			}
			return true
		})
}
func (v *RadioGroupView) Render(cx *el.Context) el.Element {
	group := el.Div().Role("radiogroup").Name(v.a11y()).Gap(theme.SpaceMd).Disabled(v.disabled)
	if v.horizontal {
		group.Row().Wrap().Gap(theme.SpaceXl)
	}
	for i := range v.options {
		group.Child(v.renderItem(cx, i))
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
