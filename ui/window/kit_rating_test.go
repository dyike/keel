package window

import (
	"bytes"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestKitRatingColorFractionAndSize(t *testing.T) {
	v := kit.Rating("Average", 2).ReadOnly().Size(32).Color(color.NRGBA{R: 255, A: 255})
	v.SetScore(1.5)
	w := openTest(t, kitPage(v))
	var rating Element
	for _, e := range w.snapshot() {
		if e.Role == "slider" && e.Name == "Average" {
			rating = e
		}
	}
	if rating.Value != "1.5/2" || rating.Height != 32 {
		t.Fatal("rating semantics/size", rating)
	}
	b, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	red := func(x0, x1 int) int {
		n := 0
		for y := rating.Y; y < rating.Y+rating.Height; y++ {
			for x := x0; x < x1; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if c.R > 240 && c.G < 15 && c.B < 15 {
					n++
				}
			}
		}
		return n
	}
	full := red(rating.X, rating.X+32)
	half := red(rating.X+34, rating.X+66)
	if full == 0 || half == 0 || half >= full {
		t.Fatal("fraction/custom color", full, half)
	}
	v.Size(16)
	for _, e := range w.snapshot() {
		if e.Role == "slider" && e.Name == "Average" && e.Height != 16 {
			t.Fatal("size did not update", e)
		}
	}
}
