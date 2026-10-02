package el

import (
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

// A press in an input's padding, above its centered line, focuses the text.
func TestInputPaddingPressFocuses(t *testing.T) {
	var s string
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		return Div().P(20).Items(Start).Child(Input().ID("q").Bind(&s).W(Dp(200)))
	})))
	h.Click(100, 22) // 2dp inside the 36dp box, well above the text line
	h.Type("x")
	if s != "x" {
		t.Fatalf("typed into %q; a press in the padding should focus the input", s)
	}
}

// FocusOnPress forwards a press on a frame's blank space to the input in it,
// and a press on the input itself still lands there.
func TestFocusOnPressForwardsToInput(t *testing.T) {
	var s string
	other := false
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		return Div().P(20).Gap(8).Items(Start).Child(
			Div().ID("frame").FocusOnPress("q").Row().W(Dp(240)).H(Dp(60)).Pl(80).Items(Center).
				Child(Input().ID("q").Bind(&s).P(0).Grow()),
			Div().Size(Dp(40)).OnClick(func() { other = true }),
		)
	})))
	h.Click(40, 50) // left of the input, inside the frame
	h.Type("a")
	if s != "a" {
		t.Fatalf("typed into %q; a press on the frame should focus its input", s)
	}
	h.Click(40, 108) // elsewhere: the frame must not swallow other presses
	h.Type("b")
	if !other || s != "a" {
		t.Fatalf("other=%t s=%q; a press outside the frame should blur the input", other, s)
	}
}
