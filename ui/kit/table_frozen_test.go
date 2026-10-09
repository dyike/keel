package kit

import (
	"fmt"
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestTableFrozenColumns(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			var cx *el.Context
			hits := [4]int{}
			built := 0
			columns := make([]*ColumnSpec, 4)
			for c := range columns {
				columns[c] = Col(fmt.Sprintf("head-%d", c)).Width(100).Cell(func(cx *el.Context, row int) el.Element {
					built++
					return Button(fmt.Sprintf("cell-%d-%d", c, row), func() { hits[c]++ }).Render(cx)
				})
			}
			table := Table(columns...).FrozenColumns(1, 1).Height(120)
			rows := make([][]string, 10000)
			for i := range rows {
				rows[i] = []string{fmt.Sprint(i), "a", "b", "z"}
			}
			table.SetRows(rows)
			width := 320
			root := el.Root(viewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().Child(table.Render(c)) }))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints.Max = image.Pt(width*scale, 400*scale)
				root.Layout(gtx)
			})
			left, right := bounds(h, "head-0"), bounds(h, "head-3")
			if left.Min.X != scale || right.Max.X != 319*scale {
				t.Fatalf("edges: %v %v", left, right)
			}
			h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(155*float32(scale), 70*float32(scale)), Scroll: f32.Pt(200*float32(scale), 0)})
			h.Frame()
			h.Frame()
			offset, _, _ := cx.ScrollStateX(autoID("table", table))
			if offset != 82 {
				t.Fatalf("offset %v", offset)
			}
			if bounds(h, "head-0") != left || bounds(h, "head-3") != right {
				t.Fatal("frozen headers moved")
			}
			click(t, h, "cell-0-0")
			click(t, h, "cell-3-0")
			if hits != [4]int{1, 0, 0, 1} {
				t.Fatalf("covered content received click: %v", hits)
			}
			built = 0
			h.Frame()
			if built > 60 {
				t.Fatalf("built %d cells", built)
			}
			table.SetValue(9999)
			h.Frame()
			h.Frame()
			if y, _, _ := cx.ScrollState(table.list.ID()); y <= 0 {
				t.Fatal("vertical reveal")
			}
			if x, _, _ := cx.ScrollStateX(autoID("table", table)); x != offset {
				t.Fatal("vertical reveal moved horizontal axis")
			}
			table.SetColumnWidth(0, 120)
			h.Frame()
			h.Frame()
			if bounds(h, "head-0").Dx() != 120*scale || bounds(h, "head-3").Max.X != 319*scale {
				t.Fatal("resize changed frozen edge")
			}
			width = 180
			h.Frame()
			h.Frame()
			// Left columns take precedence if the pinned widths exceed the viewport.
			if bounds(h, "head-0").Min.X != scale {
				t.Fatal("narrow left edge")
			}
			table.FrozenColumns(0, 0)
			width = 600
			h.Frame()
			h.Frame()
			if x, _, _ := cx.ScrollStateX(autoID("table", table)); x != 0 {
				t.Fatal("wide viewport did not reset scroll")
			}
		})
	}
}
