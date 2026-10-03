package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
)

func TestResizableHandleAppearancePixels(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		split := kit.Resizable(nil, nil).Min(40, 40).HandleAppearance(func(a kit.ResizableHandleAppearance) kit.ResizableHandleAppearance {
			a.Idle = 1
			a.Hover = 4
			a.Color = color.NRGBA{R: 255, A: 255}
			a.ActiveColor = a.Color
			a.Duration = 0
			return a
		})
		split.SetValue(100)
		if vertical {
			split.Vertical()
		}
		w := openTest(t, Options{Width: 300, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(300)).H(el.Dp(300)).Items(el.Stretch).Child(split.Render(cx))
		}))})
		count := func() int {
			t.Helper()
			e := element(t, w, locale.Current().Resize)
			data, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			n := 0
			for i := 0; i < 6; i++ {
				x, y := e.X+i, e.Y+20
				if vertical {
					x, y = e.X+20, e.Y+i
				}
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if c.R > 240 && c.G < 10 && c.B < 10 {
					n++
				}
			}
			if (!vertical && e.Width != 6) || (vertical && e.Height != 6) {
				t.Fatal("appearance changed hit area", e)
			}
			return n
		}
		if n := count(); n != 1 {
			t.Fatal("idle width", n)
		}
		w.press("Tab")
		if n := count(); n != 4 {
			t.Fatal("focused width", n)
		}
		split.SetDisabled(true)
		if n := count(); n != 1 {
			t.Fatal("disabled width", n)
		}
	}
}
