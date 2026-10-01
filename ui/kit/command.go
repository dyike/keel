package kit

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// CommandItem is one entry of a command palette. Shortcut uses
// core.ParseShortcut syntax and is only displayed; Group is shown beside it.
type CommandItem struct {
	Title, Group, Shortcut string
	Action                 func()
}

// CommandView is a command palette: a search box over a list of commands that
// narrows as you type. Prefix matches rank first, then substrings, then
// characters that merely appear in order ("设置" finds "打开设置", "nwo"
// finds "New window"). ↑ ↓ move, Enter runs, Esc closes. Bind it to a
// shortcut yourself: cx.Shortcut("mod+k", palette.Toggle).
type CommandView struct {
	items  []CommandItem
	open   bool
	query  string
	active int
}

func Command(items ...CommandItem) *CommandView      { return &CommandView{items: items} }
func (v *CommandView) SetItems(items ...CommandItem) { v.items = items }
func (v *CommandView) Value() bool                   { return v.open }
func (v *CommandView) SetValue(open bool) {
	v.open = open
	if open {
		v.query, v.active = "", 0
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
	for j, r := range []rune(t) {
		if i < len(qs) && r == qs[i] {
			if prev == j-1 {
				s += 5
			}
			if j == 0 || unicode.IsSpace([]rune(t)[j-1]) {
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
	v.open = false
	if it.Action != nil {
		it.Action()
	}
}

func (v *CommandView) Render(cx *el.Context) el.Element {
	if !v.open {
		return el.Div().Hidden(true)
	}
	id := autoID("command", v)
	text := locale.Current()
	m := v.matches()
	v.active = min(max(v.active, 0), max(len(m)-1, 0))
	search := el.Input().ID(id+"/search").Name(text.SearchCommands).Placeholder(text.SearchCommands).Bind(&v.query).
		Border(0, theme.Border).Bg(theme.Surface).Px(16).Py(12).TextSize(16).
		OnChange(func(string) { v.active = 0 }).
		OnSubmit(func(string) {
			if v.active < len(m) {
				v.run(m[v.active])
			}
		}).
		OnKey(func(e el.KeyEvent) bool {
			if e.State == el.KeyPress && len(m) > 0 {
				i, _ := listKeys(e.Name, v.active, len(m), 8)
				v.active = i
				cx.ScrollIntoView(id+"/list", float32(i)*36, float32(i+1)*36)
			}
			return true
		})
	list := el.Div().ID(id + "/list").Role("listbox").Name(text.Commands).ScrollY().MaxH(el.Dp(360)).Py(4).Items(el.Stretch)
	if len(m) == 0 {
		list.Child(el.Div().Px(16).Py(10).Child(el.Text(text.NoMatches).TextColor(theme.Muted)))
	}
	for i, it := range m {
		it := it
		row := el.Div().Role("option").Name(it.Title).Selected(i == v.active).H(el.Dp(36)).Mx(6).Px(10).Rounded(6).
			Row().Items(el.Center).Gap(8).CursorPointer().Focusable(false).OnClick(func() { v.run(it) }).
			Child(el.Text(it.Title).Grow().MaxLines(1))
		if i == v.active {
			row.Bg(theme.Highlight).TextColor(theme.PrimaryText)
		} else {
			row.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
		if it.Group != "" {
			row.Child(el.Text(it.Group).TextSize(12).TextColor(theme.Muted))
		}
		if it.Shortcut != "" {
			row.Child(Kbd(it.Shortcut).Render(cx))
		}
		list.Child(row)
	}
	panel := surface().Role("dialog").Name(text.Commands).W(el.Dp(560)).MaxW(el.Full).Items(el.Stretch).
		Child(search, el.Div().H(el.Dp(1)).Bg(theme.Border), list)
	// A strip of padding keeps the palette off the window's top edge.
	cx.Overlay(id, el.Modal(el.Div().Pt(80).Items(el.Center).Child(panel)).Placement(el.Top, el.Center).
		OnDismiss(func() { v.open = false }))
	return el.Div().Hidden(true) // the modal layer focuses the search box when it opens
}
