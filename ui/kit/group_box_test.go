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

func TestGroupBoxDescriptionAndElements(t *testing.T) {
	v := GroupBox("设置").Description("通知说明").Child(el.Text("内容"))
	h := renderView(v, 220, 1)
	if _, ok := node(h, "通知说明"); !ok {
		t.Fatal("missing description")
	}
	if bounds(h, "内容").Min.Y <= bounds(h, "通知说明").Max.Y {
		t.Fatal("content overlaps description")
	}
}
