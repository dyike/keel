package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"math"
	"testing"
)

func TestMessageMixedContent(t *testing.T) {
	for _, scale := range []int{1, 2} {
		say := func(name string) el.View {
			return el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(60)).H(el.Dp(25)).Name(name) })
		}
		first, second := Bubble(say("First")), Bubble(say("Second")).Variant(BubbleGhost)
		action := Button("Attachment", nil)
		items := []el.View{first, nil, action, second}
		mixed := MessageContent(items...)
		items[0] = nil
		copy := mixed.Items()
		copy[0] = nil
		if len(mixed.Items()) != 3 || mixed.Items()[0] != first {
			t.Fatal("slice isolation")
		}
		msg := Message("Author", mixed).Avatar(nil).Header(Label("Header")).Footer(Label("Footer"))
		h := renderView(msg, 240, scale)
		if bounds(h, "Header").Min.X != 0 || bounds(h, "Footer").Min.X != 0 {
			t.Fatal("ghost metadata")
		}
		if bounds(h, "First").Min.Y >= bounds(h, "Attachment").Min.Y || bounds(h, "Attachment").Min.Y >= bounds(h, "Second").Min.Y {
			t.Fatal("mixed ordering")
		}
		second.Variant(BubbleOutline)
		h.Frame()
		if bounds(h, "Header").Min.X != int(theme.SpaceLg)*scale {
			t.Fatal("dynamic ghost removal")
		}
		left := bounds(h, "First").Min.X
		msg.Alignment(el.End)
		h.Frame()
		if bounds(h, "First").Min.X <= left || first.mine || second.mine {
			t.Fatal("alignment or source mutation")
		}
		msg.HeaderInset(false)
		mixed.SetItems(second, first)
		h.Frame()
		if shown(h, "Attachment") || bounds(h, "Second").Min.Y >= bounds(h, "First").Min.Y {
			t.Fatal("reorder/remove")
		}
		mixed.Gap(20).Gap(-1).Gap(float32(math.NaN()))
		if mixed.gap != 20 {
			t.Fatal("invalid gap")
		}
		mixed.SetItems(first)
		if len(mixed.bubbles) != 1 {
			t.Fatal("removed bubble retained")
		}
		mixed.SetItems()
		h.Frame()
		if shown(h, "First") || shown(h, "Second") {
			t.Fatal("clear")
		}
	}
}
