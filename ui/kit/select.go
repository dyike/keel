package kit

import (
	"image/color"
	"strings"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SelectView picks one option from a list that opens under the field.
// Enter, Space or ↓ opens it; in the list ↑ ↓ move, Enter chooses, Esc closes
// and focus returns to the field. Searchable adds a filter box on top; ↓
// moves from it into the list.
type SelectView struct {
	renderItem                               func(*el.Context, SelectItemContext) el.Element
	renderValue                              func(*el.Context, []SelectOption) el.Element
	empty                                    el.View
	titlePrefix                              string
	clearable, plain                         bool
	menuWidth, menuHeight, height, rowHeight float32
	match                                    func(SelectOption, string) bool
	name                                     string // accessible name from a Form row when label is empty
	label, hint, value, err                  string
	entries                                  []SelectOption
	rows                                     []selectRow
	virtual                                  *VirtualListView
	active                                   int
	multiple, cached                         bool
	values                                   map[string]bool
	onValues                                 func([]string)
	revision, cachedRevision                 uint64
	cachedQuery                              string
	open, searchable                         bool
	disabled                                 bool
	query                                    string
	onChange                                 func(string)
	typeahead                                base.Typeahead
}

func Select(label string, options ...string) *SelectView {
	v := &SelectView{label: label, active: -1}
	v.virtual = VirtualList(0, 30, v.optionRow)
	v.SetOptions(options...)
	return v
}

// Hint is shown while nothing is chosen; empty uses the locale's SelectHint.
func (v *SelectView) Hint(s string) *SelectView                  { v.hint = s; return v }
func (v *SelectView) Searchable() *SelectView                    { v.searchable = true; v.cached = false; return v }
func (v *SelectView) OnChange(fn func(value string)) *SelectView { v.onChange = fn; return v }
func (v *SelectView) Value() string                              { return v.value }

// SetValue chooses an option ("" clears) without calling OnChange.
func (v *SelectView) SetValue(s string) {
	v.value = s
	v.values = make(map[string]bool)
	if s != "" {
		v.values[s] = true
	}
}
func (v *SelectView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.open = false
	}
}
func (v *SelectView) SetError(msg string) { v.err = msg }
func (v *SelectView) Error() string       { return v.err }
func (v *SelectView) FocusID() string     { return autoID("select", v) }

// SetOptions replaces the options, keeping the choice if it is still offered.
func (v *SelectView) SetOptions(options ...string) {
	entries := make([]SelectOption, len(options))
	for i, option := range options {
		entries[i] = SelectOption{Value: option, Label: option}
	}
	v.setEntries(entries)
}
func (v *SelectView) offered(value string) bool {
	for _, option := range v.entries {
		if option.Value == value {
			return true
		}
	}
	return false
}
func (v *SelectView) setOpen(cx *el.Context, open bool) {
	if v.disabled {
		return
	}
	v.virtual.rowH = v.optionHeight()
	v.open, v.query = open, ""
	if !open {
		return
	}
	v.buildRows()
	v.typeahead.Reset()
	v.active = v.firstEnabled()
	for i, row := range v.rows {
		if row.index >= 0 && v.entries[row.index].Value == v.value && !v.rowDisabled(i) {
			v.active = i
			break
		}
	}
	if v.active >= 0 {
		v.virtual.ScrollTo(cx, v.active)
	}
	if v.searchable {
		cx.Focus(v.FocusID() + "/search")
	} else {
		cx.Focus(v.FocusID() + "/options")
	}
}
func (v *SelectView) choose(value string) {
	if v.disabled {
		return
	}
	for _, option := range v.entries {
		if option.Value == value && option.Disabled {
			return
		}
	}
	old := v.value
	if v.multiple {
		if v.values == nil {
			v.values = make(map[string]bool)
		}
		if v.values[value] {
			delete(v.values, value)
		} else {
			v.values[value] = true
		}
		values := v.Values()
		v.value = ""
		if len(values) > 0 {
			v.value = values[len(values)-1]
		}
		if v.onValues != nil {
			v.onValues(values)
		}
	} else {
		v.value = value
		v.open = false
	}
	v.err = ""
	if old != v.value && v.onChange != nil {
		v.onChange(v.value)
	}
}

