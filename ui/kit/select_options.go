package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"slices"
	"strings"
)

// SelectOption separates the stored value from its label and section title.
type SelectOption struct {
	Value, Label, Group string
	Disabled            bool
}
type selectRow struct {
	index int
	group string
}

func (v *SelectView) SetEntries(options ...SelectOption) {
	seen := map[string]bool{}
	for _, option := range options {
		if option.Value == "" || seen[option.Value] {
			panic("kit.Select: empty or duplicate option value")
		}
		seen[option.Value] = true
	}
	v.setEntries(options)
}
func (v *SelectView) setEntries(options []SelectOption) {
	v.entries = slices.Clone(options)
	v.revision++
	for i := range v.entries {
		if v.entries[i].Label == "" {
			v.entries[i].Label = v.entries[i].Value
		}
	}
	if !v.offered(v.value) {
		v.value = ""
	}
	for value := range v.values {
		if !v.offered(value) {
			delete(v.values, value)
		}
	}
	if v.multiple && v.value == "" {
		values := v.Values()
		if len(values) > 0 {
			v.value = values[len(values)-1]
		}
	}
	v.active = -1
}
func (v *SelectView) Entries() []SelectOption { return slices.Clone(v.entries) }
func (v *SelectView) SetOptionDisabled(value string, on bool) {
	for i := range v.entries {
		if v.entries[i].Value == value {
			v.entries[i].Disabled = on
			v.revision++
		}
	}
}
func (v *SelectView) Multiple() *SelectView {
	v.multiple = true
	v.SetValues([]string{v.value})
	return v
}
func (v *SelectView) OnValuesChange(fn func([]string)) *SelectView { v.onValues = fn; return v }
func (v *SelectView) Values() []string {
	if !v.multiple {
		if v.value != "" {
			return []string{v.value}
		}
		return nil
	}
	var values []string
	seen := map[string]bool{}
	for _, option := range v.entries {
		if v.values[option.Value] && !seen[option.Value] {
			values = append(values, option.Value)
			seen[option.Value] = true
		}
	}
	return values
}
func (v *SelectView) SetValues(values []string) {
	v.values = make(map[string]bool)
	v.value = ""
	for _, value := range values {
		if value != "" && v.offered(value) {
			v.values[value] = true
			v.value = value
			if !v.multiple {
				break
			}
		}
	}
}
func (v *SelectView) picked(value string) bool {
	if v.multiple {
		return v.values[value]
	}
	return value == v.value
}
func (v *SelectView) shownValue() string {
	var labels []string
	for _, value := range v.Values() {
		label := value
		for _, option := range v.entries {
			if option.Value == value {
				label = option.Label
				break
			}
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, ", ")
}
func (v *SelectView) buildRows() {
	if v.cached && v.cachedRevision == v.revision && v.cachedQuery == v.query {
		return
	}
	v.cached, v.cachedRevision, v.cachedQuery = true, v.revision, v.query
	v.rows = v.rows[:0]
	group := ""
	query := strings.ToLower(v.query)
	for i, option := range v.entries {
		if v.searchable && query != "" && !strings.Contains(strings.ToLower(option.Label+" "+option.Value), query) {
			continue
		}
		if option.Group != "" && option.Group != group {
			v.rows = append(v.rows, selectRow{index: -1, group: option.Group})
		}
		group = option.Group
		v.rows = append(v.rows, selectRow{index: i})
	}
	v.virtual.SetCount(len(v.rows))
}
func (v *SelectView) rowDisabled(i int) bool {
	return v.rows[i].index < 0 || v.entries[v.rows[i].index].Disabled
}
func (v *SelectView) firstEnabled() int {
	i, _ := listEnabledKey(string(key.NameHome), -1, len(v.rows), v.rowDisabled)
	return i
}
func (v *SelectView) optionKey(cx *el.Context, e el.KeyEvent) bool {
	if e.Modifiers != 0 {
		return false
	}
	if key.Name(e.Name) == key.NameReturn || key.Name(e.Name) == key.NameSpace {
		if e.State == el.KeyPress && v.active >= 0 && !v.rowDisabled(v.active) {
			v.choose(v.entries[v.rows[v.active].index].Value)
		}
		return true
	}
	i, ok := listEnabledKey(e.Name, v.active, len(v.rows), v.rowDisabled)
	if key.Name(e.Name) == key.NameDownArrow || key.Name(e.Name) == key.NameUpArrow {
		direction := 1
		if key.Name(e.Name) == key.NameUpArrow {
			direction = -1
		}
		for step := 1; step <= len(v.rows); step++ {
			next := (v.active + direction*step) % len(v.rows)
			if next < 0 {
				next += len(v.rows)
			}
			if !v.rowDisabled(next) {
				i = next
				break
			}
		}
	}
	if ok && e.State == el.KeyPress && i >= 0 {
		v.active = i
		v.virtual.ScrollTo(cx, i)
	}
	return ok
}
func (v *SelectView) optionRow(cx *el.Context, i int) el.Element {
	row := v.rows[i]
	if row.index < 0 {
		// Fill the 30dp slot and sit at its bottom, next to the group it names.
		return el.Div().H(el.Dp(30)).Px(12).Pb(4).Justify(el.End).Child(el.Text(row.group).Bold().TextSize(12).TextColor(theme.Muted))
	}
	option := v.entries[row.index]
	selected := v.picked(option.Value)
	item := el.Div().Role("option").Name(option.Label).Value(option.Value).Selected(selected).Disabled(option.Disabled).H(el.Dp(28)).My(1).Mx(4).Px(8).Row().Items(el.Center).Rounded(4).Focusable(false).Child(el.Text(option.Label).Grow().MaxLines(1), checkMark(cx, selected))
	if selected {
		item.Bg(theme.Highlight)
	}
	if !option.Disabled {
		item.CursorPointer().OnClick(func() {
			v.active = i
			v.choose(option.Value)
			if v.multiple {
				cx.Focus(v.FocusID() + "/options")
			}
		}).Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		// The keyboard's place is a background, like hover; a border would
		// read as a second focus ring beside the search field's.
		if i == v.active && !selected {
			item.Bg(theme.Subtle)
		}
	}
	return item
}
func (v *SelectView) list(cx *el.Context, id string) el.Element {
	v.buildRows()
	if v.active < 0 || v.active >= len(v.rows) || v.rowDisabled(v.active) {
		v.active = v.firstEnabled()
		if v.active >= 0 {
			v.virtual.ScrollTo(cx, v.active)
		}
	}
	panel := surface().Role("listbox").Name(v.a11y()).Py(4).Items(el.Stretch)
	_, height := cx.ViewportSize()
	available := max(float32(1), min(float32(240), height-100))
	if v.searchable {
		available = max(1, available-40)
		panel.Child(el.Div().Px(4).Pb(4).Child(searchField(cx, id+"/searchbox", id+"/search", el.Input().ID(id+"/search").Name(locale.Current().Search).Placeholder(locale.Current().Search).Bind(&v.query).
			OnChange(func(string) { v.active = -1; cx.ScrollTo(v.virtual.ID(), 0) }).
			OnSubmit(func(string) {
				if v.active >= 0 && !v.rowDisabled(v.active) {
					v.choose(v.entries[v.rows[v.active].index].Value)
				}
			}).
			OnKey(func(e el.KeyEvent) bool {
				if key.Name(e.Name) != key.NameDownArrow {
					return false
				}
				if e.State == el.KeyPress {
					cx.Focus(id + "/options")
				}
				return true
			}))))
	}
	v.virtual.Height(min(available, max(30, float32(len(v.rows))*30)))
	options := el.Div().ID(id + "/options").Focusable(true).Items(el.Stretch).OnKey(func(e el.KeyEvent) bool { return v.optionKey(cx, e) })
	if len(v.rows) == 0 {
		options.Child(el.Div().Px(12).Py(6).Child(el.Text(locale.Current().NoMatches).TextColor(theme.Muted)))
	} else {
		options.Child(v.virtual.Render(cx))
	}
	return panel.Child(options)
}
