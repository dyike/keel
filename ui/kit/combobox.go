package kit

import (
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ComboboxView is a text field with a filtered list of suggestions. Typing
// opens and filters the list; clicking an option takes it. Enter takes the
// typed text if it is an option (or AllowCustom is set), else the first match. Esc or a click elsewhere closes the list.
// Without AllowCustom, leaving the field with text that is not an option
// restores the last choice. The arrow keys move the caret, as in any text
// field; pick with the pointer or Enter.
type ComboboxView struct {
	name                                 string // accessible name from a Form row when label is empty
	label, placeholder, text, value, err string
	options                              []string
	open, allowCustom, disabled, focused bool
	onChange                             func(string)
}

func Combobox(label string, options ...string) *ComboboxView {
	return &ComboboxView{label: label, options: options}
}
func (v *ComboboxView) Placeholder(s string) *ComboboxView           { v.placeholder = s; return v }
func (v *ComboboxView) AllowCustom() *ComboboxView                   { v.allowCustom = true; return v }
func (v *ComboboxView) OnChange(fn func(value string)) *ComboboxView { v.onChange = fn; return v }
func (v *ComboboxView) Value() string                                { return v.value }
func (v *ComboboxView) SetValue(s string)                            { v.value, v.text = s, s }
func (v *ComboboxView) SetOptions(options ...string)                 { v.options = options }
func (v *ComboboxView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.open = false
	}
}
func (v *ComboboxView) SetError(msg string) { v.err = msg }
func (v *ComboboxView) Error() string       { return v.err }
func (v *ComboboxView) FocusID() string     { return autoID("combobox", v) + "/text" }

func (v *ComboboxView) matches() []string {
	q := strings.ToLower(strings.TrimSpace(v.text))
	if q == "" || q == strings.ToLower(v.value) {
		return v.options
	}
	var out []string
	for _, o := range v.options {
		if strings.Contains(strings.ToLower(o), q) {
			out = append(out, o)
		}
	}
	return out
}

func (v *ComboboxView) choose(s string) {
	v.open, v.text = false, s
	if s == v.value {
		return
	}
	v.value, v.err = s, ""
	if v.onChange != nil {
		v.onChange(s)
	}
}

func (v *ComboboxView) offered(s string) bool {
	for _, o := range v.options {
		if o == s {
			return true
		}
	}
	return false
}

// settle handles leaving the field or pressing Enter with no match.
func (v *ComboboxView) settle() {
	switch {
	case v.offered(v.text) || v.allowCustom && strings.TrimSpace(v.text) != "":
		v.choose(strings.TrimSpace(v.text))
	default:
		v.text = v.value
	}
	v.open = false
}

func (v *ComboboxView) Render(cx *el.Context) el.Element {
	id := autoID("combobox", v)
	focused := cx.FocusWithin(id) || v.open && cx.FocusWithin(id+"/list")
	if v.focused && !focused && !v.open {
		v.settle()
	}
	v.focused = focused
	border := theme.Border
	switch {
	case v.err != "":
		border = theme.Danger
	case focused && !v.disabled:
		border = theme.Primary
	}
	field := el.Input().ID(v.FocusID()).Name(v.a11y()).Placeholder(v.placeholder).Bind(&v.text).
		Border(0, theme.Border).Bg(theme.Surface).P(0).Grow().
		OnChange(func(string) { v.open = true }).
		OnSubmit(func(string) {
			t := strings.TrimSpace(v.text)
			if m := v.matches(); !v.offered(t) && !v.allowCustom && len(m) > 0 {
				v.choose(m[0])
				return
			}
			v.settle()
		})
	toggle := el.Div().Name(locale.Current().Name(locale.Current().MoreOptions, v.a11y())).P(2).Rounded(4).
		Focusable(false).CursorPointer().OnClick(func() {
		v.open = !v.open
		cx.Focus(v.FocusID())
	}).Child(Icon(IconChevronDown).Size(16).Color(theme.Muted).Render(cx))
	box := el.Div().ID(id).WFull().Role("combobox").Name(v.a11y()).Value(v.value).Row().Items(el.Center).Gap(8).
		Px(10).Py(8).Rounded(6).Border(1, border).Bg(theme.Surface).Disabled(v.disabled).Child(field, toggle)
	if v.disabled {
		box.Bg(theme.Subtle)
		field.Bg(theme.Subtle)
	}
	if v.open && !v.disabled {
		list := surface().ID(id + "/list").Role("listbox").Name(v.a11y()).Py(4).Items(el.Stretch).MaxH(el.Dp(240)).ScrollY()
		m := v.matches()
		if len(m) == 0 {
			list.Child(el.Div().Px(12).Py(6).Child(el.Text(locale.Current().NoMatches).TextColor(theme.Muted)))
		}
		for i, o := range m {
			o := o
			list.Child(el.Div().ID(id+"/"+strconv.Itoa(i)).Role("option").Name(o).Selected(o == v.value).
				Mx(4).Px(8).H(el.Dp(30)).Row().Items(el.Center).Rounded(4).CursorPointer().Focusable(false).
				Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).OnClick(func() { v.choose(o) }).
				Child(el.Text(o).Grow().MaxLines(1), checkMark(cx, o == v.value)))
		}
		cx.Overlay(id, el.Anchored(id, list).MatchAnchorWidth().OnDismiss(func() { v.settle() }))
	}
	return labelled(v.label, box, v.err)
}

func (v *ComboboxView) setName(s string) { v.name = s }
func (v *ComboboxView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}
