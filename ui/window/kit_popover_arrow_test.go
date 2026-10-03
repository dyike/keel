package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestPopoverArrowPixels(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	p := kit.Popover(el.ViewFunc(func(*el.Context) el.Element { return el.Div().Size(el.Dp(40)) })).Width(80).Arrow(true).PanelStyle(func(e *el.DivEl) { e.Name("Panel").Bg(red) })
	p.Trigger(kit.Button("Target", p.Toggle))
	p.SetValue(true)
	w := openTest(t, Options{Width: 480, Height: 420, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(140).Items(el.Start).Child(p.Render(cx)) }))})
	for _, side := range []el.Side{el.Top, el.Bottom, el.Left, el.Right} {
		p.Placement(side, el.Center)
		e := element(t, w, "Panel")
		x, y := e.X+e.Width/2, e.Y+e.Height/2
		switch side {
		case el.Top:
			y = e.Y + e.Height + 3
		case el.Bottom:
			y = e.Y - 3
		case el.Left:
			x = e.X + e.Width + 3
		case el.Right:
			x = e.X - 3
		}
		capture := func() color.NRGBA {
			b, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			im, err := png.Decode(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			return color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
		}
		if got := capture(); got != red {
			t.Fatal("arrow fill", side, e, got)
		}
		p.Arrow(false)
		// Disabling the arrow moves the panel 6dp toward the trigger.
		switch side {
		case el.Top:
			y += 6
		case el.Bottom:
			y -= 6
		case el.Left:
			x += 6
		case el.Right:
			x -= 6
		}
		if got := capture(); got == red {
			t.Fatal("arrow retained", side)
		}
		p.Arrow(true)
	}
}
