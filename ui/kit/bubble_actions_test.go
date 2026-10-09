package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestBubbleReactionActions(t *testing.T) {
	for _, scale := range []int{1, 2} {
		firstCalls, secondCalls := 0, 0
		first := Button("Like", func() { firstCalls++ }).Variant(ButtonGhost).Size(24)
		second := Button("Copy", func() { secondCalls++ }).Variant(ButtonSecondary).Size(28)
		input := Input("Reaction note")
		buttons := []*ButtonView{first, nil, second}
		v := Bubble(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Message") })).Reactions(input).ReactionActions(buttons...)
		buttons[0] = nil
		h := renderView(v, 200, scale)
		click(t, h, "Like")
		v.ReactionActions(second, first).ReactionSide(BubbleReactionTop)
		h.Frame()
		h.Key(key.NameSpace, 0)
		if firstCalls != 2 || secondCalls != 0 {
			t.Fatal("reorder lost action focus", firstCalls, secondCalls)
		}
		for _, name := range []string{"Like", "Copy"} {
			r := bounds(h, name)
			if r.Min.X < 0 || r.Max.X > 150*scale {
				t.Fatal("reaction escaped narrow bubble", name, r)
			}
		}
		clickClass(t, h, "Editor", "Reaction note")
		h.Type("hello")
		v.ReactionActions()
		h.Frame()
		h.Key(key.NameRightArrow, 0)
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "ello" || shown(h, "Like") {
			t.Fatal("clearing actions lost generic slot or focus")
		}
		v.ReactionActions(first, second)
		h.Frame()
		first.SetLoading(true)
		h.Frame()
		click(t, h, "Like")
		if firstCalls != 2 {
			t.Fatal("loading action activated")
		}
		first.SetLoading(false)
		first.SetDisabled(true)
		h.Frame()
		click(t, h, "Like")
		if firstCalls != 2 {
			t.Fatal("disabled action activated")
		}
		first.SetDisabled(false)
		h.Frame()
		click(t, h, "Like")
		if firstCalls != 3 || first.height != 24 || first.variant != ButtonGhost || first.id != "" {
			t.Fatal("action configuration changed")
		}
		v.PartStyle(BubblePartRoot, func(e *el.DivEl) { e.Disabled(true) })
		h.Frame()
		click(t, h, "Copy")
		if secondCalls != 0 {
			t.Fatal("ancestor disabled ignored")
		}
		v.PartStyle(BubblePartRoot, nil).ReactionActions(nil).Reactions(nil)
		h.Frame()
		if shown(h, "Copy") || shown(h, "Reaction note") {
			t.Fatal("empty reactions retained controls")
		}
	}
}
