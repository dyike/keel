package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"image/color"
	"slices"
	"strconv"
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
// Keywords adds extra fuzzy-search terms. Constructors and SetItems copy the slice.
type CommandItem struct {
	Title, Group, Shortcut string
	// ActionName resolves the default row hint through core.Bindings.
	// Register the same Action with cx.Action to enable it outside this palette.
	ActionName string
	Icon       IconName
	Checked    bool
	Action     func()
	Disabled   bool
	Keywords   []string
	// Separator introduces a non-interactive divider; other fields are ignored.
	Separator bool
}

// CommandView is a command palette: a search box over a list of commands that
// narrows as you type. Prefix matches rank first, then substrings, then
// characters that merely appear in order ("设置" finds "打开设置", "nwo"
// finds "New window"). ↑ ↓ move, Enter runs, Esc closes. Bind it to a
// shortcut yourself: cx.Shortcut("mod+k", palette.Toggle).
type CommandView struct {
	list                                *VirtualListView
	variable                            *VariableListView
	autoRows, revealActive              bool
	inline, nonsearchable, pendingFocus bool
	header, footer, empty               el.View
	renderItem                          func(CommandItem, bool) el.View
	rowHeight                           float32
	maxHeight                           float32
	placeholder                         string
	borderless                          bool
	panelStyle                          func(*el.DivEl)
	hoveredRow, lastHovered             int
	themeRevision                       uint64
	onSelect, onConfirm                 func(int)
	onCancel                            func()
	onQuery                             func(string)
	selected                            int
	selectionPending                    bool
	selectionVersion                    uint64
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
	v := &CommandView{items: cloneCommandItems(items), active: -1, selected: -1, lastHovered: -1}
	v.list = VirtualList(0, 36, v.row).ItemKey(func(i int) string {
		prefix := "item:"
		if v.rows[i].header {
			prefix = "group:"
		}
		return prefix + strconv.Itoa(v.rows[i].index)
	})
	v.variable = VariableList(nil, 36, v.row)
	return v
}
func (v *CommandView) SetItems(items ...CommandItem) {
	v.items = cloneCommandItems(items)
	v.revision++
	v.variable.Invalidate()
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
	if !open {
		v.selectionPending = false
	}
	v.pendingFocus = open && !v.inline
	v.request++ // Invalidate results from an earlier opening.
	if open {
		v.query = ""
		v.selected = -1
		v.lastHovered = -1
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

func (v *CommandView) matches() []int {
	type scored struct{ index, score int }
	var out []scored
	var indices []int
	flush := func() {
		slices.SortStableFunc(out, func(a, b scored) int { return b.score - a.score })
		for _, entry := range out {
			indices = append(indices, entry.index)
		}
		out = out[:0]
	}
	for index, it := range v.items {
		if it.Separator {
			flush()
			indices = append(indices, index)
			continue
		}
		score, ok := fuzzy(v.query, it.Title)
		for _, keyword := range it.Keywords {
			if s, match := fuzzy(v.query, keyword); match && (!ok || s > score) {
				score, ok = s, true
			}
		}
		if ok {
			out = append(out, scored{index, score})
		}
	}
	flush()
	return indices
}

func (v *CommandView) run(entry commandRow) {
	it := entry.item
	if !v.open || it.Disabled || v.disabled || v.loading || v.searchError != "" {
		return
	}
	confirm := v.onConfirm
	if !v.inline {
		v.SetValue(false)
	}
	if it.Action != nil {
		it.Action()
	}
	if confirm != nil {
		confirm(entry.index)
	}
}

func (v *CommandView) Render(cx *el.Context) el.Element {
	if !v.open {
		return el.Div().Hidden(true)
	}
	id := autoID("command", v)
	_, height := cx.ViewportSize()
	top := min(float32(80), height/8)
	limit := v.maxHeight
	if limit == 0 {
		limit = 360
	}
	viewport := max(float32(1), min(limit, height-top-72))
	if v.themeRevision != theme.Revision() {
		v.variable.Invalidate()
		v.themeRevision = theme.Revision()
	}
	v.hoveredRow = -1
	text := locale.Current()
	v.buildRows()
	v.list.SetCount(len(v.rows))
	if v.active < 0 || v.active >= len(v.rows) || v.rowDisabled(v.active) {
		v.selectActive(base.List{Count: len(v.rows), Disabled: v.rowDisabled}.First())
		if v.active >= 0 {
			v.scrollTo(cx, v.active)
		}
	}
	if v.revealActive {
		v.scrollTo(cx, v.active)
		v.revealActive = false
	}
	keyHandler := func(e el.KeyEvent) bool {
		if v.boundKey(e) {
			return true
		}
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
				v.escape()
			}
			return true
		}
		i, ok := base.List{Count: len(v.rows), Disabled: v.rowDisabled, Wrap: true}.Key(e.Name, v.active)
		if !ok {
			return false
		}
		if e.State == el.KeyPress && i >= 0 {
			v.selectActive(i)
			v.scrollTo(cx, i)
			v.notifySelection()
		}
		return true
	}
	placeholder := v.placeholder
	if placeholder == "" {
		placeholder = text.SearchCommands
	}
	search := el.Input().ID(id + "/search").Name(text.SearchCommands).Placeholder(placeholder).Bind(&v.query).
		OnChange(func(string) { v.queryChanged(); cx.ScrollTo(v.listID(), 0) }).
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
		results = el.Div().Role("listbox").Name(text.Commands).Items(el.Stretch).Child(v.renderList(cx, viewport))
	}
	hovered := -1
	if v.hoveredRow >= 0 {
		hovered = v.rows[v.hoveredRow].index
	}
	if hovered != v.lastHovered && v.hoveredRow >= 0 {
		v.selectActive(v.hoveredRow)
	}
	v.lastHovered = hovered
	v.queueSelection(cx)
	// Resolve bindings each render so hints and handlers follow rebinding.
	for _, entry := range v.rows {
		if entry.header || entry.item.Separator || entry.item.Disabled || entry.item.ActionName == "" || !cx.FocusWithin(id+"/panel") || !cx.Enabled(id+"/panel") {
			continue
		}
		cx.Action(entry.item.ActionName, func() {
			if cx.Enabled(id+"/panel") && cx.FocusWithin(id+"/panel") {
				v.run(entry)
			}
		})
	}
	panel := floating(theme.ElevationLg).ID(id + "/panel").Role("group").Name(text.Commands).Disabled(v.disabled).W(el.Dp(560)).MaxW(el.Full).Items(el.Stretch)
	if v.borderless {
		panel.Border(0, color.NRGBA{}).Rounded(0).Shadow(theme.Elevation{})
	}
	if v.panelStyle != nil {
		v.panelStyle(panel)
	}
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
	cx.Overlay(id, el.Modal(el.Div().Pt(top).Items(el.Center).Child(panel)).Placement(el.Top, el.Center).OnEscape(func() bool { return v.clearQuery() }).OnDismiss(v.cancel))
	return el.Div().Hidden(true)
}

