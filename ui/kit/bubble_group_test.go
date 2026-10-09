package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestBubbleGroupReorderAndFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		input := Input("Draft")
		first := Bubble(input)
		calls := 0
		second := Bubble(Button("Reply", func() { calls++ })).Mine()
		items := []el.View{first, nil, second}
		group := BubbleGroup(items...).Name("Conversation")
		items[0] = nil
		copy := group.Items()
		copy[0] = nil
		if len(group.Items()) != 2 || group.Items()[0] != first {
			t.Fatal("items not copied")
		}
		h := renderView(group, 240, scale)
		firstRect := bounds(h, "Draft")
		secondRect := bounds(h, "Reply")
		if firstRect.Min.Y >= secondRect.Min.Y || secondRect.Max.X > 240*scale {
			t.Fatal("group layout", firstRect, secondRect)
		}
		clickClass(t, h, "Editor", "Draft")
		h.Type("hello")
		group.SetItems(second, first)
		h.Frame()
		h.Key(key.NameRightArrow, 0)
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "ello" {
			t.Fatal("reordering lost focus", input.Value())
		}
		if bounds(h, "Reply").Min.Y >= bounds(h, "Draft").Min.Y {
			t.Fatal("reorder ignored")
		}
		click(t, h, "Reply")
		group.Gap(20).Gap(-1).Gap(float32(math.NaN())).Gap(float32(math.Inf(1)))
		h.Frame()
		h.Key(key.NameSpace, 0)
		if calls != 2 || group.gap != 20 {
			t.Fatal("gap changed focus or accepted invalid value", calls, group.gap)
		}
		group.SetDisabled(true)
		h.Frame()
		click(t, h, "Reply")
		if calls != 2 {
			t.Fatal("disabled group activated")
		}
		group.SetDisabled(false)
		group.Style(func(e *el.DivEl) { e.P(10).Gap(3) })
		h.Frame()
		click(t, h, "Reply")
		if calls != 3 {
			t.Fatal("restored group disabled")
		}
		group.Style(nil).SetItems(first)
		h.Frame()
		if shown(h, "Reply") || input.Value() != "ello" {
			t.Fatal("removal changed retained content")
		}
		group.SetItems()
		h.Frame()
		if shown(h, "Draft") {
			t.Fatal("empty group retained content")
		}
	}
}
