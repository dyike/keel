package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestNumberInputPlainAppearancePixels(t *testing.T) {
	v := kit.NumberInput("")
	backdrop := color.NRGBA{R: 120, G: 40, B: 80, A: 255}
	w := openTest(t, Options{Width: 260, Height: 100, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Bg(backdrop).P(16).Child(v.Render(cx))
	}))})
	sample := func(y int) color.NRGBA {
		t.Helper()
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(80, y)).(color.NRGBA)
	}
	if sample(16) == backdrop || sample(18) == backdrop {
		t.Fatal("default frame missing")
	}
	v.Appearance(false)
	if sample(16) != backdrop || sample(18) != backdrop {
		t.Fatal("plain frame still painted")
	}
	v.Appearance(true)
	if sample(16) == backdrop || sample(18) == backdrop {
		t.Fatal("frame not restored")
	}
}
