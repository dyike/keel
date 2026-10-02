package kit

import (
	"slices"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ComboboxView is a text field with a filtered list of suggestions. Typing
// opens and filters the list; clicking an option takes it. Enter takes the
// typed text if it is an option (or AllowCustom is set), else the first match. Esc or a click elsewhere closes the list.
// ↓ opens the list and moves the highlight, ↑ moves it back, and Enter takes
// the highlighted option. Without AllowCustom, leaving the field with text
// that is not an option restores the last choice.
type ComboboxView struct {
	multiple, loading, cached            bool
	values, filtered                     []string
	onValues                             func([]string)
	virtual                              *VirtualListView
	revision, cachedRevision             uint64
	cachedText, cachedValue              string
	searchError                          string
	request                              uint64
	onSearch                             func(string, uint64)
	name                                 string // accessible name from a Form row when label is empty
	label, placeholder, text, value, err string
	options                              []string
	open, allowCustom, disabled, focused bool
	active                               int // highlighted match while open, -1 none
	onChange                             func(string)
}

func Combobox(label string, options ...string) *ComboboxView {
	v := &ComboboxView{label: label, options: slices.Clone(options), active: -1}
	v.virtual = VirtualList(0, 30, v.optionRow)
	return v
}
func (v *ComboboxView) Placeholder(s string) *ComboboxView           { v.placeholder = s; return v }
func (v *ComboboxView) AllowCustom() *ComboboxView                   { v.allowCustom = true; return v }
func (v *ComboboxView) OnChange(fn func(value string)) *ComboboxView { v.onChange = fn; return v }
func (v *ComboboxView) Value() string                                { return v.value }
func (v *ComboboxView) SetValue(s string)                            { v.SetValues([]string{s}) }
func (v *ComboboxView) SetOptions(options ...string) {
	v.options = slices.Clone(options)
	v.revision++
	v.active = -1
}
func (v *ComboboxView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.cancelDraft()
	}
}
func (v *ComboboxView) SetError(msg string) { v.err = msg }
func (v *ComboboxView) Error() string       { return v.err }
func (v *ComboboxView) FocusID() string     { return autoID("combobox", v) + "/text" }

func (v *ComboboxView) matches() []string {
	if !v.cached || v.cachedRevision != v.revision || v.cachedText != v.text || v.cachedValue != v.value {
		v.filtered = v.filteredMatches()
		v.cached = true
		v.cachedRevision = v.revision
		v.cachedText = v.text
		v.cachedValue = v.value
	}
	return v.filtered
}
func (v *ComboboxView) choose(value string) {
	if v.disabled || v.loading || v.searchError != "" || v.multiple && value == "" {
		return
	}
	old := v.value
	if v.multiple {
		added := !slices.Contains(v.values, value)
		if added {
			v.values = append(v.values, value)
		}
		v.value = value
		v.text = ""
		v.open = true
		if added && v.onValues != nil {
			v.onValues(v.Values())
		}
		v.searchChanged()
	} else {
		v.close()
		v.text = value
		v.value = value
		v.active = -1
	}
	v.err = ""
	if old != v.value && v.onChange != nil {
		v.onChange(v.value)
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
	if v.loading || v.searchError != "" {
		v.cancelDraft()
		return
	}
	switch {
	case v.offered(v.text) || v.allowCustom && strings.TrimSpace(v.text) != "":
		v.choose(strings.TrimSpace(v.text))
	default:
		v.text = v.value
		if v.multiple {
			v.text = ""
		}
	}
	v.close()
}

func (v *ComboboxView) Render(cx *el.Context) el.Element {
	id := autoID("combobox", v)
	if v.focused && !cx.Enabled(id) {
		v.cancelDraft()
	}
	focused := !v.disabled && (cx.FocusWithin(id) || v.open && cx.FocusWithin(id+"/list"))
	if v.focused && !focused && !v.open {
		cx.AfterEnabled(id, comboboxBlurKey{id}, 0, func() { v.settle(); v.focused = false })
	} else {
		v.focused = focused
	}
	field := fieldText(el.Input().ID(v.FocusID()).Name(v.a11y()).Placeholder(v.placeholder).Bind(&v.text)).
		OnChange(func(string) { v.open = true; v.searchChanged(); cx.ScrollTo(v.virtual.ID(), 0) }).
		OnKey(func(e el.KeyEvent) bool { return v.optionKey(cx, e) }).
		OnSubmit(func(string) {
			if v.loading || v.searchError != "" {
				return
			}
			if m := v.matches(); v.open && v.active >= 0 && v.active < len(m) {
				v.choose(m[v.active])
				return
			}
			t := strings.TrimSpace(v.text)
			if m := v.matches(); !v.offered(t) && !v.allowCustom && len(m) > 0 {
				v.choose(m[0])
				return
			}
			v.settle()
		})
	toggle := el.Div().Name(locale.Current().Name(locale.Current().MoreOptions, v.a11y())).P(2).Rounded(4).
		Focusable(false).CursorPointer().OnClick(func() {
		if v.open {
			v.close()
		} else {
			v.open = true
			v.searchChanged()
		}
		cx.Focus(v.FocusID())
	}).Child(Icon(IconChevronDown).Size(16).Color(theme.Muted).Render(cx))
	box := fieldFrame(id, focused, v.err != "", v.disabled, false).Role("combobox").Name(v.a11y()).Value(strings.Join(v.Values(), ", "))
	if v.multiple {
		box.Wrap()
		for _, value := range v.values {
			tag := Tag(value).OnRemove(func() { v.removeValue(value); cx.Focus(v.FocusID()) })
			box.Child(el.Div().ID("tag/" + value).Child(tag.Render(cx)))
		}
		field.MinW(el.Dp(100))
	}
	box.Child(field, toggle)
	if v.open && !v.disabled {
		cx.Overlay(id, el.Anchored(id, v.suggestions(cx, id)).MatchAnchorWidth().OnDismiss(func() {
			if !cx.Enabled(id) {
				v.cancelDraft()
			} else {
				v.settle()
			}
			v.active = -1
		}))
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

func (v *ComboboxView) commitForm() {
	if !v.disabled {
		v.settle()
	}
}
