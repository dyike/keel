package kit

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestSidebarSlotsStayVisibleAndKeyboardScrolls(t *testing.T) {
	items := make([]SidebarItem, 80)
	for i := range items {
		items[i] = SidebarItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Item %d", i)}
	}
	v := Sidebar().Height(280).Header(Button("Workspace", func() {})).Footer(Button("Profile", func() {})).Section("", items...)
	h := page(v)
	click(t, h, "Item 0")
	h.Key(key.NameEnd, 0)
	h.Key(key.NameReturn, 0)
	if v.Value() != "79" {
		t.Fatalf("last item unreachable %q", v.Value())
	}
	if !shown(h, "Workspace") || !shown(h, "Profile") || !shown(h, "收起侧栏") {
		t.Fatal("scroll moved fixed slots offscreen")
	}
	if b := bounds(h, "Item 79"); b.Empty() || b.Max.Y > bounds(h, "Profile").Min.Y {
		t.Fatalf("last item clipped %v", b)
	}
}

func TestSidebarNestedSelectionAndDisabled(t *testing.T) {
	calls := 0
	v := Sidebar().Height(280).Section("", SidebarItem{ID: "parent", Label: "Parent", Children: []SidebarItem{
		{ID: "a", Label: "A", Disabled: true}, {ID: "b", Label: "B"},
	}}, SidebarItem{ID: "other", Label: "Other"}).OnChange(func(string) { calls++ })
	h := page(v)
	click(t, h, "Parent")
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	if !shown(h, "B") || !v.Expanded("parent") {
		t.Fatal("Right did not expand")
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if v.Value() != "b" {
		t.Fatal("did not skip disabled child")
	}
	h.Key(key.NameLeftArrow, 0)
	h.Key(key.NameLeftArrow, 0)
	h.Frame()
	if v.Expanded("parent") || shown(h, "B") {
		t.Fatal("Left did not collapse parent")
	}
	before := calls
	v.SetValue("b")
	h.Frame()
	h.Frame()
	if !v.Expanded("parent") || !shown(h, "B") || calls != before {
		t.Fatal("program selection did not reveal ancestor")
	}
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "Other")
	if v.Value() != "b" || calls != before {
		t.Fatal("disabled sidebar")
	}
}

func TestSidebarCopiesNestedItemsAndRejectsDuplicateIDs(t *testing.T) {
	children := []SidebarItem{{ID: "child", Label: "Child"}}
	v := Sidebar().Section("", SidebarItem{ID: "parent", Label: "Parent", Children: children})
	children[0].Label = "changed"
	if v.find("child").Label != "Child" {
		t.Fatal("nested alias")
	}
	v.Section("bad", SidebarItem{ID: "child", Label: "Duplicate"})
	if len(v.sections) != 1 {
		t.Fatal("duplicate ID accepted")
	}
	v.SetBadge("child", 9)
	v.SetItemDisabled("child", true)
	if it := v.find("child"); it.Badge != 9 || !it.Disabled {
		t.Fatal("nested updates")
	}
}

func TestSidebarCollapsedSlotsAndProgramReveal(t *testing.T) {
	v := Sidebar().Height(250)
	v.Header(el.ViewFunc(func(*el.Context) el.Element {
		if v.Collapsed() {
			return el.Text("K")
		}
		return el.Text("Workspace")
	}))
	var items []SidebarItem
	for i := range 40 {
		items = append(items, SidebarItem{ID: fmt.Sprint(i), Label: fmt.Sprintf("Item %d", i)})
	}
	v.Section("", items...)
	v.SetValue("39")
	v.SetCollapsed(true)
	h := page(v)
	h.Frame()
	h.Frame()
	if !shown(h, "K") || !shown(h, "Item 39") || !shown(h, "展开侧栏") {
		t.Fatal("collapsed selected item not revealed")
	}
}
