package kit

import "testing"

func TestDescriptionListCopiesAndWraps(t *testing.T) {
	items := []Description{{"订单 123", "SO-123 中英文 mixed long content"}, {"客户", "张三"}}
	v := DescriptionList(items...)
	items[0].Text = "changed"
	if v.items[0].Text == "changed" {
		t.Fatal("aliased items")
	}
	for _, scale := range []int{1, 2} {
		h := renderView(v, 100, scale)
		n, ok := semanticNode(h, "descriptionlist:2")
		if !ok || n.Desc.Bounds.Dx() > 100*scale || n.Desc.Bounds.Dy() <= 50*scale {
			t.Fatalf("bad list bounds: %+v", n)
		}
	}
	v.SetItems()
	if len(v.items) != 0 {
		t.Fatal("items not cleared")
	}
}
