package kit

import (
	"slices"
	"strings"

	"github.com/dyike/keel/ui/el"
)

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
			if visible && !v.shown(it) {
				continue
			}
			out = append(out, it.ID)
			if !visible || v.open(it) {
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

// Filter shows only items whose label contains query, ignoring case, with the
// parents that lead to them; "" shows everything. Sections without a match
// drop out with their heading. The selection is kept even when hidden.
func (v *SidebarView) Filter(query string) { v.query = strings.ToLower(strings.TrimSpace(query)) }

func (v *SidebarView) matches(it SidebarItem) bool {
	return v.query == "" || strings.Contains(strings.ToLower(it.Label), v.query)
}

// shown reports whether the filter keeps it: it matches, or a descendant does.
func (v *SidebarView) shown(it SidebarItem) bool {
	return v.matches(it) || slices.ContainsFunc(it.Children, v.shown)
}

// open reports whether its children are listed: expanded, or filtering
// reaches a match below it.
func (v *SidebarView) open(it SidebarItem) bool {
	return v.expanded[it.ID] || v.query != "" && slices.ContainsFunc(it.Children, v.shown)
}
