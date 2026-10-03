package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestPopoverPanelStylePixels(t *testing.T) {
	p := kit.Popover(el.ViewFunc(func(*el.Context) el.Element { return el.Div().Size(el.Dp(40)) })).Width(120)
	p.Trigger(kit.Button("Open", p.Toggle))
	p.SetValue(true)
	red := color.NRGBA{R: 255, A: 255}
	p.Appearance(false).PanelStyle(func(e *el.DivEl) { e.Bg(red).P(10).Rounded(0).Name("Panel") })
	w := openTest(t, Options{Width: 300, Height: 220, Content: el.Root(p)})
	sample := func() color.NRGBA {
		t.Helper()
		e := element(t, w, "Panel")
		if e.Role != "dialog" {
			t.Fatal(e)
		}
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(e.X+e.Width/2, e.Y+e.Height/2)).(color.NRGBA)
	}
	if got := sample(); got != red {
		t.Fatal("custom background", got)
	}
	p.PanelStyle(func(e *el.DivEl) { e.Name("Panel") })
	if got := sample(); got == red {
		t.Fatal("old custom background retained", got)
	}
	p.Appearance(true)
	if got := sample(); got == red {
		t.Fatal("default appearance not restored", got)
	}
}
