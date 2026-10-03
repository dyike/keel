package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestSheetPanelStylePixels(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	s := kit.Sheet(el.Right, "Styled sheet").Size(180).MarginTop(32).PanelStyle(func(e *el.DivEl) { e.Bg(red).Border(4, blue).P(12).Gap(8) })
	s.SetValue(true)
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(s)})
	sample := func(border bool) color.NRGBA {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		e := element(t, w, "Styled sheet")
		if e.Role != "dialog" {
			t.Fatal("not dialog", e)
		}
		x := e.X + e.Width/2
		if border {
			x = e.X + 1
		}
		return color.NRGBAModel.Convert(im.At(x, e.Y+e.Height/2)).(color.NRGBA)
	}
	if got := sample(false); got != red {
		t.Fatal("background", got)
	}
	if got := sample(true); got.B < 252 || got.R > 3 || got.G > 3 || got.A != 255 {
		t.Fatal("border", got)
	}
	s.PanelStyle(nil)
	if got := sample(false); got == red {
		t.Fatal("default not restored")
	}
}
