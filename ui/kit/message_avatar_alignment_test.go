package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestMessageAvatarAtContentBottom(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, end := range []bool{false, true} {
			height := float32(60)
			avatarHeight := float32(28)
			content := el.ViewFunc(func(*el.Context) el.Element { return el.Div().WFull().H(el.Dp(height)).Name("Body") })
			avatar := el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(40)).H(el.Dp(avatarHeight)).Name("Identity") })
			msg := Message("Author", content).Avatar(avatar).Header(Label("Header")).Footer(Button("Footer", nil)).Actions(Button("Action", nil))
			if end {
				msg.Alignment(el.End)
			}
			h := renderView(msg, 240, scale)
			verify := func() {
				t.Helper()
				a, b, f := bounds(h, "Identity"), bounds(h, "Body"), bounds(h, "Footer")
				if a.Max.Y != b.Max.Y || f.Min.Y <= a.Max.Y || a.Min.Y < 0 {
					t.Fatal("avatar/body/footer", a, b, f)
				}
			}
			verify()
			height = 160
			h.Frame()
			verify()
			avatarHeight = 220
			h.Frame()
			verify()
			msg.Footer(nil).Actions()
			h.Frame()
			if bounds(h, "Identity").Max.Y != bounds(h, "Body").Max.Y {
				t.Fatal("removing footer changed anchor")
			}
		}
	}
}
