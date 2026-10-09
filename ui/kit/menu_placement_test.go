package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"math"
	"testing"
)

func TestDropdownMenuPlacementAndOffset(t *testing.T) {
	m := Menu().Width(60).Item("One", "", nil)
	d := DropdownButton("Target", m)
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(100).Items(el.Start).Child(d.Render(cx)) })))
	click(t, h, "Target")
	for _, gap := range []float32{4, 12, 0, -4} {
		d.Offset(gap)
		for _, side := range []el.Side{el.Top, el.Bottom, el.Left, el.Right} {
			for _, align := range []el.Align{el.Start, el.Center, el.End} {
				d.Placement(side, align)
				h.Frame()
				n, ok := semanticNode(h, "menu")
				if !ok {
					t.Fatal("menu disappeared")
				}
				a, b := bounds(h, "Target"), n.Desc.Bounds
				distance := 0
				switch side {
				case el.Top:
					distance = a.Min.Y - b.Max.Y
				case el.Bottom:
					distance = b.Min.Y - a.Max.Y
				case el.Left:
					distance = a.Min.X - b.Max.X
				case el.Right:
					distance = b.Min.X - a.Max.X
				}
				if distance != int(gap) {
					t.Fatal("anchor gap", side, align, gap, a, b)
				}
				low, high, start, end := a.Min.X, a.Max.X, b.Min.X, b.Max.X
				if side == el.Left || side == el.Right {
					low, high, start, end = a.Min.Y, a.Max.Y, b.Min.Y, b.Max.Y
				}
				switch align {
				case el.Start:
					if start != low {
						t.Fatal("start alignment")
					}
				case el.End:
					if end != high {
						t.Fatal("end alignment")
					}
				case el.Center:
					if delta := (start + end) - (low + high); delta < -1 || delta > 1 {
						t.Fatal("center alignment")
					}
				}
			}
		}
	}
	d.Offset(float32(math.NaN())).Placement(el.Side(255), el.Start)
	if m.offset != -4 || m.side != el.Right || m.align != el.End {
		t.Fatal("invalid config accepted")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Frame()
	if m.Value() {
		t.Fatal("Esc failed")
	}
	h.Key(key.NameSpace, 0)
	h.Frame()
	if !m.Value() {
		t.Fatal("focus not restored to trigger")
	}
}

func TestSplitDropdownPlacementFlipsAtEdge(t *testing.T) {
	m := Menu().Item("Choice", "", nil)
	d := DropdownButton("Save", m).Split(func() {}).Placement(el.Bottom, el.End).Offset(8)
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Pt(230).Items(el.Start).Child(d.Render(cx)) })))
	m.SetValue(true)
	h.Frame()
	h.Frame()
	n, ok := semanticNode(h, "menu")
	if !ok {
		t.Fatal("menu absent")
	}
	a := bounds(h, "Save 更多选项")
	if n.Desc.Bounds.Max.Y != a.Min.Y-8 || n.Desc.Bounds.Min.X < 0 || n.Desc.Bounds.Max.X > 400 {
		t.Fatal("edge flip/clamp", a, n.Desc.Bounds)
	}
}
