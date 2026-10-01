package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestGroupBoxRetainsChildren(t *testing.T) {
	child := Tag("原文")
	children := []el.View{child, nil}
	v := GroupBox("分组 123", children...)
	children[0] = nil
	for _, scale := range []int{1, 2} {
		h := renderView(v, 160, scale)
		n, ok := semanticNode(h, "group")
		if !ok || n.Desc.Bounds.Dx() > 160*scale {
			t.Fatal("group exceeded bounds")
		}
		child.SetText("更新")
		h.Frame()
		if n, ok = semanticNode(h, "tag:neutral"); !ok || n.Desc.Label != "更新" {
			t.Fatal("child was not retained")
		}
	}
	v.SetChildren()
	if len(v.children) != 0 {
		t.Fatal("children not cleared")
	}
}
