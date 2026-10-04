package kit

import (
	"fmt"
	"slices"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
)

func TestTableHeaderDragReordersWithoutSorting(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			v := Table(Col("A").Width(100), Col("B").Width(120), Col("C").Width(140)).Height(100)
			v.SetRows([][]string{{"z", "b", "c"}, {"a", "e", "f"}})
			v.SetValue(1)
			var calls [][3]int
			v.OnColumnMove(func(c, from, to int) { calls = append(calls, [3]int{c, from, to}) })
			h := renderView(v, 400, scale)
			a, c := bounds(h, "A"), bounds(h, "C")
			h.Drag(float32(a.Min.X+30*scale), float32(a.Min.Y+15*scale), float32(c.Max.X-15*scale), float32(c.Min.Y+15*scale))
			h.Frame()
			if !slices.Equal(v.columns, []int{1, 2, 0}) || !slices.Equal(calls, [][3]int{{0, 0, 2}}) {
				t.Fatal(v.columns, calls)
			}
			if v.sortCol != -1 || v.Value() != 1 || v.Row(0)[0] != "z" {
				t.Fatal("drag changed row state")
			}
			if bounds(h, "A").Min.X <= bounds(h, "C").Min.X {
				t.Fatal("display order unchanged")
			}
			click(t, h, "A")
			if v.sortCol != 0 {
				t.Fatal("click no longer sorts")
			}
			b := bounds(h, "B")
			before := v.cols[1].width
			h.Drag(float32(b.Max.X-2*scale), float32(b.Min.Y+15*scale), float32(b.Max.X+28*scale), float32(b.Min.Y+15*scale))
			if v.cols[1].width <= before || !slices.Equal(v.columns, []int{1, 2, 0}) || len(calls) != 1 {
				t.Fatal("resize interfered with order", v.cols[1].width, v.columns, calls)
			}
		})
	}
}

func TestTableHeaderDragLocksHiddenAndCancellation(t *testing.T) {
	for _, mode := range []string{"hidden", "locked", "disabled", "cancel", "outside", "restore"} {
		t.Run(mode, func(t *testing.T) {
			cols := []*ColumnSpec{Col("A").Width(100), Col("B").Width(100), Col("C").Width(100), Col("D").Width(100)}
			if mode == "locked" {
				cols[1].Movable(false)
			}
			v := Table(cols...).Height(80)
			if mode == "hidden" {
				v.SetColumnVisible(1, false)
			}
			h := renderView(v, 500, 1)
			a, d := bounds(h, "A"), bounds(h, "D")
			start, end := f32.Pt(float32(a.Min.X+20), float32(a.Min.Y+15)), f32.Pt(float32(d.Max.X-20), float32(d.Min.Y+15))
			h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: start})
			h.Frame()
			h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: end})
			h.Frame()
			if mode == "disabled" {
				v.SetDisabled(true)
				h.Frame()
			}
			if mode == "restore" {
				if err := v.SetLayoutState(v.LayoutState()); err != nil {
					t.Fatal(err)
				}
				h.Frame()
			}
			kind := pointer.Release
			if mode == "cancel" {
				kind = pointer.Cancel
			}
			if mode == "outside" {
				end.Y += 150
			}
			h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Mouse, Position: end})
			h.Frame()
			want := []int{0, 1, 2, 3}
			if mode == "hidden" {
				want = []int{1, 2, 3, 0}
			}
			if !slices.Equal(v.columns, want) {
				t.Fatal(v.columns, want)
			}
			if v.columnDrag != nil {
				t.Fatal("drag state stuck")
			}
		})
	}
}

func TestTableHeaderDragWithFrozenAndScrolledColumns(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			v := Table(Col("A").Width(80).Movable(false), Col("B").Width(120), Col("C").Width(120), Col("D").Width(120), Col("E").Width(80)).FrozenColumns(1, 1).Height(80)
			var cx *el.Context
			h := renderView(viewFunc(func(c *el.Context) el.Element { cx = c; return v.Render(c) }), 360, scale)
			cx.ScrollToX(autoID("table", v), 110)
			h.Frame()
			h.Frame()
			d, e := bounds(h, "D"), bounds(h, "E")
			x := float32(min(d.Max.X-10*scale, e.Min.X-10*scale))
			h.Drag(x, float32(d.Min.Y+15*scale), float32(e.Max.X-15*scale), float32(e.Min.Y+15*scale))
			h.Frame()
			if !slices.Equal(v.columns, []int{0, 1, 2, 4, 3}) {
				t.Fatal("frozen target lost to underneath header", v.columns, d, e)
			}
			if bounds(h, "D").Max.X != 359*scale {
				t.Fatal("new right frozen column not pinned", bounds(h, "D"))
			}
		})
	}
}

func TestTableHeaderTouchDragPreservesColumnSelectionAndEditor(t *testing.T) {
	v := Table(Col("A").Width(100), Col("B").Width(100).NoSort().Cell(func(cx *el.Context, row int) el.Element { return el.Input().Name("draft editor") }), Col("C").Width(100)).Height(80)
	v.SetRows([][]string{{"a", "b", "c"}})
	h := renderView(v, 400, 1)
	clickClass(t, h, "Editor", "draft editor")
	h.Type("kept")
	if desc(h, "draft editor") != "kept" {
		t.Fatal("editor did not receive text before dragging")
	}
	v.ColumnSelect()
	v.SetSelectedColumns([]int{1})
	h.Frame()
	b, a := bounds(h, "B"), bounds(h, "A")
	start, end := f32.Pt(float32(b.Min.X+30), float32(b.Min.Y+15)), f32.Pt(float32(a.Min.X+15), float32(a.Min.Y+15))
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 1, Position: start})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Touch, PointerID: 1, Position: end})
	h.Frame()
	if v.columnDrag == nil || !v.columnDrag.valid || !slices.Equal(v.columns, []int{0, 1, 2}) {
		t.Fatal("missing pending drop preview", v.columnDrag)
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Touch, PointerID: 1, Position: end})
	h.Frame()
	h.Frame()
	if !slices.Equal(v.columns, []int{1, 0, 2}) || !slices.Equal(v.SelectedColumns(), []int{1}) || desc(h, "draft editor") != "kept" {
		t.Fatal("lost source-column state", v.columns, v.SelectedColumns(), desc(h, "draft editor"))
	}
}
