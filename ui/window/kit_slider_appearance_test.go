package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestSliderCustomAppearancePixels(t *testing.T) {
	track := color.NRGBA{R: 180, G: 20, B: 20, A: 255}
	fill := color.NRGBA{R: 20, G: 170, B: 30, A: 255}
	thumb := color.NRGBA{R: 20, G: 30, B: 180, A: 255}
	border := color.NRGBA{R: 180, G: 20, B: 180, A: 255}
	for _, vertical := range []bool{false, true} {
		v := kit.Slider("Level", 0, 100).Appearance(func(a *kit.SliderAppearance) {
			a.TrackSize, a.ThumbSize, a.ThumbBorderWidth = 12, 32, 3
			a.TrackColor, a.FillColor, a.ThumbColor, a.ThumbBorderColor = track, fill, thumb, border
			a.TrackRadius, a.ThumbRadius = 0, 0
		})
		v.SetValue(50)
		if vertical {
			v.Vertical(200)
		}
		w := openTest(t, Options{Width: 260, Height: 280, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(232)).Child(v.Render(cx)) }))})
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		counts := map[color.NRGBA]int{}
		for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
			for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
				counts[color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)]++
			}
		}
		for _, c := range []color.NRGBA{track, fill, thumb, border} {
			if counts[c] < 100 {
				t.Fatalf("vertical=%v missing custom color %v: %d pixels", vertical, c, counts[c])
			}
		}
	}
}
