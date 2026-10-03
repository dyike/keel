package kit

import (
	"github.com/dyike/keel/ui/theme"
	"math"
	"testing"
)

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

func TestDescriptionListSpansSeparatorsAndVertical(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := DescriptionList().Columns(3).Vertical().Item("A", "Alpha").Span(2).Item("B", "Beta").Separator().Item("C", "Gamma").Span(3).Item("D", "Delta")
		h := renderView(v, 324, scale)
		a, b, c, d := bounds(h, "A：Alpha"), bounds(h, "B：Beta"), bounds(h, "C：Gamma"), bounds(h, "D：Delta")
		if a.Dx() != 212*scale || b.Min.X != 224*scale || c.Dx() != 324*scale || c.Min.Y <= a.Max.Y || d.Min.Y <= c.Max.Y {
			t.Fatalf("grid: %v %v %v %v", a, b, c, d)
		}
		if bounds(h, "Alpha").Min.Y <= bounds(h, "A").Max.Y {
			t.Fatal("value not below label")
		}
		separator, ok := semanticNode(h, "separator")
		if !ok || separator.Desc.Bounds.Dx() != 324*scale {
			t.Fatal("separator not full width")
		}
		v.Columns(1)
		h.Frame()
		if bounds(h, "B：Beta").Min.Y <= bounds(h, "A：Alpha").Max.Y {
			t.Fatal("column update did not reflow")
		}
	}
}

func TestDescriptionListRichSpanAndNarrowLayout(t *testing.T) {
	calls := 0
	v := DescriptionList().Columns(2).Vertical().Bordered(true).Size(theme.TextSm).
		Item("姓名", "张三 Ada").Item("状态", "已发货").
		ItemView("操作", Button("查看", func() { calls++ })).Span(2)
	for _, scale := range []int{1, 2} {
		h := renderView(v, 160, scale)
		for _, name := range []string{"姓名：张三 Ada", "状态：已发货", "查看"} {
			b := bounds(h, name)
			if b.Empty() || b.Min.X < 0 || b.Max.X > 160*scale {
				t.Fatalf("narrow layout: %s %v", name, b)
			}
		}
		click(t, h, "查看")
	}
	if calls != 2 {
		t.Fatal("rich value lost interaction")
	}
	v.LabelWidth(float32(math.Inf(1))).Size(float32(math.NaN())).Columns(0)
	if v.labelWidth != 96 || v.size != theme.TextSm || v.columns != 1 {
		t.Fatal("invalid dimensions changed settings")
	}
	v.SetItems(Description{"Reset", "Value"})
	h := renderView(v, 160, 1)
	if shown(h, "查看") || !shown(h, "Reset：Value") {
		t.Fatal("SetItems did not clear rich entries/separators")
	}
}
