package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestMessageIndependentAlignment(t *testing.T) {
	for _, scale := range []int{1, 2} {
		content := el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(60)).H(el.Dp(30)).Name("Body") })
		avatar := el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(28)).H(el.Dp(28)).Name("Identity") })
		for _, user := range []bool{false, true} {
			surface := Bubble(content)
			msg := Message("Author", nil).Bubble(surface).Avatar(avatar).Header(Label("Header")).Footer(Label("Footer"))
			if user {
				msg.User()
			}
			msg.Alignment(el.Start)
			h := renderView(msg, 240, scale)
			left := bounds(h, "Body")
			variant := msg.bubble.variant
			if bounds(h, "Identity").Min.X >= left.Min.X {
				t.Fatal("leading avatar")
			}
			msg.Alignment(el.End).Alignment(el.Center)
			h.Frame()
			right := bounds(h, "Body")
			if right.Min.X <= left.Min.X || bounds(h, "Identity").Min.X <= right.Min.X || msg.bubble.variant != variant {
				t.Fatal("alignment changed surface or failed layout")
			}
			if bounds(h, "Header").Min.X <= left.Min.X || bounds(h, "Footer").Min.X <= left.Min.X {
				t.Fatal("metadata did not align")
			}
			msg.ResetAlignment()
			h.Frame()
			expected := left
			if user {
				expected = right
			}
			if bounds(h, "Body") != expected {
				t.Fatal("reset alignment")
			}
			if surface.mine || surface.variant != BubbleAuto {
				t.Fatal("source mutated")
			}
		}
		raw := Message("Plain", content).Avatar(nil)
		h := renderView(raw, 240, scale)
		before := bounds(h, "Body")
		raw.Alignment(el.End)
		h.Frame()
		if bounds(h, "Body").Min.X <= before.Min.X {
			t.Fatal("plain content did not align")
		}
	}
}
