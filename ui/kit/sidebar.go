package kit

import (
	"slices"
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SidebarItem is one destination in a Sidebar. IDs must be unique.
type SidebarItem struct {
	ID, Label string
	Icon      IconName // optional; IconNone shows no icon
	Disabled  bool
	Children  []SidebarItem
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
	sections       []sidebarSection
	selected       string
	collapsed      bool
	width          float32
	height         float32
	disabled       bool
	header, footer el.View
	query          string
	expanded       map[string]bool
	positions      map[string]float32
	revealID       string
	tips           map[string]*TooltipView
	onChange       func(id string)
}

func Sidebar() *SidebarView {
	return &SidebarView{width: 220, tips: map[string]*TooltipView{}, expanded: map[string]bool{}}
}

// Section adds a titled group of items; an empty title adds no heading.
func (v *SidebarView) Section(title string, items ...SidebarItem) *SidebarView {
	seen := map[string]bool{}
	for _, id := range v.allIDs() {
		seen[id] = true
	}
	if copy, ok := copySidebarItems(items, seen); ok {
		v.sections = append(v.sections, sidebarSection{title, copy})
	}
	return v
}

// Width sets the expanded width in dp, 220 by default.
func (v *SidebarView) Width(dp float32) *SidebarView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.width = dp
	}
	return v
}
func (v *SidebarView) OnChange(fn func(id string)) *SidebarView { v.onChange = fn; return v }
func (v *SidebarView) Value() string                            { return v.selected }
func (v *SidebarView) SetValue(id string) {
	v.selected = id
	v.revealID = id
	v.expandParents(id)
}
func (v *SidebarView) Collapsed() bool      { return v.collapsed }
func (v *SidebarView) SetCollapsed(on bool) { v.collapsed = on }

// SetBadge updates an item's count.
func (v *SidebarView) SetBadge(id string, n int) {
	if it := v.find(id); it != nil {
		it.Badge = n
	}
}

func (v *SidebarView) choose(id string) {
	if it := v.find(id); it == nil || it.Disabled || v.disabled || id == v.selected {
		return
	}
	v.selected = id
	if v.onChange != nil {
		v.onChange(id)
	}
}

