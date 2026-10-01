package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SidebarItem is one destination in a Sidebar. IDs must be unique.
type SidebarItem struct {
	ID, Label string
	Icon      IconName
	Badge     int // a count shown at the end; 0 hides it
}

type sidebarSection struct {
	title string
	items []SidebarItem
}

// SidebarView is the app's navigation column: sections of items, one
// selected. Collapsed it shows only icons, each with a tooltip. Items take
// focus with Tab; ↑ ↓ move between them.
type SidebarView struct {
	sections  []sidebarSection
	selected  string
	collapsed bool
	width     float32
	tips      map[string]*TooltipView
	onChange  func(id string)
}

func Sidebar() *SidebarView { return &SidebarView{width: 220, tips: map[string]*TooltipView{}} }

// Section adds a titled group of items; an empty title adds no heading.
func (v *SidebarView) Section(title string, items ...SidebarItem) *SidebarView {
	v.sections = append(v.sections, sidebarSection{title, items})
	return v
}

// Width sets the expanded width in dp, 220 by default.
func (v *SidebarView) Width(dp float32) *SidebarView {
	if dp > 0 {
		v.width = dp
	}
	return v
}
func (v *SidebarView) OnChange(fn func(id string)) *SidebarView { v.onChange = fn; return v }
func (v *SidebarView) Value() string                            { return v.selected }
func (v *SidebarView) SetValue(id string)                       { v.selected = id }
func (v *SidebarView) Collapsed() bool                          { return v.collapsed }
func (v *SidebarView) SetCollapsed(on bool)                     { v.collapsed = on }

// SetBadge updates an item's count.
func (v *SidebarView) SetBadge(id string, n int) {
	for s := range v.sections {
		for i := range v.sections[s].items {
			if v.sections[s].items[i].ID == id {
				v.sections[s].items[i].Badge = n
			}
		}
	}
}

func (v *SidebarView) ids() []string {
	var out []string
	for _, s := range v.sections {
		for _, it := range s.items {
			out = append(out, it.ID)
		}
	}
	return out
}

func (v *SidebarView) choose(id string) {
	if id == v.selected {
		return
	}
	v.selected = id
	if v.onChange != nil {
		v.onChange(id)
	}
}

func (v *SidebarView) item(cx *el.Context, base string, it SidebarItem, ids []string) el.Element {
	on := it.ID == v.selected
	fg := theme.Text
	if on {
		fg = theme.PrimaryText
	}
	row := el.Div().ID(base + "/" + it.ID).Role("link").Name(it.Label).Selected(on).
		Row().Items(el.Center).Gap(10).H(el.Dp(34)).Px(10).Rounded(6).CursorPointer().TextColor(fg).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnClick(func() { v.choose(it.ID) }).
		OnKey(func(e el.KeyEvent) bool {
			at := 0
			for i, id := range ids {
				if id == it.ID {
					at = i
				}
			}
			j, ok := listKeys(e.Name, at, len(ids), 0)
			if key.Name(e.Name) == key.NamePageUp || key.Name(e.Name) == key.NamePageDown {
				ok = false
			}
			if ok && e.State == el.KeyPress {
				cx.Focus(base + "/" + ids[j])
			}
			return ok
		}).
		Child(Icon(it.Icon).Size(18).Color(fg).Render(cx))
	if on {
		row.Bg(theme.Highlight)
	} else {
		row.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	}
	if !v.collapsed {
		row.Child(el.Text(it.Label).Grow().MaxLines(1))
		if it.Badge > 0 {
			row.Child(Badge(it.Badge).Tone(ToneInfo).Render(cx))
		}
		return row
	}
	// Collapsed: icon only, the label in a tooltip; a dot stands for the badge.
	row.Justify(el.Center).Px(0)
	if it.Badge > 0 {
		row = el.Div().Items(el.Stretch).Child(row, el.Div().Absolute().Top(4).Right(8).Child(Badge(1).Dot().Tone(ToneInfo).Render(cx)))
	}
	tip := v.tips[it.ID]
	if tip == nil {
		tip = &TooltipView{}
		v.tips[it.ID] = tip
	}
	tip.text, tip.target = it.Label, el.ViewFunc(func(*el.Context) el.Element { return row })
	return tip.Render(cx)
}

func (v *SidebarView) Render(cx *el.Context) el.Element {
	base := autoID("sidebar", v)
	text := locale.Current()
	width := v.width
	if v.collapsed {
		width = 56
	}
	ids := v.ids()
	nav := el.Div().Role("navigation").Name(text.Menu).W(el.Dp(width)).NoShrink().Bg(theme.Surface).
		Px(8).Py(12).Gap(2).Items(el.Stretch)
	for i, s := range v.sections {
		if s.title != "" && !v.collapsed {
			top := float32(8)
			if i == 0 {
				top = 0
			}
			nav.Child(el.Div().Px(10).Pt(top).Pb(4).Child(el.Text(s.title).TextSize(12).TextColor(theme.Muted)))
		} else if i > 0 {
			nav.Child(el.Div().H(el.Dp(1)).My(6).Bg(theme.Border))
		}
		for _, it := range s.items {
			nav.Child(v.item(cx, base, it, ids))
		}
	}
	name, icon := text.CollapseSidebar, IconChevronLeft
	if v.collapsed {
		name, icon = text.ExpandSidebar, IconChevronRight
	}
	nav.Child(el.Div().Grow(), el.Div().Items(el.Start).Child(
		Button("", func() { v.collapsed = !v.collapsed }).Name(name).Icon(icon).Variant(ButtonGhost).Size(28).Render(cx)))
	return el.Div().Row().Items(el.Stretch).Child(nav, el.Div().W(el.Dp(1)).Bg(theme.Border))
}
