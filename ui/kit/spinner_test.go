package kit

import "testing"

func TestSpinnerSemanticsAndBounds(t *testing.T) {
	for _, scale := range []int{1, 2} {
		h := renderView(Spinner().Size(24).Label("加载 123"), 200, scale)
		n, ok := semanticNode(h, "progressbar:indeterminate")
		if !ok || n.Desc.Label != "加载 123" || n.Desc.Bounds.Dy() != 24*scale || n.Desc.Bounds.Dx() > 200*scale {
			t.Fatalf("invalid spinner %+v", n)
		}
	}
}
