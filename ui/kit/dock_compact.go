package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// Compact presents open panels as one horizontally scrollable tab strip and
// one full-width body. It preserves the desktop layout, including collapsed
// regions and zoom. Moving, splitting and detaching remain desktop operations.
// The host chooses the breakpoint; changing presentation does not call OnLayoutChange.
func (v *DockView) Compact(on bool) *DockView {
	if v.compact != on {
		v.cancelResize()
		v.drag = dockDrag{}
		v.compact = on
		if v.compactMenu != nil {
			v.compactMenu.SetValue(false)
		}
		for _, menu := range v.menus {
			menu.SetValue(false)
		}
		clear(v.groupRects)
		clear(v.tabRects)
		clear(v.tabOrigins)
	}
	return v
}

func (v *DockView) renderCompact(cx *el.Context) el.Element {
	id := autoID("dock", v)
	var ids []string
	for _, side := range []DockSide{DockCenter, DockLeft, DockRight, DockBottom} {
		panels, _ := v.side(side)
		for _, p := range *panels {
			if v.Visible(p) {
				ids = append(ids, p)
			}
		}
	}
	if !slices.Contains(ids, v.compactActive) {
		v.compactActive = ""
		if slices.Contains(ids, v.layout.Zoomed) {
			v.compactActive = v.layout.Zoomed
		} else if len(ids) > 0 {
			v.compactActive = ids[0]
		}
	}
	box := el.Div().ID(id).Role("group").Disabled(v.disabled).Grow().Items(el.Stretch)
	if len(ids) == 0 {
		if v.center != nil {
			box.Child(v.center.Render(cx))
		}
		return box
	}
	tabs := el.Div().ID(id + "/compact-tabs").Role("tablist").Row().ScrollX().NoShrink().Gap(theme.SpaceXxs).Pb(scrollbarGutter)
	for _, panelID := range ids {
		p := v.panels[panelID]
		on := panelID == v.compactActive
		tab := el.Div().ID(v.tabID(panelID)).Role("tab").Name(p.Title).Selected(on).Focusable(true).CursorPointer().NoShrink().MinH(el.Dp(44)).Px(12).Py(8).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).OnClick(func() { v.compactActive = panelID })
		if p.Tab != nil {
			tab.Child(p.Tab(on).Render(cx))
		} else {
			if p.Icon != IconNone {
				tab.Row().Items(el.Center).Gap(theme.SpaceXs).Child(Icon(p.Icon).Size(14).Render(cx))
			}
			tab.Child(el.Text(p.Title).MaxLines(1))
		}
		if on {
			tab.Bg(theme.Highlight)
		}
		if v.skin != nil && v.skin.Tab != nil {
			v.skin.Tab(tab, on)
		}
		tabs.Child(tab)
	}
	p := v.panels[v.compactActive]
	head := el.Div().NoShrink().Items(el.Stretch).Bg(theme.Subtle).Child(tabs)
	tools := el.Div().Row().Wrap().NoShrink().Items(el.Center).Gap(theme.SpaceXs)
	if p.Toolbar != nil {
		tools.Child(p.Toolbar.Render(cx))
	}
	if p.Menu != nil {
		if v.compactMenu == nil {
			v.compactMenu = Menu()
		}
		m := v.compactMenu
		m.items = m.items[:0]
		p.Menu(m)
		m.Trigger(Button("", m.Toggle).Name(locale.Current().Name(locale.Current().More, p.Title)).Icon(IconChevronDown).Variant(ButtonGhost).Size(44))
		tools.Child(m.Render(cx))
	}
	if !p.NoClose {
		tools.Child(Button(locale.Current().Close, func() {
			v.SetVisible(p.ID, false)
			v.changed()
		}).Variant(ButtonGhost).Size(44).Render(cx))
	}
	head.Child(tools)
	body := el.Div().Role("region").Name(p.Title).Grow().H(el.Dp(0)).Items(el.Stretch)
	if !p.NoPadding {
		body.P(theme.SpaceMd)
	}
	if p.View != nil {
		body.Child(p.View.Render(cx))
	}
	if v.skin != nil {
		if v.skin.Header != nil {
			v.skin.Header(head)
		}
		if v.skin.Body != nil {
			v.skin.Body(body)
		}
		if v.skin.Panel != nil {
			v.skin.Panel(box)
		}
	}
	return box.Child(head, body)
}
