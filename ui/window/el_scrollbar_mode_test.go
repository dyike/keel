package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"image/color"
	"image/png"
	"testing"
)

func TestScrollbarModesPaintAndPreserveLayout(t *testing.T) {
	bg := color.NRGBA{R: 230, G: 210, B: 170, A: 255}
	mode := el.ScrollbarScrolling
	w := openTest(t, Options{Width: 160, Height: 160, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Items(el.Start).Child(el.Div().ID("scroll").W(el.Dp(100)).H(el.Dp(100)).ScrollY().Scrollbars(mode).Child(el.Div().W(el.Dp(100)).H(el.Dp(400)).Bg(bg)))
	}))})
	pixel := func() color.NRGBA {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(95, 12)).(color.NRGBA)
	}
	if got := pixel(); got != bg {
		t.Fatal("idle scrolling mode painted scrollbar", got)
	}
	mode = el.ScrollbarAlways
	if got := pixel(); got == bg {
		t.Fatal("always mode did not paint scrollbar")
	}
	mode = el.ScrollbarScrolling
	if got := pixel(); got != bg {
		t.Fatal("mode switch left bar visible", got)
	}
}
