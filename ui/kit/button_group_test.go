package kit

import (
	"slices"
	"testing"
)

// A selected button reports it to agents; the group keeps each button's own
// click, and disabling the group disables them all.
func TestButtonSelectedAndGroup(t *testing.T) {
	var clicks []string
	on := func(s string) func() { return func() { clicks = append(clicks, s) } }
	left := Button("左对齐", on("left")).Variant(ButtonSecondary).Selected(true)
	center := Button("居中", on("center")).Variant(ButtonSecondary)
	right := Button("右对齐", on("right")).Variant(ButtonSecondary)
	g := ButtonGroup(left, center).Add(right).Name("对齐方式")
	h := page(g)
	h.Frame()
	if n, ok := semanticNode(h, "group"); !ok || n.Desc.Label != "对齐方式" {
		t.Fatalf("group semantics %+v", n.Desc)
	}
	if n, _ := node(h, "左对齐"); !n.Desc.Selected {
		t.Fatal("selected button not reported as selected")
	}
	if n, _ := node(h, "居中"); n.Desc.Selected {
		t.Fatal("unselected button reported as selected")
	}
	click(t, h, "居中")
	h.Frame()
	if !slices.Equal(clicks, []string{"center"}) {
		t.Fatalf("clicks %v", clicks)
	}
	left.SetSelected(false)
	center.SetSelected(true)
	h.Frame()
	if n, _ := node(h, "居中"); !n.Desc.Selected || left.IsSelected() {
		t.Fatal("selection did not move")
	}
	// Buttons sit side by side, touching: the group is one control.
	a, b := bounds(h, "左对齐"), bounds(h, "居中")
	if b.Min.X-a.Max.X > 2 || a.Min.Y != b.Min.Y {
		t.Fatalf("buttons not joined in a row: %v %v", a, b)
	}
	g.SetDisabled(true)
	h.Frame()
	click(t, h, "右对齐")
	if len(clicks) != 1 {
		t.Fatal("a disabled group's button ran its click")
	}
	g.SetDisabled(false)
	g.Vertical(true)
	h.Frame()
	a, b = bounds(h, "左对齐"), bounds(h, "居中")
	if b.Min.Y-a.Max.Y > 2 || a.Min.X != b.Min.X {
		t.Fatalf("vertical group not stacked: %v %v", a, b)
	}
}