type commandRow struct {
	item   CommandItem
	header bool
	index  int
}

func (v *CommandView) buildRows() {
	if v.cached && v.cachedQuery == v.query && v.cachedRevision == v.revision {
		return
	}
	v.cached, v.cachedQuery, v.cachedRevision = true, v.query, v.revision
	indices := make([]int, len(v.items))
	for i := range indices {
		indices[i] = i
	}
	if v.onSearch == nil && !v.nonsearchable {
		indices = v.matches()
	}
	v.rows = v.rows[:0]
	group := ""
	pending := -1
	for _, index := range indices {
		item := v.items[index]
		if item.Separator {
			pending = index
			continue
		}
		if pending >= 0 && len(v.rows) > 0 {
			v.rows = append(v.rows, commandRow{item: CommandItem{Separator: true}, index: pending})
			group = ""
		}
		pending = -1
		if item.Group != "" && item.Group != group {
			v.rows = append(v.rows, commandRow{item: CommandItem{Title: item.Group}, header: true, index: index})
		}
		group = item.Group
		v.rows = append(v.rows, commandRow{item: item, index: index})
	}
	keys := make([]string, len(v.rows))
	for i, row := range v.rows {
		prefix := "item:"
		if row.header {
			prefix = "group:"
		}
		if row.item.Separator {
			prefix = "separator:"
		}
		keys[i] = prefix + strconv.Itoa(row.index)
	}
	v.variable.SetKeys(keys)
	// Preserve original model coordinates when entries are refreshed or filtered.
	if v.active < 0 && v.selected >= 0 {
		for i, row := range v.rows {
			if !v.rowDisabled(i) && row.index == v.selected {
				v.active = i
				break
			}
		}
	}
}
func (v *CommandView) rowDisabled(i int) bool {
	return v.rows[i].header || v.rows[i].item.Separator || v.rows[i].item.Disabled
}
func (v *CommandView) row(cx *el.Context, i int) el.Element {
	entry := v.rows[i]
	it := entry.item
	if it.Separator {
		height := v.itemHeight()
		if v.autoRows {
			height = 9
		}
		return el.Div().Role("separator").H(el.Dp(height)).Px(theme.SpaceMd).Justify(el.Center).Child(el.Div().H(el.Dp(1)).Bg(theme.Border))
	}
	if entry.header {
		// Group headings share the configured virtual slot height.
		return el.Div().H(el.Dp(v.itemHeight())).Px(theme.SpaceXl).Pb(theme.SpaceXs).Justify(el.End).Child(el.Text(it.Title).Bold().TextSize(theme.TextSm).TextColor(theme.Muted))
	}
	// Keep a 2dp margin on each side of the virtual slot.
	rowID := autoID("command", v) + "/item:" + strconv.Itoa(entry.index)
	if !it.Disabled && !v.loading && v.searchError == "" && cx.Hovered(rowID) {
		v.hoveredRow = i
	}
	row := el.Div().ID(rowID).Role("option").Name(it.Title).Selected(i == v.active).Disabled(it.Disabled).H(el.Dp(max(1, v.itemHeight()-4))).My(2).Mx(6).Px(10).Rounded(theme.RadiusMd).Row().Items(el.Center).Gap(theme.SpaceMd).Focusable(false)
	if v.autoRows {
		row.H(el.Auto).MinH(el.Dp(max(1, v.itemHeight()-4)))
	}
	if !it.Disabled {
		row.Child(el.Div().ID("activate").Absolute().Top(0).Left(0).W(el.Full).H(el.Full).OnClick(func() {
			revision, request := v.revision, v.request
			v.selectActive(i)
			v.notifySelection()
			if revision == v.revision && request == v.request {
				v.run(entry)
			}
		}))
	}
	var content el.View
	if v.renderItem != nil {
		copy := it
		copy.Keywords = slices.Clone(it.Keywords)
		content = v.renderItem(copy, i == v.active)
	}
	if content != nil {
		row.Child(el.Div().ID("content").Grow().MinW(el.Dp(0)).Child(content.Render(cx)))
	} else {
		if it.Icon != IconNone {
			row.Child(Icon(it.Icon).Size(16).Render(cx))
		}
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
	if content == nil {
		shortcut := it.Shortcut
		if it.ActionName != "" {
			shortcut = ""
			if bindings := core.Bindings(it.ActionName); len(bindings) > 0 {
				shortcut = bindings[0]
			}
		}
		if shortcut != "" {
			row.Child(Kbd(shortcut).Render(cx))
		} else if it.Checked {
			row.Child(Icon(IconDone).Size(16).Render(cx))
		}
	}
	return row
}
