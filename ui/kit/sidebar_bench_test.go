package kit

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/el"
)

// Exercise real hover suffixes and per-item context menus, as in project/task
// sidebars. Closed descendants stay in the model but must not be painted.
func sidebarScrollFixture(rows, children int) *SidebarView {
	item := func(id string) SidebarItem {
		return SidebarItem{ID: id, Label: "任务 " + id, Icon: IconFolder,
			Suffix: Button("", func() {}).Icon(IconMore).Variant(ButtonGhost).Size(24), SuffixOnHover: true,
			ContextMenu: Menu().Item("Rename", "", func() {})}
	}
	items := make([]SidebarItem, rows)
	for i := range items {
		items[i] = item(fmt.Sprint(i))
		for j := range children {
			items[i].Children = append(items[i].Children, item(fmt.Sprintf("%d/%d", i, j)))
		}
	}
	return Sidebar().Width(260).Height(600).Collapsible(false).Section("项目", items...)
}

func BenchmarkSidebarScroll(b *testing.B) {
	for _, tc := range []struct {
		name           string
		rows, children int
	}{
		{"100-rows", 100, 0}, {"1000-rows", 1000, 0}, {"100-projects-20-closed-children", 100, 20},
	} {
		b.Run(tc.name, func(b *testing.B) {
			v := sidebarScrollFixture(tc.rows, tc.children)
			root := el.Root(v)
			var router input.Router
			var ops op.Ops
			frame := func() {
				ops.Reset()
				root.Layout(layout.Context{Ops: &ops, Now: time.Now(), Source: router.Source(), Constraints: layout.Exact(image.Pt(800, 1200)), Metric: unit.Metric{PxPerDp: 2, PxPerSp: 2}})
				router.Frame(&ops)
			}
			for range 3 {
				frame()
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := range b.N {
				dy := float32(48)
				if i/80%2 != 0 {
					dy = -dy
				}
				router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(200, 600), Scroll: f32.Pt(0, dy)})
				frame()
			}
		})
	}
}
