package window

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/kit"
)

func TestKitImageFitAndRoundedPixels(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 300, 100))
	colors := []color.NRGBA{{R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255}}
	for y := 0; y < 100; y++ {
		for x := 0; x < 300; x++ {
			pixels.SetNRGBA(x, y, colors[x/100])
		}
	}
	for _, fit := range []kit.ImageFit{kit.ImageContain, kit.ImageCover, kit.ImageFill} {
		pic := kit.Image(pixels, "photo").Size(100, 100).Fit(fit).Rounded(20)
		w := openTest(t, kitPage(pic))
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		e := element(t, w, "photo")
		at := func(x, y int) color.NRGBA { return color.NRGBAModel.Convert(img.At(e.X+x, e.Y+y)).(color.NRGBA) }
		if at(50, 50) != colors[1] {
			t.Fatal("image center", fit, at(50, 50))
		}
		if fit == kit.ImageCover && at(15, 50) != colors[1] {
			t.Fatal("cover did not crop")
		}
		if fit != kit.ImageCover && at(15, 50) != colors[0] {
			t.Fatal("fit lost image edges")
		}
		if at(0, 0) == colors[0] || at(0, 0) == colors[1] {
			t.Fatal("rounded corner not clipped")
		}
		if fit == kit.ImageContain && at(50, 10) == colors[1] {
			t.Fatal("contain stretched vertically")
		}
		if fit != kit.ImageContain && at(50, 10) != colors[1] {
			t.Fatal("fixed frame not filled")
		}
	}
}
