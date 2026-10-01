package window

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestHorizontalScrollSnapshotAndClick(t *testing.T) {
	var cx *el.Context
	calls := 0
	w := openTest(t, Options{Width: 240, Height: 120, Content: el.Root(el.ViewFunc(func(c *el.Context) el.Element {
		cx = c
		return el.Div().Items(el.Start).Child(el.Div().ID("strip").W(el.Dp(100)).H(el.Dp(50)).ScrollX().Child(
			el.Div().Row().Child(el.Div().W(el.Dp(160)).NoShrink().Child(el.Text("start")), el.Div().W(el.Dp(100)).NoShrink().H(el.Dp(40)).Name("end").OnClick(func() { calls++ }))))
	}))})
	for _, e := range w.snapshot() {
		if e.Name == "end" {
			t.Fatal("offscreen button in snapshot")
		}
	}
	cx.ScrollIntoViewX("strip", 160, 260)
	w.render()
	e := element(t, w, "end")
	if e.X < 0 || e.X >= 100 {
		t.Fatalf("shifted bounds %+v", e)
	}
	w.click(e.center())
	if calls != 1 {
		t.Fatal("scrolled button not clickable")
	}
}
