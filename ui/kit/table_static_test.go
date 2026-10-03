package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestStaticTableCompositionAndInteraction(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		input := Input("").Placeholder("note")
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
			return StaticTable().W(el.Dp(480)).Name("invoices").Child(
				TableHeader().Child(TableRow().Child(TableHead().Flex(0).W(el.Dp(100)).Child(el.Text("header fixed")), TableHead().Child(el.Text("header flex")))),
				TableBody().Child(TableRow().Child(TableDataCell().Flex(0).W(el.Dp(100)).Child(el.Text("body fixed")), TableDataCell().Child(Button("action", func() { calls++ }).Render(cx)))),
				TableFooter().Child(TableRow().Child(TableDataCell().Flex(0).W(el.Dp(100)).Child(el.Text("footer fixed")), TableDataCell().Child(input.Render(cx)))),
				TableCaption().Child(el.Text("caption")),
			)
		}), 500, scale)
		if _, ok := semanticNode(h, "table"); !ok {
			t.Fatal("missing table semantics")
		}
		a, b, c := bounds(h, "header fixed"), bounds(h, "body fixed"), bounds(h, "footer fixed")
		if a.Min.X != b.Min.X || b.Min.X != c.Min.X || a.Min.Y >= b.Min.Y || b.Min.Y >= c.Min.Y || bounds(h, "caption").Min.Y < c.Max.Y {
			t.Fatal("section layout", a, b, c)
		}
		click(t, h, "action")
		h.Frame()
		if calls != 1 {
			t.Fatal("child action", calls)
		}
		clickClass(t, h, "Editor", "note")
		h.Type("retained")
		h.Frame()
		h.Frame()
		if input.Value() != "retained" {
			t.Fatal("retained child view", input.Value())
		}
	}
}
