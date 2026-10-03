package kit

import (
	"gioui.org/io/key"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// CommandItem is one entry of a command palette. Shortcut uses
// core.ParseShortcut syntax and is only displayed; Group introduces a section.
type CommandItem struct {
	Title, Group, Shortcut string
	Action                 func()
	Disabled               bool
}

// CommandView is a command palette: a search box over a list of commands that
// narrows as you type. Prefix matches rank first, then substrings, then
// characters that merely appear in order ("设置" finds "打开设置", "nwo"
// finds "New window"). ↑ ↓ move, Enter runs, Esc closes. Bind it to a
// shortcut yourself: cx.Shortcut("mod+k", palette.Toggle).
type CommandView struct {
	list                                *VirtualListView
	inline, nonsearchable, pendingFocus bool
	header, footer, empty               el.View
	renderItem                          func(CommandItem, bool) el.View
	rowHeight                           float32
	rows                                []commandRow
	revision, cachedRevision            uint64
	cachedQuery                         string
	cached                              bool
	disabled, loading                   bool
	searchError                         string
	request                             uint64
	onSearch                            func(string, uint64)
	items                               []CommandItem
	open                                bool
	query                               string
	active                              int
}

func Command(items ...CommandItem) *CommandView {
	v := &CommandView{items: slices.Clone(items), active: -1}
	v.list = VirtualList(0, 36, v.row)
	return v
}
func (v *CommandView) SetItems(items ...CommandItem) {
	v.items = slices.Clone(items)
	v.revision++
	v.active = -1
}
func (v *CommandView) Value() bool { return v.open }
func (v *CommandView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.SetValue(false)
	}
}
func (v *CommandView) SetValue(open bool) {
	if open && v.disabled {
		return
	}
	v.open = open
	v.pendingFocus = open && !v.inline
	v.request++ // Invalidate results from an earlier opening.
	if open {
		v.query = ""
		v.active = -1
		v.cached = false
		v.searchChanged()
	}
}

// Toggle opens or closes the palette; bind it to a shortcut or a button.
func (v *CommandView) Toggle() { v.SetValue(!v.open) }

// fuzzy scores how well query matches text: higher is better; ok is false
// when the query's runes do not all appear in order.
func fuzzy(query, text string) (score int, ok bool) {
	q, t := strings.ToLower(strings.TrimSpace(query)), strings.ToLower(text)
	if q == "" {
		return 0, true
	}
	switch {
	case strings.HasPrefix(t, q):
		return 3000 - utf8.RuneCountInString(t), true
	case strings.Contains(t, q):
		return 2000 - strings.Index(t, q), true
	}
	// Subsequence: reward runs of adjacent matches and starts of words.
	qs := []rune(q)
	i, prev, s := 0, -2, 1000
	runes := []rune(t)
	for j, r := range runes {
		if i < len(qs) && r == qs[i] {
			if prev == j-1 {
				s += 5
			}
			if j == 0 || unicode.IsSpace(runes[j-1]) {
				s += 3
			}
			prev, i = j, i+1
		} else if i > 0 && i < len(qs) {
			s--
		}
	}
	return s, i == len(qs)
}

func (v *CommandView) matches() []CommandItem {
	type scored struct {
		item  CommandItem
		score int
	}
	var out []scored
	for _, it := range v.items {
		if s, ok := fuzzy(v.query, it.Title); ok {
			out = append(out, scored{it, s})
		}
	}
	slices.SortStableFunc(out, func(a, b scored) int { return b.score - a.score })
	items := make([]CommandItem, len(out))
	for i, s := range out {
		items[i] = s.item
	}
	return items
}

func (v *CommandView) run(it CommandItem) {
	if it.Disabled || v.disabled || v.loading || v.searchError != "" {
		return
	}
	if !v.inline {
		v.SetValue(false)
	}
	if it.Action != nil {
		it.Action()
	}
}

