package kit

import "github.com/dyike/keel/ui/el"

// Side sets the application edge, Left by default. The application places the
// sidebar in its containing row; this changes the border, toggle and tooltip.
func (v *SidebarView) Side(side el.Side) *SidebarView {
	if side == el.Left || side == el.Right {
		v.side = side
	}
	return v
}

// BorderWidth sets the inner dividing line in dp; zero hides it.
func (v *SidebarView) BorderWidth(dp float32) *SidebarView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.borderWidth = dp
	}
	return v
}

// Collapsible controls the built-in toggle. SetCollapsed remains available.
func (v *SidebarView) Collapsible(on bool) *SidebarView { v.collapsible = on; return v }

// SetSuffix replaces an item's independent trailing content. Nil removes it.
func (v *SidebarView) SetSuffix(id string, content el.View) {
	if it := v.find(id); it != nil {
		it.Suffix = content
	}
}

// SetContextMenu replaces an item's menu and closes the previous one. The menu
// belongs to this item; its Trigger is unused. Nil removes the menu.
func (v *SidebarView) SetContextMenu(id string, menu *MenuView) {
	if it := v.find(id); it != nil && it.ContextMenu != menu {
		if it.ContextMenu != nil {
			it.ContextMenu.SetValue(false)
		}
		it.ContextMenu = menu
	}
}

func (v *SidebarView) itemMenu(cx *el.Context, it SidebarItem, content el.Element) el.Element {
	menu := it.ContextMenu
	if menu == nil {
		return content
	}
	id := autoID("menu", menu)
	owner := autoID("sidebar", v) + "/" + it.ID
	disabled := v.disabled || it.Disabled
	if disabled {
		menu.SetValue(false)
	}
	if menu.open {
		cx.Overlay(id, el.Anchored(id, menu.panel(cx)).Owner(owner).
			Placement(menu.side, menu.align).Offset(menu.offset).Modal().TrapFocus().
			OnDismiss(func() { menu.SetValue(false) }))
		menu.renderSub(cx)
	}
	return el.Div().ID(id).WFull().Items(el.Stretch).Disabled(disabled).
		OnContextMenu(func() { menu.SetValue(true) }).Child(content)
}
