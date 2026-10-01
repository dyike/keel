package kit

import (
	"gioui.org/io/input"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func renderView(v el.View, width, scale int) *uitest.Harness {
	root := el.Embed(v)
	return uitest.NewFunc(func(gtx core.C) {
		gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
		gtx.Constraints.Max = image.Pt(width*scale, 1000*scale)
		root.Layout(gtx)
	})
}
func semanticNode(h *uitest.Harness, role string) (input.SemanticNode, bool) {
	var found input.SemanticNode
	ok := false
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Description == role {
			found = n
			ok = true
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	return found, ok
}
func TestAlertWrapsAndRetainsStatus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		a := Alert("保存 123", "中英文 Mixed description that wraps in a narrow container. 多行提示内容。").Tone(Success)
		h := renderView(a, 160, scale)
		for i := 0; i < 3; i++ {
			h.Frame()
		}
		n, ok := semanticNode(h, "alert:success")
		if !ok || n.Desc.Label != "保存 123" || n.Desc.Bounds.Dx() > 160*scale || n.Desc.Bounds.Dy() <= 40*scale {
			t.Fatalf("invalid alert: %+v", n)
		}
		a.SetTitle("更新")
		a.SetDescription("")
		a.Tone(Warning)
		h.Frame()
		if n, ok = semanticNode(h, "alert:warning"); !ok || n.Desc.Label != "更新" {
			t.Fatal("updated alert not rendered")
		}
	}
}
