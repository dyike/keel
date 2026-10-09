package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestGroupBoxRetainsChildren(t *testing.T) {
	child := Tag("原文")
	children := []el.View{child, nil}
	v := GroupBox("分组 123").Child(children...)
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
	v := GroupBox("设置").Description("通知说明").Child(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("内容") }))
	h := renderView(v, 220, 1)
	if _, ok := node(h, "通知说明"); !ok {
		t.Fatal("missing description")
	}
	if bounds(h, "内容").Min.Y <= bounds(h, "通知说明").Max.Y {
		t.Fatal("content overlaps description")
	}
}

func TestGroupBoxVariantsStylesFooterAndState(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		input := Input("Editor")
		v := GroupBox("Settings").Description("Details").Child(input).Footer(Button("Apply", func() { calls++ }))
		h := renderView(v, 220, scale)
		clickClass(t, h, "Editor", "Editor")
		h.Type("a")
		for _, variant := range []GroupBoxVariant{GroupBoxNormal, GroupBoxFill, GroupBoxOutline, GroupBoxSurface} {
			v.Variant(variant).TitleStyle(func(e *el.TextEl) { e.TextSize(theme.TextHeading).TextColor(theme.PrimaryText) }).ContentStyle(func(e *el.DivEl) { e.P(20).Gap(8) })
			h.Frame()
			if input.Value() != "a" {
				t.Fatal("variant reset input")
			}
			if bounds(h, "Apply").Min.Y <= bounds(h, "Editor").Max.Y {
				t.Fatal("footer overlaps body")
			}
			n, ok := semanticNode(h, "group")
			if !ok || n.Desc.Bounds.Dx() > 220*scale {
				t.Fatal("group overflow")
			}
		}
		h.Key(key.NameRightArrow, 0)
		v.SetTitle("")
		v.Description("")
		h.Frame()
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "" {
			t.Fatal("title removal lost caret or focus", input.Value())
		}
		h.Type("b")
		if input.Value() != "b" {
			t.Fatal("title removal lost body focus", input.Value())
		}
		v.TitleStyle(nil).ContentStyle(nil)
		h.Frame()
		click(t, h, "Apply")
		h.Key(key.NameSpace, 0)
		if calls != 2 {
			t.Fatal("footer activation")
		}
		v.Footer(nil)
		h.Frame()
		if shown(h, "Apply") || input.Value() != "b" {
			t.Fatal("footer reset")
		}
	}
}
