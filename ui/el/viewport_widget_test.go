package el

import (
	"fmt"
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

// A wrapped embedded tree keeps its natural height, but only paints children
// inside the parent's scroller. Padding/translation and click routing survive.
func TestEmbeddedWidgetRetainsScrollViewport(t *testing.T) {
	painted, clicked := 0, -1
	inner := Embed(ViewFunc(func(*Context) Element {
		column := Div()
		for i := range 100 {
			column.Child(Div().W(Full).H(Dp(25)).Name(fmt.Sprintf("row-%d", i)).OnClick(func() { clicked = i }).
				Decorate(func(gtx core.C, draw func()) {
					if gtx.Enabled() {
						painted++
					}
					draw()
				}).Child(Text(fmt.Sprint(i))))
		}
		return column
	}))
	var visible image.Rectangle
	wrapped := core.ViewportFunc(func(gtx core.C, view image.Rectangle) core.D {
		if gtx.Enabled() {
			visible = view
		}
		return inner.LayoutViewport(gtx, view)
	})
	root := Root(ViewFunc(func(*Context) Element {
		return Div().P(20).Child(Div().ID("scroller").W(Dp(200)).H(Dp(100)).ScrollY().Child(Widget(wrapped)))
	}))
	frame := 0
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = time.Unix(100, 0).Add(time.Duration(frame) * time.Second / 60)
		frame++
		painted = 0
		root.Layout(gtx)
	})
	check := func() {
		t.Helper()
		if painted < 1 || painted > 5 {
			t.Fatalf("painted %d rows in a 100dp viewport", painted)
		}
		if visible.Dy() != 100 || visible.Dx() != 200 {
			t.Fatalf("parent viewport lost: %v", visible)
		}
		var content int
		for _, state := range root.store.states {
			if state.id == "scroller" {
				content = state.scrollContent
			}
		}
		if content != 2500 {
			t.Fatalf("natural content height = %d, want 2500", content)
		}
	}
	h.Frame()
	check()
	h.Click(50, 32)
	if clicked != 0 {
		t.Fatalf("first row click selected %d", clicked)
	}
	h.Scroll(50, 50, 100)
	h.Frame()
	check()
	if visible.Min.Y <= 0 {
		t.Fatalf("scroll translation lost: %v", visible)
	}
	h.Click(50, 32)
	if clicked <= 0 {
		t.Fatalf("scrolled click still selected %d", clicked)
	}
}
