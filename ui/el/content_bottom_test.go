package el

import (
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestContentBottomFirstFrameAndGrowingContent(t *testing.T) {
	for _, scale := range []int{1, 2} {
		contentHeight, avatarHeight := float32(20), float32(40)
		calls := 0
		var row, body, avatar, content, footer *DivEl
		root := Root(viewFunc(func(*Context) Element {
			content = Div().H(Dp(contentHeight))
			footer = Div().H(Dp(30)).OnClick(func() { calls++ })
			body = Div().W(Dp(100)).Gap(5).Child(Div().H(Dp(10)), content, footer).ContentBottom(content)
			avatar = Div().W(Dp(30)).H(Dp(avatarHeight))
			row = Div().Row().Items(ContentBottom).Gap(8).Child(avatar, body)
			return Div().Items(Start).Child(row)
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(400*scale, 400*scale)
			root.Layout(gtx)
		})
		verify := func() {
			t.Helper()
			cb := body.n.pos.Y + content.n.pos.Y + content.n.size.Y
			ab := avatar.n.pos.Y + avatar.n.size.Y
			fb := body.n.pos.Y + footer.n.pos.Y + footer.n.size.Y
			if cb != ab || fb != row.n.size.Y || body.n.pos.Y < 0 || avatar.n.pos.Y < 0 {
				t.Fatalf("alignment: content=%d avatar=%d footer=%d row=%d", cb, ab, fb, row.n.size.Y)
			}
		}
		verify()
		h.Click(float32(body.n.pos.X+10*scale), float32(body.n.pos.Y+footer.n.pos.Y+10*scale))
		if calls != 1 {
			t.Fatal("footer hit geometry")
		}
		avatarHeight = 100
		h.Frame()
		verify()
		contentHeight = 120
		h.Frame()
		verify()
		avatarHeight = 10
		h.Frame()
		verify()
	}
}

func TestContentBottomMissingAndHiddenTargets(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		target := Div().H(Dp(30)).Hidden(hidden)
		body := Div().W(Dp(100)).Child(Div().H(Dp(10))).ContentBottom(target)
		if hidden {
			body.Child(target)
		}
		avatar := Div().W(Dp(20)).H(Dp(50))
		row := Div().Row().Items(ContentBottom).Child(avatar, body)
		render(t, Div().Items(Start).Child(row))
		if body.n.pos.Y+body.n.size.Y != 50 || row.n.size.Y != 50 {
			t.Fatal("fallback bottom")
		}
	}
}

func TestContentBottomWrappedRows(t *testing.T) {
	firstTarget := Div().H(Dp(20))
	firstBody := Div().W(Dp(70)).Gap(5).Child(Div().H(Dp(10)), firstTarget, Div().H(Dp(30))).ContentBottom(firstTarget)
	firstAvatar := Div().W(Dp(30)).H(Dp(40))
	secondTarget := Div().H(Dp(80))
	secondBody := Div().W(Dp(70)).Gap(5).Child(Div().H(Dp(10)), secondTarget, Div().H(Dp(30))).ContentBottom(secondTarget)
	secondAvatar := Div().W(Dp(30)).H(Dp(20))
	row := Div().Wrap().W(Dp(120)).Gap(5).Items(ContentBottom).Child(firstAvatar, firstBody, secondAvatar, secondBody)
	render(t, Div().Items(Start).Child(row))
	if firstAvatar.n.pos.Y+firstAvatar.n.size.Y != firstBody.n.pos.Y+firstTarget.n.pos.Y+firstTarget.n.size.Y {
		t.Fatal("first line anchor")
	}
	if secondAvatar.n.pos.Y+secondAvatar.n.size.Y != secondBody.n.pos.Y+secondTarget.n.pos.Y+secondTarget.n.size.Y {
		t.Fatal("second line anchor")
	}
	if secondBody.n.pos.Y < firstBody.n.pos.Y+firstBody.n.size.Y {
		t.Fatal("wrapped lines overlap")
	}
}
