package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestMessageSurfaceInsets(t *testing.T) {
	for _, scale := range []int{1, 2} {
		content := el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(60)).H(el.Dp(30)).Name("Body") })
		surface := Bubble(content).Mine()
		msg := Message("Author", nil).Avatar(nil).Bubble(surface).Header(Label("Header")).Footer(Label("Footer"))
		h := renderView(msg, 220, scale)
		normal := bounds(h, "Header").Min.X
		if normal != int(theme.SpaceLg)*scale || bounds(h, "Body").Min.X != normal || !surface.mine {
			t.Fatal("typed surface alignment or source mutation")
		}
		surface.Variant(BubbleGhost)
		h.Frame()
		if bounds(h, "Header").Min.X != 0 || bounds(h, "Footer").Min.X != 0 || bounds(h, "Body").Min.X != 0 {
			t.Fatal("ghost inherited inset")
		}
		msg.HeaderInset(true).FooterInset(false)
		h.Frame()
		if bounds(h, "Header").Min.X != normal || bounds(h, "Footer").Min.X != 0 {
			t.Fatal("independent override")
		}
		surface.Variant(BubbleSecondary)
		h.Frame()
		if bounds(h, "Footer").Min.X != 0 {
			t.Fatal("explicit false lost")
		}
		msg.ResetContentInsets()
		h.Frame()
		if bounds(h, "Footer").Min.X != normal {
			t.Fatal("reset automatic inset")
		}
		msg.Content(content)
		h.Frame()
		if bounds(h, "Header").Min.X != 0 || bounds(h, "Body").Min.X != 0 {
			t.Fatal("plain content retained surface")
		}
		msg.Bubble(nil)
		h.Frame()
		if shown(h, "Body") || !shown(h, "Header") {
			t.Fatal("clear typed body")
		}
	}
}