func (v *CommandView) Render(cx *el.Context) el.Element {
	if !v.open {
		return el.Div().Hidden(true)
	}
	id := autoID("command", v)
	_, height := cx.ViewportSize()
	top := min(float32(80), height/8)
	viewport := max(float32(1), min(float32(360), height-top-72))
	text := locale.Current()
	v.buildRows()
	v.list.SetCount(len(v.rows))
	if v.active < 0 || v.active >= len(v.rows) || v.rowDisabled(v.active) {
		v.active = base.List{Count: len(v.rows), Disabled: v.rowDisabled}.First()
		if v.active >= 0 {
			v.list.ScrollTo(cx, v.active)
		}
	}
	keyHandler := func(e el.KeyEvent) bool {
		if e.Modifiers != 0 {
			return false
		}
		if v.nonsearchable && key.Name(e.Name) == key.NameReturn {
			if e.State == el.KeyPress {
				v.confirmActive()
			}
			return true
		}
		if v.inline && key.Name(e.Name) == key.NameEscape {
			if e.State == el.KeyPress {
				v.SetValue(false)
			}
			return true
		}
		i, ok := base.List{Count: len(v.rows), Disabled: v.rowDisabled}.Key(e.Name, v.active)
		if !ok {
			return false
		}
		if e.State == el.KeyPress && i >= 0 {
			v.active = i
			v.list.ScrollTo(cx, i)
		}
		return true
	}
	search := el.Input().ID(id + "/search").Name(text.SearchCommands).Placeholder(text.SearchCommands).Bind(&v.query).
		OnChange(func(string) { v.active = -1; v.searchChanged(); cx.ScrollTo(v.list.ID(), 0) }).
		OnSubmit(func(string) { v.confirmActive() }).OnKey(keyHandler)
	// Bound supplementary content separately so a long header/footer cannot hide
	// the list. The conservative reservation does not require a measurement pass.
	slotHeight := max(float32(1), min(float32(80), height/5))
	if v.header != nil {
		viewport = max(1, viewport-slotHeight)
	}
	if v.footer != nil {
		viewport = max(1, viewport-slotHeight)
	}

	var results el.Element
	switch {
	case v.loading:
		results = el.Div().P(theme.SpaceXl).Row().Gap(theme.SpaceMd).Child(Spinner().Render(cx), el.Text(text.Loading))
	case v.searchError != "":
		results = el.Div().P(theme.SpaceXl).Gap(theme.SpaceMd).Child(el.Text(v.searchError).TextColor(theme.Danger), Button(text.Retry, v.searchChanged).Render(cx))
	case len(v.rows) == 0:
		results = el.Div().Px(theme.SpaceXl).Py(10).Child(el.Text(text.NoMatches).TextColor(theme.Muted))
		if v.empty != nil {
			results = el.Div().MaxH(el.Dp(viewport)).ScrollY().Child(v.empty.Render(cx))
		}
	default:
		v.list.SetCount(len(v.rows))
		v.list.Height(min(viewport, float32(len(v.rows))*v.itemHeight()))
		results = el.Div().Role("listbox").Name(text.Commands).Items(el.Stretch).Child(v.list.Render(cx))
	}
	panel := floating(theme.ElevationLg).ID(id + "/panel").Role("group").Name(text.Commands).Disabled(v.disabled).W(el.Dp(560)).MaxW(el.Full).Items(el.Stretch)
	if v.nonsearchable {
		panel.Focusable(true).OnKey(keyHandler)
	}
	if v.header != nil {
		panel.Child(el.Div().ID(id + "/header").MaxH(el.Dp(slotHeight)).ScrollY().Child(v.header.Render(cx)))
	}
	if !v.nonsearchable {
		panel.Child(el.Div().P(theme.SpaceMd).Child(searchField(cx, id+"/searchbox", id+"/search", search)), el.Div().H(el.Dp(1)).Bg(theme.Border))
	}
	panel.Child(el.Div().ID(id + "/results").MaxH(el.Dp(viewport)).ScrollY().Items(el.Stretch).Child(results))
	if v.footer != nil {
		panel.Child(el.Div().ID(id + "/footer").MaxH(el.Dp(slotHeight)).ScrollY().Child(v.footer.Render(cx)))
	}
	if v.pendingFocus {
		cx.AfterEnabled(v.FocusID(), v, 0, func() {
			if v.open && !v.disabled {
				cx.Focus(v.FocusID())
			}
			v.pendingFocus = false
		})
	}
	if v.inline {
		return panel
	}
	panel.Role("dialog")
	cx.Overlay(id, el.Modal(el.Div().Pt(top).Items(el.Center).Child(panel)).Placement(el.Top, el.Center).OnDismiss(func() { v.SetValue(false) }))
	return el.Div().Hidden(true)
}

type commandRow struct {
	item   CommandItem
	header bool
}

func (v *CommandView) buildRows() {
	if v.cached && v.cachedQuery == v.query && v.cachedRevision == v.revision {
		return
	}
	v.cached, v.cachedQuery, v.cachedRevision = true, v.query, v.revision
	items := v.items
	if v.onSearch == nil && !v.nonsearchable {
		items = v.matches()
	}
	v.rows = v.rows[:0]
	group := ""
	for _, item := range items {
		if item.Group != "" && item.Group != group {
			v.rows = append(v.rows, commandRow{CommandItem{Title: item.Group}, true})
		}
		group = item.Group
		v.rows = append(v.rows, commandRow{item, false})
	}
}
func (v *CommandView) rowDisabled(i int) bool { return v.rows[i].header || v.rows[i].item.Disabled }
func (v *CommandView) row(cx *el.Context, i int) el.Element {
	entry := v.rows[i]
	it := entry.item
	if entry.header {
		// Group headings share the configured virtual slot height.
		return el.Div().H(el.Dp(v.itemHeight())).Px(theme.SpaceXl).Pb(theme.SpaceXs).Justify(el.End).Child(el.Text(it.Title).Bold().TextSize(theme.TextSm).TextColor(theme.Muted))
	}
	// Keep a 2dp margin on each side of the virtual slot.
	row := el.Div().Role("option").Name(it.Title).Selected(i == v.active).Disabled(it.Disabled).H(el.Dp(max(1, v.itemHeight()-4))).My(2).Mx(6).Px(10).Rounded(theme.RadiusMd).Row().Items(el.Center).Gap(theme.SpaceMd).Focusable(false)
	if !it.Disabled {
		row.Child(el.Div().ID("activate").Absolute().Top(0).Left(0).W(el.Full).H(el.Full).OnClick(func() { v.run(it) }))
	}
	var content el.View
	if v.renderItem != nil {
		content = v.renderItem(it, i == v.active)
	}
	if content != nil {
		row.Child(el.Div().ID("content").Grow().MinW(el.Dp(0)).Child(content.Render(cx)))
	} else {
		row.Child(el.Text(it.Title).Grow().MaxLines(1))
	}
	if !it.Disabled {
		row.CursorPointer()
		if i == v.active {
			row.Bg(theme.Highlight).TextColor(theme.PrimaryText)
		} else {
			row.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	if content == nil && it.Shortcut != "" {
		row.Child(Kbd(it.Shortcut).Render(cx))
	}
	return row
}
