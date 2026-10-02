package kit

import "github.com/dyike/keel/ui/el"

func (v *SidebarView) Header(view el.View) *SidebarView { v.header = view; return v }
func (v *SidebarView) Footer(view el.View) *SidebarView { v.footer = view; return v }

// Height fixes the sidebar height; by default it sizes to content within the viewport.
func (v *SidebarView) Height(dp float32) *SidebarView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}
func (v *SidebarView) SetDisabled(on bool) { v.disabled = on }
func (v *SidebarView) SetItemDisabled(id string, on bool) {
	if it := v.find(id); it != nil {
		it.Disabled = on
	}
}
func (v *SidebarView) Expanded(id string) bool { return v.expanded[id] }
func (v *SidebarView) SetExpanded(id string, on bool) {
	if it := v.find(id); it != nil && len(it.Children) > 0 {
		v.expanded[id] = on
	}
}
func (v *SidebarView) find(id string) *SidebarItem {
	var walk func([]SidebarItem) *SidebarItem
	walk = func(items []SidebarItem) *SidebarItem {
		for i := range items {
			if items[i].ID == id {
				return &items[i]
			}
			if found := walk(items[i].Children); found != nil {
				return found
			}
		}
		return nil
	}
	for _, s := range v.sections {
		if found := walk(s.items); found != nil {
			return found
		}
	}
	return nil
}
func (v *SidebarView) parent(id string) string {
	var walk func([]SidebarItem, string) (string, bool)
	walk = func(items []SidebarItem, parent string) (string, bool) {
		for _, it := range items {
			if it.ID == id {
				return parent, true
			}
			if p, ok := walk(it.Children, it.ID); ok {
				return p, true
			}
		}
		return "", false
	}
	for _, s := range v.sections {
		if p, ok := walk(s.items, ""); ok {
			return p
		}
	}
	return ""
}
func (v *SidebarView) expandParents(id string) {
	for p := v.parent(id); p != ""; p = v.parent(p) {
		v.expanded[p] = true
	}
}
func (v *SidebarView) ids() []string    { return v.collectIDs(true) }
func (v *SidebarView) allIDs() []string { return v.collectIDs(false) }
func (v *SidebarView) collectIDs(visible bool) []string {
	var out []string
	var walk func([]SidebarItem)
	walk = func(items []SidebarItem) {
		for _, it := range items {
			out = append(out, it.ID)
			if !visible || v.expanded[it.ID] {
				walk(it.Children)
			}
		}
	}
	for _, s := range v.sections {
		walk(s.items)
	}
	return out
}
func copySidebarItems(items []SidebarItem, seen map[string]bool) ([]SidebarItem, bool) {
	out := make([]SidebarItem, len(items))
	for i, it := range items {
		if it.ID == "" || seen[it.ID] {
			return nil, false
		}
		seen[it.ID] = true
		children, ok := copySidebarItems(it.Children, seen)
		if !ok {
			return nil, false
		}
		out[i] = it
		out[i].Children = children
	}
	return out, true
}
func (v *SidebarView) focusItem(cx *el.Context, id string) {
	if it := v.find(id); it == nil || it.Disabled {
		return
	}
	base := autoID("sidebar", v)
	if top, ok := v.positions[id]; ok {
		cx.ScrollIntoView(base+"/scroll", top, top+36)
	}
	cx.Focus(base + "/" + id)
}

type sidebarRevealKey struct {
	view *SidebarView
	id   string
}
