package kit

import "testing"

func TestSidebarFilterKeepsParentsAndDropsEmptySections(t *testing.T) {
	v := Sidebar().
		Section("工作台", SidebarItem{ID: "inbox", Label: "Inbox"},
			SidebarItem{ID: "orders", Label: "Orders", Children: []SidebarItem{{ID: "pending", Label: "Pending orders"}}}).
		Section("团队", SidebarItem{ID: "members", Label: "Members"})
	h := sized(240, v)
	v.Filter("PEND")
	h.Frame()
	// A match below a collapsed parent shows the parent and opens it.
	if !shown(h, "Orders") || !shown(h, "Pending orders") || shown(h, "Inbox") {
		t.Fatal("filter should keep only the match and its parent")
	}
	if shown(h, "团队") || !shown(h, "工作台") {
		t.Fatal("a section without matches should drop its heading")
	}
	click(t, h, "Pending orders")
	if v.Value() != "pending" {
		t.Fatalf("selected %q", v.Value())
	}
	v.Filter("")
	h.Frame()
	if !shown(h, "Inbox") || !shown(h, "团队") || shown(h, "Pending orders") {
		t.Fatal("clearing the filter should restore every section and the collapsed parent")
	}
}

func TestSidebarItemWithoutIcon(t *testing.T) {
	v := Sidebar().Section("", SidebarItem{ID: "a", Label: "Alpha"})
	h := sized(240, v)
	if !shown(h, "Alpha") {
		t.Fatal("label")
	}
	v.SetCollapsed(true)
	h.Frame()
	if !shown(h, "A") {
		t.Fatal("collapsed item without an icon should show its first letter")
	}
}
