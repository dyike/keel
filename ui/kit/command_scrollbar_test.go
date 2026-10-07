package kit

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestCommandScrollbarGutter(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, autoRows := range []bool{false, true} {
			t.Run(fmt.Sprintf("scale%d/autoRows%v", scale, autoRows), func(t *testing.T) {
				ran := 0
				items := make([]CommandItem, 30)
				for i := range items {
					items[i] = CommandItem{Title: fmt.Sprintf("Row %d", i), Shortcut: "mod+t", Action: func() { ran++ }}
				}
				v := Command(items...).Inline(true).AutoRowHeight(autoRows).MaxHeight(200)
				var cx *el.Context
				h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element {
					cx = ctx
					return v.Render(ctx)
				}), 400, scale)
				settle(h)
				list, ok := semanticNode(h, "listbox")
				if !ok {
					t.Fatal("missing command list")
				}
				row := bounds(h, "Row 0")
				// Reserve the 10dp track plus 2dp clearance from the row's
				// background and hit target, including the selected row.
				if row.Empty() || row.Max.X > list.Desc.Bounds.Max.X-12*scale {
					t.Fatalf("row %v overlaps scrollbar gutter in list %v", row, list.Desc.Bounds)
				}
				click(t, h, "Row 0")
				if ran != 1 {
					t.Fatal("row click failed")
				}
				h.Click(float32(list.Desc.Bounds.Max.X-5*scale), float32(list.Desc.Bounds.Min.Y+150*scale))
				settle(h)
				if ran != 1 {
					t.Fatal("scrollbar click activated a command")
				}
				if offset, _, _ := cx.ScrollState(v.listID()); offset <= 0 {
					t.Fatal("scrollbar track did not scroll")
				}
			})
		}
	}
}
