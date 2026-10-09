package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestSheetPanelStylePreservesStateAndLayout(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	for _, scale := range []int{1, 2} {
		input := Input("Editor")
		calls := 0
		s := Sheet(el.Right, "Panel").Size(220).MarginTop(32).Body(input).Footer(Button("Apply", func() { calls++ }))
		s.SetValue(true)
		h := renderView(s, 400, scale)
		original := bounds(h, "Editor")
		clickClass(t, h, "Editor", "Editor")
		h.Type("a")
		s.PanelStyle(func(e *el.DivEl) { e.P(8).Gap(4).W(el.Dp(9999)).H(el.Dp(9999)).Role("wrong").Name("wrong") })
		h.Frame()
		n, ok := semanticNode(h, "dialog")
		if !ok || n.Desc.Label != "Panel" || n.Desc.Bounds.Dx() != 220*scale || n.Desc.Bounds.Min.Y != 32*scale {
			t.Fatal("style replaced sizing or identity", n.Desc)
		}
		if bounds(h, "Editor").Min.X >= original.Min.X {
			t.Fatal("padding not applied")
		}
		h.Key(key.NameRightArrow, 0)
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "" {
			t.Fatal("style lost input focus", input.Value())
		}
		s.PanelStyle(nil)
		h.Frame()
		h.Type("b")
		if input.Value() != "b" || bounds(h, "Editor") != original {
			t.Fatal("reset lost state or default layout")
		}
		click(t, h, "Apply")
		if calls != 1 {
			t.Fatal("footer interaction")
		}
	}
}
