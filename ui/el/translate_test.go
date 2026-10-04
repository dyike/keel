package el

import (
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestTranslateMovesCullingHitAreaAndAnchorWithoutLayout(t *testing.T) {
	calls := 0
	shift := float32(-450)
	var target *DivEl
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		target = Div().ID("translated").Name("target").W(Dp(30)).H(Dp(20)).NoShrink().Translate(shift, 30).OnClick(func() { calls++ })
		cx.Overlay("tip", Anchored("translated", Div().Name("tip").W(Dp(30)).H(Dp(10))).Placement(Bottom, Start))
		return Div().Row().Child(Div().W(Dp(500)).NoShrink(), target)
	})))
	if rect(target).Min != image.Pt(500, 0) {
		t.Fatal("translation changed layout", rect(target))
	}
	if got := nodeBounds(h, "target"); got != image.Rect(50, 30, 80, 50) {
		t.Fatal("translated visible area", got)
	}
	if got := nodeBounds(h, "tip"); got.Min != image.Pt(50, 54) {
		t.Fatal("anchor did not move", got)
	}
	h.Click(60, 40)
	if calls != 1 {
		t.Fatal("hit area did not move", calls)
	}
	shift = 0
	h.Frame()
	if !nodeBounds(h, "target").Empty() || !nodeBounds(h, "tip").Empty() {
		t.Fatal("offscreen translated content remained visible")
	}
}
