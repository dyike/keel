package kit

import (
	"fmt"
	"image"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestTableHorizontalScroll(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			var cx *el.Context
			built := 0
			cols := []*ColumnSpec{Col("first").Width(200), Col("second").Width(200), Col("third").Width(200)}
			cols[2].Cell(func(cx *el.Context, row int) el.Element {
				built++
				return el.Text(fmt.Sprint(row)).Name(fmt.Sprintf("third-%d", row))
			})
			table := Table(cols...).Height(120)
			rows := make([][]string, 10000)
			for i := range rows {
				rows[i] = []string{fmt.Sprint(i), "middle", "last"}
			}
			table.SetRows(rows)
			width := 300
			root := el.Root(viewFunc(func(c *el.Context) el.Element {
				cx = c
				return el.Div().Child(table.Render(c))
			}))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints.Max = image.Pt(width*scale, 400*scale)
				root.Layout(gtx)
			})
			id := autoID("table", table)
			x, view, content := cx.ScrollStateX(id)
			if x != 0 || view != 298 || content != 600 {
				t.Fatalf("initial: %g %g %g", x, view, content)
			}
			before := bounds(h, "second").Min.X
			h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(50*float32(scale), 70*float32(scale)), Scroll: f32.Pt(300*float32(scale), 0)})
			h.Frame()
			h.Frame()
			x, _, _ = cx.ScrollStateX(id)
			after := bounds(h, "third")
			if x != 300 || before-bounds(h, "second").Min.X != 300*scale {
				t.Fatalf("header shift: %g %v to %v", x, before, after)
			}
			cell := bounds(h, "third-0")
			if cell.Min.X != after.Min.X+12*scale {
				t.Fatalf("header/cell mismatch: %v %v", after, cell)
			}
			if built >= 100 {
				t.Fatalf("virtualization built %d cells", built)
			}
			h.Click(float32(cell.Min.X+3), float32(cell.Min.Y+3))
			if table.Value() != 0 {
				t.Fatal("scrolled row cannot be selected")
			}
			table.SetValue(9999)
			h.Frame()
			h.Frame()
			off, _, _ := cx.ScrollState(table.list.ID())
			if off <= 0 {
				t.Fatal("vertical selection reveal lost")
			}
			if x, _, _ := cx.ScrollStateX(id); x != 300 {
				t.Fatal("vertical reveal moved horizontal axis")
			}
			width = 900
			h.Frame()
			h.Frame()
			if x, view, content := cx.ScrollStateX(id); x != 0 || view != 898 || content != 898 {
				t.Fatalf("resize: %g %g %g", x, view, content)
			}
		})
	}
}

func TestTableFlexibleColumnsFillAndOverflow(t *testing.T) {
	var cx *el.Context
	fixed := Col("fixed").Width(160)
	table := Table(fixed, Col("flex-a"), Col("flex-b").Flex(2)).Height(120)
	table.SetRows([][]string{{"a", "b", "c"}})
	width := float32(500)
	h := render(func(c *el.Context) el.Element {
		cx = c
		return el.Div().W(el.Dp(width)).Child(table.Render(c))
	})
	if a, b := table.widths[1], table.widths[2]; a < 100 || b < 200 || b < 1.9*a {
		t.Fatalf("flex widths %v", table.widths)
	}
	width = 200
	h.Frame()
	h.Frame()
	if _, view, content := cx.ScrollStateX(autoID("table", table)); view != 198 || content != 240 {
		t.Fatalf("narrow: %g %g", view, content)
	}
	cx.ScrollIntoViewX(autoID("table", table), 200, 240)
	h.Frame()
	h.Frame()
	if table.widths[0] != 160 || table.widths[1] != 40 || table.widths[2] != 40 {
		t.Fatalf("minimum widths %v", table.widths)
	}
	table.SetColumnWidth(0, 350)
	h.Frame()
	h.Frame()
	if _, _, content := cx.ScrollStateX(autoID("table", table)); content != 430 {
		t.Fatalf("column resize: %g", content)
	}
}
