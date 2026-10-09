package kit

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"math"
	"testing"
)

func sidebarContextClick(t *testing.T, h *uitest.Harness, name string) {
	t.Helper()
	b := bounds(h, name)
	if b.Empty() {
		t.Fatal("missing item", name)
	}
	p := f32.Pt(float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2))
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: p, Buttons: pointer.ButtonSecondary}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
	h.Frame()
	h.Frame()
}

func TestSidebarSuffixIndependentAndDisabled(t *testing.T) {
	calls, selected := 0, 0
	v := Sidebar().Section("", SidebarItem{ID: "parent", Label: "Parent", Children: []SidebarItem{{ID: "child", Label: "Child"}}, Suffix: Button("Refresh", func() { calls++ })}, SidebarItem{ID: "other", Label: "Other"}).OnChange(func(string) { selected++ })
	v.SetValue("other")
	h := renderView(v, 300, 2)
	click(t, h, "Refresh")
	if calls != 1 || selected != 0 || v.Value() != "other" || v.Expanded("parent") {
		t.Fatal("suffix activated navigation", calls, selected, v.Value())
	}
	v.SetItemDisabled("parent", true)
	h.Frame()
	click(t, h, "Refresh")
	if calls != 1 {
		t.Fatal("disabled suffix activated")
	}
	v.SetItemDisabled("parent", false)
	v.SetCollapsed(true)
	h.Frame()
	if shown(h, "Refresh") {
		t.Fatal("collapsed suffix visible")
	}
	v.SetCollapsed(false)
	v.SetSuffix("parent", nil)
	h.Frame()
	if shown(h, "Refresh") {
		t.Fatal("removed suffix visible")
	}
}

func TestSidebarCustomIconUsesStandardSlotAndSelection(t *testing.T) {
	for _, scale := range []int{1, 2} {
		icon := el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().Role("image").Name("Agent logo").Size(el.Dp(16))
		})
		v := Sidebar().Section("", SidebarItem{ID: "agent", Label: "Agent", Icon: IconFolder, IconView: icon})
		h := renderView(v, 300, scale)
		logo := bounds(h, "Agent logo")
		if logo.Dx() != 16*scale || logo.Dy() != 16*scale {
			t.Fatalf("custom icon did not use standard slot: %v", logo)
		}
		click(t, h, "Agent")
		if v.Value() != "agent" {
			t.Fatal("custom icon blocked row activation")
		}
		v.SetCollapsed(true)
		h.Frame()
		h.Frame()
		if !shown(h, "Agent logo") {
			t.Fatal("collapsed sidebar lost custom icon")
		}
		v.SetDisabled(true)
		v.SetValue("")
		h.Frame()
		click(t, h, "Agent")
		if v.Value() != "" {
			t.Fatal("custom icon bypassed disabled navigation")
		}
	}
}

func TestSidebarItemMenuLifecycle(t *testing.T) {
	calls := 0
	menu := Menu().Item("Run item", "", func() { calls++ })
	v := Sidebar().Section("", SidebarItem{ID: "a", Label: "Alpha", ContextMenu: menu}, SidebarItem{ID: "b", Label: "Beta"})
	v.SetValue("b")
	h := renderView(v, 400, 1)
	sidebarContextClick(t, h, "Alpha")
	if !menu.open || !shown(h, "Run item") || v.Value() != "b" {
		t.Fatal("right-click failed or changed selection")
	}
	click(t, h, "Run item")
	if calls != 1 || menu.open {
		t.Fatal("menu command not delivered or closed")
	}
	sidebarContextClick(t, h, "Alpha")
	v.Filter("Beta")
	h.Frame()
	if menu.open || shown(h, "Run item") {
		t.Fatal("filtered item retained menu")
	}
	v.Filter("")
	h.Frame()
	sidebarContextClick(t, h, "Alpha")
	v.SetItemDisabled("a", true)
	h.Frame()
	if menu.open || shown(h, "Run item") {
		t.Fatal("disabled item retained menu")
	}
	v.SetItemDisabled("a", false)
	h.Frame()
	sidebarContextClick(t, h, "Alpha")
	v.SetContextMenu("a", Menu().Item("Replacement", "", nil))
	h.Frame()
	if menu.open || shown(h, "Run item") {
		t.Fatal("replaced menu retained")
	}
	sidebarContextClick(t, h, "Alpha")
	if !shown(h, "Replacement") {
		t.Fatal("replacement unavailable")
	}
}

func TestSidebarRightEdgeAndToggleVisibility(t *testing.T) {
	v := Sidebar().Width(220).BorderWidth(3).Section("", SidebarItem{ID: "a", Label: "Alpha"})
	h := renderView(v, 300, 2)
	left := bounds(h, "Alpha")
	v.Side(el.Right)
	h.Frame()
	right := bounds(h, "Alpha")
	if right.Min.X-left.Min.X != 6 || right.Size() != left.Size() {
		t.Fatal("right divider placement", left, right)
	}
	v.BorderWidth(float32(math.NaN())).BorderWidth(-1).Side(el.Top)
	h.Frame()
	if bounds(h, "Alpha") != right {
		t.Fatal("invalid presentation changed layout")
	}
	v.Collapsible(false)
	h.Frame()
	if shown(h, "收起侧栏") {
		t.Fatal("hidden toggle remained")
	}
	v.SetCollapsed(true)
	h.Frame()
	if !v.Collapsed() || shown(h, "展开侧栏") {
		t.Fatal("programmatic collapse requires toggle")
	}
}

func TestSidebarMenuSubmenuAndAncestorDisabled(t *testing.T) {
	calls := 0
	sub := Menu().Item("Nested command", "", func() { calls++ })
	menu := Menu().Sub("More", sub)
	v := Sidebar().Section("", SidebarItem{ID: "a", Label: "Alpha", ContextMenu: menu, Suffix: Button("Action", func() { calls++ })})
	disabled := false
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 400, 1)
	sidebarContextClick(t, h, "Alpha")
	click(t, h, "More")
	h.Frame()
	if !shown(h, "Nested command") {
		t.Fatal("submenu missing")
	}
	click(t, h, "Nested command")
	if calls != 1 || menu.open {
		t.Fatal("submenu callback or close failed")
	}
	disabled = true
	h.Frame()
	click(t, h, "Action")
	sidebarContextClick(t, h, "Alpha")
	if calls != 1 || shown(h, "More") {
		t.Fatal("ancestor disabled did not block interactions")
	}
	disabled = false
	h.Frame()
	sidebarContextClick(t, h, "Alpha")
	if !shown(h, "More") {
		t.Fatal("menu did not recover")
	}
}