func (v *SidebarView) item(cx *el.Context, prefix string, it SidebarItem, ids []string, depth int) el.Element {
	on := it.ID == v.selected
	disabled := v.disabled || it.Disabled
	fg := theme.Muted
	if on {
		fg = theme.PrimaryText
	}
	if disabled {
		fg = theme.Muted
	}
	row := el.Div().ID(prefix + "/" + it.ID).Role("link").Name(it.Label).Selected(on).Disabled(disabled).
		Row().Items(el.Center).Gap(10).H(el.Dp(36)).Px(10).Pl(float32(10 + depth*14)).Rounded(theme.RadiusLg).CursorPointer().TextColor(theme.Text).TextSize(theme.TextControl).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnClick(func() {
			if len(it.Children) > 0 {
				v.expanded[it.ID] = !v.expanded[it.ID]
			} else {
				v.choose(it.ID)
			}
		}).
		OnKey(func(e el.KeyEvent) bool {
			if e.Modifiers != 0 {
				return false
			}
			if key.Name(e.Name) == key.NameRightArrow && len(it.Children) > 0 {
				if e.State == el.KeyPress {
					v.expanded[it.ID] = true
				}
				return true
			}
			if key.Name(e.Name) == key.NameLeftArrow {
				if e.State == el.KeyPress {
					if v.expanded[it.ID] {
						v.expanded[it.ID] = false
					} else if parent := v.parent(it.ID); parent != "" {
						v.focusItem(cx, parent)
					}
				}
				return true
			}
			at := 0
			for i, id := range ids {
				if id == it.ID {
					at = i
					break
				}
			}
			j, ok := base.List{Count: len(ids), Disabled: func(i int) bool { return v.find(ids[i]).Disabled }}.Key(e.Name, at)
			if ok && e.State == el.KeyPress && j >= 0 {
				v.focusItem(cx, ids[j])
			}
			return ok
		})
	if it.Icon != IconNone {
		row.Child(Icon(it.Icon).Size(16).Color(fg).Render(cx))
	} else if v.collapsed {
		row.Child(el.Text(string([]rune(it.Label)[:min(1, len([]rune(it.Label)))])))
	}
	if len(it.Children) > 0 {
		row.Role("button").Value(strconv.FormatBool(v.expanded[it.ID]))
	}
	if disabled {
		row.TextColor(theme.Muted)
	}
	if on {
		row.Bg(theme.Subtle).Bold()
	} else {
		row.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	}
	if !v.collapsed {
		if len(it.Children) > 0 {
			icon := IconChevronRight
			if v.expanded[it.ID] {
				icon = IconChevronDown
			}
			row.Child(Icon(icon).Size(14).Color(fg).Render(cx))
		}
		row.Child(el.Text(it.Label).Grow().MaxLines(1))
		if it.Badge > 0 {
			count := strconv.Itoa(it.Badge)
			if it.Badge > 99 {
				count = "99+"
			}
			row.Child(el.Div().Role("badge").Name(strconv.Itoa(it.Badge)).Value(count).
				MinW(el.Dp(24)).H(el.Dp(20)).Center().NoShrink().Rounded(theme.RadiusSm).
				TextSize(theme.TextXs).TextColor(theme.Muted).Child(el.Text(count)))
		}
		return row
	}
	// Collapsed: icon only, the label in a tooltip; a dot stands for the badge.
	row.Justify(el.Center).Px(0).WFull()
	if it.Badge > 0 {
		row = el.Div().WFull().Items(el.Stretch).Child(row, el.Div().Absolute().Top(4).Right(8).Child(Badge(1).Dot().Tone(ToneInfo).Render(cx)))
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
	nav := el.Div().Role("navigation").Name(text.Menu).W(el.Dp(width)).NoShrink().When(theme.BgGradient.IsZero(), func(d *el.DivEl) { d.Bg(theme.Bg) }).Px(theme.SpaceLg).Py(theme.SpaceLg).Gap(theme.SpaceMd).Items(el.Stretch).Disabled(v.disabled)
	if v.collapsed {
		nav.Px(theme.SpaceMd)
	}
	if v.height > 0 {
		nav.H(el.Dp(v.height))
	}
	if v.header != nil {
		nav.Child(el.Div().ID(base + "/header").NoShrink().Child(v.header.Render(cx)))
	}
	// Put the 10dp scrollbar hit area in the outer gutter, leaving 2dp
	// between it and the rows. Extend into the existing navigation padding
	// so expanded labels keep their available width.
	body := el.Div().ID(base + "/scroll").Grow().MinH(el.Dp(0)).ScrollY().Gap(theme.SpaceXs).Items(el.Stretch).Mx(-12).Px(theme.SpaceLg)
	if v.collapsed {
		body.Mx(-8).Pl(theme.SpaceMd)
	}
	_, viewportHeight := cx.ViewportSize()
	nav.MaxH(el.Dp(viewportHeight))
	if v.revealID != "" {
		v.expandParents(v.revealID)
	}
	ids := v.ids()
	v.positions = map[string]float32{}
	y := float32(0)
	var addItems func([]SidebarItem, int)
	addItems = func(items []SidebarItem, depth int) {
		for _, it := range items {
			if !v.shown(it) {
				continue
			}
			v.positions[it.ID] = y
			y += 40
			body.Child(v.item(cx, base, it, ids, depth))
			if v.open(it) {
				addItems(it.Children, depth+1)
			}
		}
	}
	for i, s := range v.sections {
		if !slices.ContainsFunc(s.items, v.shown) {
			continue
		}
		if s.title != "" && !v.collapsed {
			body.Child(el.Div().H(el.Dp(28)).Px(10).Justify(el.Center).Child(el.Text(s.title).TextSize(theme.TextSm).TextColor(theme.Muted)))
			y += 32
		} else if i > 0 {
			body.Child(el.Div().H(el.Dp(17)).Px(theme.SpaceMd).Justify(el.Center).Child(el.Div().H(el.Dp(1)).Bg(theme.Border)))
			y += 21
		}
		addItems(s.items, 0)
	}
	nav.Child(body)
	if v.footer != nil {
		nav.Child(el.Div().ID(base + "/footer").NoShrink().Child(v.footer.Render(cx)))
	}
	name, icon := text.CollapseSidebar, IconChevronLeft
	if v.collapsed {
		name, icon = text.ExpandSidebar, IconChevronRight
	}
	label := name
	if v.collapsed {
		label = ""
	}
	nav.Child(el.Div().H(el.Dp(1)).NoShrink().Mx(8).Bg(theme.Border),
		el.Div().NoShrink().Items(el.Stretch).Child(Button(label, func() { v.collapsed = !v.collapsed; v.revealID = v.selected }).Name(name).Icon(icon).Variant(ButtonGhost).Size(28).Render(cx)))
	if v.revealID != "" {
		target := v.revealID
		cx.AfterEnabled(base+"/scroll", sidebarRevealKey{v, target}, 0, func() {
			if top, ok := v.positions[target]; ok {
				cx.ScrollIntoView(base+"/scroll", top, top+36)
			}
			v.revealID = ""
		})
	}
	return el.Div().Row().Items(el.Stretch).Child(nav, el.Div().W(el.Dp(1)).Bg(theme.Border))
}
