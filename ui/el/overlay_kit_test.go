package el

import (
	"image"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

// A layer without OnDismiss (a notification stack) must not swallow Escape
// meant for the dialog below it.
func TestOverlayEscapeSkipsLayersWithoutDismiss(t *testing.T) {
	open, closes := true, 0
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		if open {
			cx.Overlay("dialog", Modal(Div().Size(Dp(100)).Name("dialog")).OnDismiss(func() { open = false; closes++ }))
		}
		cx.Overlay("toasts", Anchored("corner", Div().Size(Dp(40)).Name("toasts")))
		return Div().Child(Div().ID("corner").Size(Dp(1)))
	})))
	h.Key(key.NameEscape, 0)
	h.Frame()
	if closes != 1 || open {
		t.Fatalf("escape did not reach the dialog: closes=%d", closes)
	}
}

func TestModalEdgePlacement(t *testing.T) {
	for _, tc := range []struct {
		side  Side
		align Align
		want  image.Point
	}{
		{Right, Start, image.Pt(300, 0)}, {Left, Start, image.Pt(0, 0)},
		{Bottom, Center, image.Pt(150, 250)}, {Top, End, image.Pt(300, 0)},
	} {
		h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
			cx.Overlay("sheet", Modal(Div().Name("sheet").W(Dp(100)).H(Dp(50))).Placement(tc.side, tc.align))
			return Div()
		})))
		if b := nodeBounds(h, "sheet"); b.Min != tc.want {
			t.Errorf("side %d align %d: %v want %v", tc.side, tc.align, b.Min, tc.want)
		}
	}
}

func TestFocusWithin(t *testing.T) {
	var inside, outside bool
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		inside, outside = cx.FocusWithin("wrap"), cx.FocusWithin("other")
		return Div().Child(
			Div().ID("wrap").Child(Div().ID("btn").Name("btn").Size(Dp(20)).OnClick(func() {})),
			Div().ID("other").Size(Dp(20)),
		)
	})))
	b := nodeBounds(h, "btn")
	h.Click(float32(b.Min.X+5), float32(b.Min.Y+5)) // a press focuses a focusable element
	h.Frame()
	if !inside || outside {
		t.Fatalf("focus within: wrap=%v other=%v", inside, outside)
	}
}
