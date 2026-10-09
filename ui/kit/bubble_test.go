package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestBubbleReactionLayoutAndState(t *testing.T) {
	for _, scale := range []int{1, 2} {
		clicks := 0
		input := Input("Draft")
		reactions := el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Child(Button("Like", func() { clicks++ }).Size(24).Render(cx))
		})
		v := Bubble(input).Reactions(reactions).PartStyle(BubblePartContent, func(e *el.DivEl) { e.Name("Surface") })
		h := renderView(v, 240, scale)
		before := bounds(h, "Surface")
		if before.Dx() > 180*scale || bounds(h, "Like").Min.Y < before.Max.Y {
			t.Fatal("bottom reaction layout", before, bounds(h, "Like"))
		}
		click(t, h, "Like")
		v.ReactionSide(BubbleReactionTop).Alignment(el.End).Variant(BubbleOutline)
		h.Frame()
		h.Key(key.NameSpace, 0)
		if clicks != 2 {
			t.Fatal("reaction side or variant lost focus", clicks)
		}
		after := bounds(h, "Surface")
		if bounds(h, "Like").Max.Y > after.Min.Y || after.Max.X != 240*scale {
			t.Fatal("top/end layout", after, bounds(h, "Like"))
		}
		clickClass(t, h, "Editor", "Draft")
		h.Type("hello")
		v.Variant(BubbleGhost).ReactionAlignment(el.Start)
		h.Frame()
		h.Key(key.NameRightArrow, 0)
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "ello" {
			t.Fatal("content focus lost", input.Value())
		}
		if bounds(h, "Surface").Dx() != 240*scale {
			t.Fatal("ghost not full width", bounds(h, "Surface"))
		}
		v.Reactions(nil)
		h.Frame()
		if shown(h, "Like") {
			t.Fatal("reaction not removed")
		}
		v.Reactions(reactions)
		h.Frame()
		click(t, h, "Like")
		if clicks != 3 {
			t.Fatal("reaction not restored")
		}
		v.PartStyle(BubblePartRoot, func(e *el.DivEl) { e.Disabled(true) })
		h.Frame()
		click(t, h, "Like")
		if clicks != 3 {
			t.Fatal("ancestor disable ignored")
		}
	}
}
