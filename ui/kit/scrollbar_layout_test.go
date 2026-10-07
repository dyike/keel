package kit

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestScrollableComponentRowsClearScrollbar(t *testing.T) {
	for _, kind := range []string{"select", "combobox", "menu", "list", "tree"} {
		for _, scale := range []int{1, 2} {
			t.Run(fmt.Sprintf("%s/scale%d", kind, scale), func(t *testing.T) {
				labels := make([]string, 60)
				for i := range labels {
					labels[i] = fmt.Sprintf("Row %d", i)
				}
				var view el.View
				var scrollID string
				calls := 0
				switch kind {
				case "select":
					v := Select("Choices", labels...).OnChange(func(string) { calls++ })
					scrollID = v.virtual.ID()
					view = viewFunc(func(cx *el.Context) el.Element { return v.list(cx, "choices") })
				case "combobox":
					v := Combobox("Choices", labels...).OnChange(func(string) { calls++ })
					scrollID = v.virtual.ID()
					view = viewFunc(func(cx *el.Context) el.Element { return v.suggestions(cx, "choices") })
				case "menu":
					v := Menu()
					for _, label := range labels {
						v.Item(label, "mod+t", func() { calls++ })
					}
					scrollID = autoID("menu-scroll", v)
					view = viewFunc(v.panel)
				case "list":
					v := List(labels...).OnChange(func(int) { calls++ })
					scrollID = v.list.ID()
					view = v
				case "tree":
					nodes := make([]*TreeNode, len(labels))
					for i, label := range labels {
						nodes[i] = &TreeNode{ID: label, Label: label}
					}
					v := Tree(nodes...).OnChange(func(string) { calls++ })
					scrollID = v.list.ID()
					view = v
				}
				var cx *el.Context
				h := renderView(viewFunc(func(c *el.Context) el.Element {
					cx = c
					return el.Div().Role("scroll-test").Items(el.Stretch).Child(view.Render(c))
				}), 320, scale)
				settle(h)
				container, ok := semanticNode(h, "scroll-test")
				row := bounds(h, "Row 0")
				if !ok || row.Empty() || row.Max.X > container.Desc.Bounds.Max.X-scrollbarGutter*scale {
					t.Fatalf("row %v overlaps scrollbar in %v", row, container.Desc.Bounds)
				}
				_, viewport, _ := cx.ScrollState(scrollID)
				h.Click(float32(row.Max.X+(scrollbarGutter-5)*scale), float32(row.Min.Y)+viewport*.8*float32(scale))
				settle(h)
				if calls != 0 {
					t.Fatal("scrollbar click selected a row")
				}
				if offset, _, _ := cx.ScrollState(scrollID); offset <= 0 {
					t.Fatal("track did not scroll")
				}
			})
		}
	}
}