func (v *SelectView) Render(cx *el.Context) el.Element {
	id := v.FocusID()
	name := v.a11y()
	if name == "" { // name an unlabelled select after the hint it shows
		name = v.hint
		if name == "" {
			name = locale.Current().SelectHint
		}
	}
	shown, textColor := v.shownValue(), theme.Text
	if shown == "" {
		shown, textColor = v.hint, theme.Muted
		if shown == "" {
			shown = locale.Current().SelectHint
		}
	}
	ratio := v.sizeRatio()
	field := fieldFrame(id+"/frame", cx.FocusWithin(id), v.err != "", v.disabled, false).
		MinH(el.Dp(float32(theme.ControlHeight) * ratio)).P(0).Gap(theme.SpaceMd * ratio)
	if v.height > 0 {
		field.TextSize(float32(theme.BodySize) * ratio)
	}
	if v.plain {
		field.Bg(color.NRGBA{}).Border(0, color.NRGBA{})
	}
	// The focusable trigger owns the whole field, including its padding. Keep
	// the framed variant's border in the outer box and use a fill for plain
	// controls; a text-height focus ring would overlap the selected label.
	triggerHeight := float32(theme.ControlHeight) * ratio
	if !v.plain {
		triggerHeight -= 2 // the frame's 1dp border on each side
	}
	trigger := el.Div().ID(id).Role("select").Name(name).Value(strings.Join(v.Values(), ", ")).
		Grow().W(el.Dp(0)).MinH(el.Dp(triggerHeight)).Px(10 * ratio).Py(theme.SpaceXs * ratio).
		Rounded(theme.RadiusMd).Row().Items(el.Center).Gap(theme.SpaceSm).Focusable(true).FocusStyle(func(s *el.Style) {
		s.BorderColor(color.NRGBA{})
		if v.plain {
			s.Bg(theme.Highlight)
		}
	}).
		OnClick(func() { v.setOpen(cx, !v.open) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) != key.NameDownArrow {
				return false
			}
			if e.State == el.KeyPress && !v.open {
				v.setOpen(cx, true)
			}
			return true
		})
	var display el.Element
	if v.value != "" && v.renderValue != nil {
		var selection []SelectOption
		for _, entry := range v.entries {
			if v.picked(entry.Value) {
				selection = append(selection, entry)
			}
		}
		if len(selection) > 0 {
			display = v.renderValue(cx, selection)
		}
	}
	if display == nil {
		display = el.Text(shown).TextColor(textColor).MaxLines(1)
	}
	if v.value != "" && v.titlePrefix != "" {
		trigger.Child(el.Text(v.titlePrefix).MaxW(el.Frac(.5)).MaxLines(1))
	}
	trigger.Child(el.Div().Grow().W(el.Dp(0)).Child(display), Icon(IconChevronDown).Size(16*ratio).Color(theme.Muted).Render(cx))
	field.Child(trigger)
	if v.clearable && len(v.Values()) > 0 {
		clear := Button("", func() { cx.Focus(id); v.clearSelection() }).ID(id + "/clear").Name(locale.Current().Name(locale.Current().Clear, name)).Icon(IconClose).Variant(ButtonGhost).Size(24 * ratio)
		field.Child(clear.Render(cx))
	}

	if v.disabled {
		field.TextColor(theme.Muted)
	} else {
		field.CursorPointer()
	}
	if v.open {
		layer := el.Anchored(id+"/frame", v.list(cx, id)).Modal().TrapFocus().OnDismiss(func() { v.open, v.query = false, "" })
		if v.menuWidth == 0 {
			layer.MatchAnchorWidth()
		}
		cx.Overlay(id, layer)
	}
	return labelled(v.label, field, v.err)
}

func checkMark(cx *el.Context, on bool) el.Element {
	if !on {
		return el.Div().Size(el.Dp(14))
	}
	return Icon(IconDone).Size(14).Color(theme.PrimaryText).Render(cx)
}

func (v *SelectView) setName(s string) { v.name = s }
func (v *SelectView) a11y() string {
	if v.label != "" {
		return v.label
	}
	return v.name
}

func (v *SelectView) sizeRatio() float32 {
	if v.height > 0 {
		return v.height / float32(theme.ControlHeight)
	}
	return 1
}
func (v *SelectView) optionHeight() float32 {
	if v.rowHeight > 0 {
		return v.rowHeight
	}
	return 30 * v.sizeRatio()
}
