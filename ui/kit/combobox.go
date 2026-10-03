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
	disabledOptions                      map[string]bool
	footer                               el.View
	footerHeight                         float32
	footerMeasured                       bool
	clearable                            bool
	height                               float32
	checkIcon                            *IconView
	itemLabels                           map[string]string
	itemDisabled                         map[string]bool
	itemRenderer                         func(ComboboxItem, bool) el.View
	rowHeight                            float32
	matchKeys                            []string
	groupFor, groupLabels                map[string]string
	displayRows                          []comboboxDisplayRow
	optionRows                           []int
	open, allowCustom, disabled, focused bool
	nonsearchable                        bool
	active                               int // highlighted match while open, -1 none
	onChange                             func(string)
	onConfirm                            func([]string)
}

func Combobox(label string, options ...string) *ComboboxView {
	v := &ComboboxView{label: label, options: slices.Clone(options), active: -1}
	v.virtual = VirtualList(0, 30, v.displayRow).ItemKey(func(i int) string { v.matches(); return v.matchKeys[i] })
	return v
}
func (v *ComboboxView) Placeholder(s string) *ComboboxView           { v.placeholder = s; return v }
func (v *ComboboxView) AllowCustom() *ComboboxView                   { v.allowCustom = true; return v }
func (v *ComboboxView) OnChange(fn func(value string)) *ComboboxView { v.onChange = fn; return v }
func (v *ComboboxView) Value() string                                { return v.value }
func (v *ComboboxView) SetValue(s string)                            { v.SetValues([]string{s}) }
func (v *ComboboxView) SetOptions(options ...string) {
	oldDisplay := v.optionLabel(v.value)
	v.itemLabels, v.itemDisabled = nil, nil
	v.groupFor, v.groupLabels = nil, nil
	if !v.multiple && v.text == oldDisplay {
		v.text = v.value
	}
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
		v.buildDisplayRows()
		v.cached = true
		v.cachedRevision = v.revision
		v.cachedText = v.text
		v.cachedValue = v.value
	}
	return v.filtered
}
func (v *ComboboxView) choose(value string) { v.chooseValue(value, false) }
func (v *ComboboxView) chooseValue(value string, finish bool) {
	if v.disabled || v.optionDisabled(value) || v.loading || v.searchError != "" || v.multiple && value == "" {
		return
	}
	old, wasOpen := v.value, v.open
	if v.multiple {
		if slices.Contains(v.values, value) {
			v.values = slices.DeleteFunc(v.values, func(s string) bool { return s == value })
			v.value = ""
			if len(v.values) > 0 {
				v.value = v.values[len(v.values)-1]
			}
		} else {
			v.values = append(v.values, value)
			v.value = value
		}
		v.text = ""
	} else {
		v.text = v.optionLabel(value)
		v.value = value
		v.active = -1
	}
	closes := finish || !v.multiple
	if closes {
		v.close()
	} else {
		v.open = true
	}
	v.err = ""
	values, selected, request := v.Values(), v.value, v.request
	if v.multiple && v.onValues != nil {
		v.onValues(slices.Clone(values))
	}
	if old != selected && v.onChange != nil {
		v.onChange(selected)
	}
	if closes {
		v.emitConfirm(wasOpen, values)
	} else if v.open && !v.disabled && v.request == request {
		v.searchChanged()
	}
}

func (v *ComboboxView) offered(s string) bool { _, ok := v.inputValue(s); return ok }

// settle handles leaving the field or pressing Enter with no match.
func (v *ComboboxView) settle() {
	if v.nonsearchable {
		v.cancelAndConfirm()
		return
	}
	if v.loading || v.searchError != "" {
		v.cancelAndConfirm()
		return
	}
	if !v.multiple && v.value != "" && v.text == v.optionLabel(v.value) {
		v.confirmClose()
		return
	}
	value, offered := v.inputValue(strings.TrimSpace(v.text))
	switch {
	case !v.optionDisabled(value) && (offered || v.allowCustom && value != ""):
		v.chooseValue(value, true)
		return
	default:
		v.text = v.optionLabel(v.value)
		if v.multiple {
			v.text = ""
		}
	}
	v.confirmClose()
}

func (v *ComboboxView) Render(cx *el.Context) el.Element {
	id := autoID("combobox", v)
	ratio := v.sizeRatio()
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
				if i := v.enabledOption(m, 0, 1); i >= 0 {
					v.choose(m[i])
					return
				}
			}
			v.settle()
		})
	toggle := el.Div().Name(locale.Current().Name(locale.Current().MoreOptions, v.a11y())).P(theme.SpaceXxs * ratio).Rounded(theme.RadiusSm).
		Focusable(false).CursorPointer().OnClick(func() {
		if v.open {
			v.confirmClose()
		} else {
			v.open = true
			v.searchChanged()
		}
		cx.Focus(v.FocusID())
	}).Child(Icon(IconChevronDown).Size(16 * ratio).Color(theme.Muted).Render(cx))
	box := fieldFrame(id, focused, v.err != "", v.disabled, false).FocusOnPress(v.FocusID()).Role("combobox").Name(v.a11y()).Value(strings.Join(v.Values(), ", "))
	box.MinH(el.Dp(float32(theme.ControlHeight) * ratio)).Px(10 * ratio).Py(theme.SpaceXs * ratio).Gap(theme.SpaceMd * ratio)
	if v.height > 0 {
		box.TextSize(float32(theme.BodySize) * ratio)
	}
	if v.multiple {
		box.Wrap()
		for _, value := range v.values {
			tag := Tag(v.optionLabel(value)).OnRemove(func() { v.removeValue(value); cx.Focus(v.FocusID()) })
			if v.height > 0 {
				tag.Size(max(16, 24*ratio))
			}
			box.Child(el.Div().ID("tag/" + value).Child(tag.Render(cx)))
		}
		field.MinW(el.Dp(100))
	}
	var fieldView el.Element = field
	if v.nonsearchable {
		fieldView = v.staticTrigger(cx)
	}
	box.Child(fieldView)
	if v.clearable && len(v.Values()) > 0 {
		clear := Button("", func() { cx.Focus(v.FocusID()); v.clearSelection() }).ID(id + "/clear").Name(locale.Current().Name(locale.Current().Clear, v.a11y())).Icon(IconClose).Variant(ButtonGhost).Size(24 * ratio)
		clear.SetDisabled(v.disabled)
		box.Child(clear.Render(cx))
	}
	box.Child(toggle)
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
