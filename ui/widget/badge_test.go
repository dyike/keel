package widget

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestBadgeVisibilityAndChildInteraction(t *testing.T) {
	clicks := 0
	child := Button("通知", func() { clicks++ })
	badge := Badge(150).Child(child)
	h := uitest.New(badge)
	clickNamed(t, h, "通知")
	if clicks != 1 {
		t.Fatal("badge blocked child")
	}
	if badge.label() != "99+" {
		t.Fatal("count overflow not formatted")
	}
	badge.Max(9)
	if badge.label() != "9+" {
		t.Fatal("custom max ignored")
	}
	badge.SetCount(0)
	h.Frame()
	if !namedBounds(h, "150").Empty() {
		t.Fatal("zero badge still exposed")
	}
	clickNamed(t, h, "通知")
	if clicks != 2 {
		t.Fatal("hidden badge blocked child")
	}
	var plain, hidden image.Point
	uitest.NewFunc(func(gtx core.C) { plain = child.Layout(gtx).Size })
	uitest.NewFunc(func(gtx core.C) { hidden = badge.Layout(gtx).Size })
	if hidden != plain {
		t.Fatalf("hidden badge reserves space: %v != %v", hidden, plain)
	}
	empty := Badge(-1)
	var dims core.D
	uitest.NewFunc(func(gtx core.C) { dims = empty.Layout(gtx) })
	if dims.Size != (image.Point{}) {
		t.Fatal("negative count should be hidden")
	}
	dot := Badge(1).Dot()
	uitest.NewFunc(func(gtx core.C) { dims = dot.Layout(gtx) })
	if dims.Size != image.Pt(8, 8) {
		t.Fatalf("unexpected dot size %v", dims.Size)
	}
	// Overlay must stay within the allocated bounds, including very narrow parents.
	for _, width := range []int{0, 4, 20, 100} {
		uitest.NewFunc(func(gtx core.C) {
			gtx.Constraints.Max = image.Pt(width, 40)
			dims = Badge(150).Child(Text("x")).Layout(gtx)
		})
		if dims.Size.X > width || dims.Size.Y > 40 {
			t.Fatalf("badge exceeded constraints: %v", dims.Size)
		}
	}
}
