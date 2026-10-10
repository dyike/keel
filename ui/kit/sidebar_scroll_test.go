package kit

import (
	"fmt"
	"image"
	"testing"

	"gioui.org/io/key"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestSidebarLongListPaintsOnlyVisibleRows(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			painted := 0
			icon := el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Size(el.Dp(16)).Decorate(func(gtx core.C, draw func()) {
					if gtx.Enabled() {
						painted++
					}
					draw()
				})
			})
			v := sidebarScrollFixture(1000, 0).Height(280).
				Header(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Header") })).
				Footer(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Footer") }))
			for i := range v.sections[0].items {
				v.sections[0].items[i].IconView = icon
			}
			root := el.Root(v)
			h := uitest.NewFunc(func(gtx core.C) {
				painted = 0
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints.Max = image.Pt(400*scale, 300*scale)
				root.Layout(gtx)
			})
			check := func() {
				t.Helper()
				if painted < 1 || painted > 8 {
					t.Fatalf("painted %d of 1000 rows", painted)
				}
				if !shown(h, "Header") || !shown(h, "Footer") {
					t.Fatal("fixed slots scrolled away")
				}
			}
			h.Frame()
			check()
			click(t, h, "任务 0")
			h.Key(key.NameEnd, 0)
			h.Frame()
			h.Key(key.NameReturn, 0)
			h.Frame()
			check()
			if v.Value() != "999" || !shown(h, "任务 999") || shown(h, "任务 0") {
				t.Fatal("last row could not be revealed and activated")
			}
		})
	}
}

func TestSidebarClosesMenuInCollapsedDescendants(t *testing.T) {
	menu := Menu().Item("Child command", "", func() {})
	v := Sidebar().Section("", SidebarItem{ID: "parent", Label: "Parent", Children: []SidebarItem{
		{ID: "child", Label: "Child", ContextMenu: menu},
	}})
	v.SetExpanded("parent", true)
	h := page(v)
	sidebarContextClick(t, h, "Child")
	if !menu.Value() {
		t.Fatal("child menu did not open")
	}
	v.SetExpanded("parent", false)
	h.Frame()
	if menu.Value() || shown(h, "Child command") {
		t.Fatal("collapsed descendant retained its menu")
	}
}
