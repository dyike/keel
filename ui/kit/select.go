package kit

import (
	"strings"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SelectView picks one option from a list that opens under the field.
// Enter, Space or ↓ opens it; in the list ↑ ↓ move, Enter chooses, Esc closes
// and focus returns to the field. Searchable adds a filter box on top; ↓
// moves from it into the list.
type SelectView struct {
	name                     string // accessible name from a Form row when label is empty
	label, hint, value, err  string
	entries                  []SelectOption
	rows                     []selectRow
	virtual                  *VirtualListView
	active                   int
	multiple, cached         bool
	values                   map[string]bool
	onValues                 func([]string)
	revision, cachedRevision uint64
	cachedQuery              string
	open, searchable         bool
	disabled                 bool
	query                    string
	onChange                 func(string)
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
	v.open, v.query = open, ""
	if !open {
		return
	}
	v.buildRows()
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
	shown, color := v.shownValue(), theme.Text
	if shown == "" {
		shown, color = v.hint, theme.Muted
		if shown == "" {
			shown = locale.Current().SelectHint
		}
	}
	field := fieldFrame(id, false, v.err != "", v.disabled, false).Role("select").Name(name).Value(strings.Join(v.Values(), ", ")).
		Focusable(true).OnClick(func() { v.setOpen(cx, !v.open) }).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) != key.NameDownArrow {
				return false
			}
			if e.State == el.KeyPress && !v.open {
				v.setOpen(cx, true)
			}
			return true
		}).
		Child(el.Text(shown).TextColor(color).Grow().MaxLines(1), Icon(IconChevronDown).Size(16).Color(theme.Muted).Render(cx))
	if v.disabled {
		field.TextColor(theme.Muted)
	} else {
		field.CursorPointer()
	}
	if v.open {
		cx.Overlay(id, el.Anchored(id, v.list(cx, id)).MatchAnchorWidth().Modal().TrapFocus().
			OnDismiss(func() { v.open, v.query = false, "" }))
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
