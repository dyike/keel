package kit

import "testing"

func TestDescriptionListCopiesAndWraps(t *testing.T) {
	items := []Description{{"订单 123", "SO-123 中英文 mixed long content"}, {"客户", "张三"}}
	v := DescriptionList()
	v.SetItems(items...)
	items[0].Text = "changed"
	if v.items[0].Text == "changed" {
		t.Fatal("aliased items")
	}
	for _, scale := range []int{1, 2} {
		h := renderView(v, 100, scale)
		n, ok := node(h, "订单 123：SO-123 中英文 mixed long content")
		if !ok || n.Desc.Bounds.Dx() > 100*scale || n.Desc.Bounds.Dy() <= 20*scale {
			t.Fatalf("bad list bounds: %+v", n)
		}
	}
	v.SetItems()
	if len(v.items) != 0 {
		t.Fatal("items not cleared")
	}
}

func TestDescriptionColumns(t *testing.T) {
	v := DescriptionList().Item("短", "值").Item("较长标签", "长值").LabelWidth(80)
	h := renderView(v, 300, 1)
	a := bounds(h, "值")
	b := bounds(h, "长值")
	if a.Min.X != b.Min.X || a.Min.X != 92 {
		t.Fatalf("value columns differ: %v %v", a, b)
	}
}
