package kit

import (
	"slices"
	"strings"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

func (v *ComboboxView) Multiple() *ComboboxView {
	v.multiple = true
	v.SetValues([]string{v.value})
	return v
}
func (v *ComboboxView) Values() []string {
	if !v.multiple {
		if v.value != "" {
			return []string{v.value}
		}
		return nil
	}
	return slices.Clone(v.values)
}
func (v *ComboboxView) SetValues(values []string) {
	v.close()
	v.searchError = ""
	v.cached = false
	v.values = nil
	v.value = ""
	v.text = ""
	for _, value := range values {
		if value != "" && !slices.Contains(v.values, value) {
			v.values = append(v.values, value)
			v.value = value
			if !v.multiple {
				v.text = v.optionLabel(value)
				break
			}
		}
	}
}
func (v *ComboboxView) OnValuesChange(fn func([]string)) *ComboboxView { v.onValues = fn; return v }
func (v *ComboboxView) removeValue(value string) {
	before := v.Values()
	v.values = slices.DeleteFunc(v.values, func(s string) bool { return s == value })
	old := v.value
	v.value = ""
	if len(v.values) > 0 {
		v.value = v.values[len(v.values)-1]
	}
	if !slices.Equal(before, v.values) && v.onValues != nil {
		v.onValues(v.Values())
	}
	if old != v.value && v.onChange != nil {
		v.onChange(v.value)
	}
}
func (v *ComboboxView) close() { v.open = false; v.request++; v.loading = false }
func (v *ComboboxView) cancelDraft() {
	v.close()
	v.focused = false
	v.text = v.optionLabel(v.value)
	if v.multiple {
		v.text = ""
	}
}

// OnSearch delegates suggestion lookup to the application. Deliver asynchronous
// results on the UI loop via SetResults/SetSearchError; stale tokens are ignored.
func (v *ComboboxView) OnSearch(fn func(query string, token uint64)) *ComboboxView {
	v.onSearch = fn
	v.cached = false
	if v.open {
		v.searchChanged()
	}
	return v
}
func (v *ComboboxView) searchChanged() {
	v.request++
	v.active = v.enabledOption(v.matches(), 0, 1)
	v.searchError = ""
	v.loading = v.onSearch != nil
	if v.onSearch != nil {
		query := v.text
		if v.nonsearchable {
			query = ""
		}
		v.onSearch(query, v.request)
	}
}
func (v *ComboboxView) SetResults(token uint64, options ...string) bool {
	if !v.open || v.disabled || token != v.request || v.onSearch == nil {
		return false
	}
	v.SetOptions(options...)
	v.loading = false
	v.searchError = ""
	v.active = v.enabledOption(v.matches(), 0, 1)
	return true
}
func (v *ComboboxView) SetSearchError(token uint64, message string) bool {
	if !v.open || token != v.request || v.onSearch == nil {
		return false
	}
	v.loading = false
	v.searchError = message
	v.active = -1
	return true
}
func (v *ComboboxView) optionKey(cx *el.Context, e el.KeyEvent) bool {
	switch key.Name(e.Name) {
	case key.NameUpArrow, key.NameDownArrow, key.NamePageUp, key.NamePageDown:
	default:
		return false
	}
	if e.State != el.KeyPress || e.Modifiers != 0 {
		return true
	}
	if !v.open {
		if key.Name(e.Name) == key.NameDownArrow {
			v.open = true
			v.searchChanged()
		}
		return true
	}
	if v.loading || v.searchError != "" {
		return true
	}
	matches := v.matches()
	if len(matches) == 0 {
		return true
	}
	switch key.Name(e.Name) {
	case key.NameDownArrow:
		v.active = (v.active + 1) % len(matches)
	case key.NameUpArrow:
		if v.active < 0 {
			v.active = len(matches) - 1
		} else {
			v.active = (v.active - 1 + len(matches)) % len(matches)
		}
	default:
		v.active, _ = base.List{Count: len(matches), Page: 8}.Key(e.Name, v.active)
	}
	direction := 1
	if key.Name(e.Name) == key.NameUpArrow || key.Name(e.Name) == key.NamePageUp {
		direction = -1
	}
	v.active = v.enabledOption(matches, v.active, direction)
	v.virtual.SetCount(len(v.displayRows))
	if v.active >= 0 {
		v.virtual.ScrollTo(cx, v.displayIndex(v.active))
	}
	return true
}
func (v *ComboboxView) optionRow(cx *el.Context, i int) el.Element {
	option := v.matches()[i]
	ratio := v.sizeRatio()
	selected := option == v.value
	if v.multiple {
		selected = slices.Contains(v.values, option)
	}
	row := el.Div().Disabled(v.optionDisabled(option)).DisabledStyle(func(s *el.Style) { s.TextColor(theme.Muted) }).Role("option").Name(v.optionLabel(option)).Value(option).Selected(selected).H(el.Dp(max(1, v.optionHeight()-2*ratio))).My(ratio).Mx(4 * ratio).Px(theme.SpaceMd * ratio).Row().Items(el.Center).Rounded(theme.RadiusSm).CursorPointer().Focusable(false).
		Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	row.Child(el.Div().ID("activate").Absolute().Top(0).Left(0).W(el.Full).H(el.Full).OnClick(func() { v.choose(option); cx.Focus(v.FocusID()) }), v.renderItem(cx, option, selected), v.renderCheck(cx, selected))
	if v.height > 0 {
		row.TextSize(float32(theme.BodySize) * ratio)
	}
	switch {
	case selected:
		row.Bg(theme.Highlight)
	case i == v.active:
		row.Bg(theme.Subtle)
	}
	return row
}
func (v *ComboboxView) suggestions(cx *el.Context, id string) el.Element {
	list := floating(theme.ElevationMd).ID(id + "/list").Role("listbox").Name(v.a11y()).Py(theme.SpaceXs).Items(el.Stretch)
	_, height := cx.ViewportSize()
	footer, reserve := v.renderFooter(cx, id, height)
	available := max(1, height-80-reserve)
	body := el.Div().ID(id + "/results").Items(el.Stretch).MaxH(el.Dp(available)).ScrollY()
	switch {
	case v.loading:
		body.Child(el.Div().P(theme.SpaceLg).Row().Gap(theme.SpaceMd).Child(Spinner().Render(cx), el.Text(locale.Current().Loading)))
	case v.searchError != "":
		body.Child(el.Div().P(theme.SpaceLg).Gap(theme.SpaceMd).Child(el.Text(v.searchError).TextColor(theme.Danger), Button(locale.Current().Retry, v.searchChanged).Render(cx)))
	default:
		matches := v.matches()
		if len(matches) == 0 {
			body.Child(el.Div().Px(theme.SpaceLg).Py(theme.SpaceSm).Child(el.Text(locale.Current().NoMatches).TextColor(theme.Muted)))
		} else {
			v.virtual.SetCount(len(v.displayRows))
			v.virtual.rowH = v.optionHeight()
			v.virtual.Height(min(min(240, available), float32(len(v.displayRows))*v.virtual.rowH))
			body.Child(v.virtual.Render(cx))
		}
	}
	list.Child(body)
	if footer != nil {
		list.Child(footer)
	}
	return list
}

type comboboxBlurKey struct{ id string }

func (v *ComboboxView) filteredMatches() []string {
	if v.onSearch != nil || v.nonsearchable {
		return v.options
	}
	q := strings.ToLower(strings.TrimSpace(v.text))
	if q == "" || !v.multiple && q == strings.ToLower(v.optionLabel(v.value)) {
		return v.options
	}
	var matches []string
	for _, option := range v.options {
		if strings.Contains(strings.ToLower(option), q) || strings.Contains(strings.ToLower(v.optionLabel(option)), q) {
			matches = append(matches, option)
		}
	}
	return matches
}
