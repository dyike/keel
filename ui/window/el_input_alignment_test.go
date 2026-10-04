package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestInputTextAlignsWithLeadingIcon(t *testing.T) {
	for _, value := range []string{"H", "国"} {
		w := openTest(t, Options{Width: 300, Height: 80, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().Items(el.Center).Gap(8).P(8).H(el.Dp(40)).Child(
				kit.Icon(kit.IconSearch).Size(18).Color(color.NRGBA{B: 255, A: 255}).Render(cx),
				el.Input().Bind(&value).TextSize(15).TextColor(color.NRGBA{R: 255, A: 255}).Border(0, color.NRGBA{}).P(0).MinH(el.Auto).Grow())
		}))})
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		ink, icon := image.Rectangle{}, image.Rectangle{}
		for y := 0; y < 80; y++ {
			for x := 0; x < 300; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				p := image.Rect(x, y, x+1, y+1)
				if int(c.R)-int(c.B) > 80 {
					ink = ink.Union(p)
				}
				if int(c.B)-int(c.R) > 80 {
					icon = icon.Union(p)
				}
			}
		}
		delta := ink.Min.Y + ink.Max.Y - icon.Min.Y - icon.Max.Y
		if ink.Empty() || icon.Empty() || delta < -4 || delta > 4 {
			t.Fatalf("%s: text %v icon %v", value, ink, icon)
		}
	}
}
