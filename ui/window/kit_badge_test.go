package window

import (
	"bytes"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestKitBadgeIconCustomColorSnapshot(t *testing.T) {
	yellow := color.NRGBA{R: 255, G: 255, A: 255}
	v := kit.Badge(0).Icon(kit.IconCheck).Name("Verified").Size(24).Color(yellow)
	w := openTest(t, kitPage(v))
	e := element(t, w, "Verified")
	if e.Role != "badge" || e.Value != "icon" || e.Width != 24 || e.Height != 24 {
		t.Fatalf("icon semantics %+v", e)
	}
	data, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	yellowCount, blackCount := 0, 0
	for y := e.Y; y < e.Y+e.Height; y++ {
		for x := e.X; x < e.X+e.Width; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if c.R > 248 && c.G > 248 && c.B < 8 {
				yellowCount++
			}
			if c == (color.NRGBA{A: 255}) {
				blackCount++
			}
		}
	}
	if yellowCount < 100 || blackCount < 5 {
		t.Fatalf("custom fill/contrast missing: %d %d sample=%v box=%+v", yellowCount, blackCount, color.NRGBAModel.Convert(img.At(e.X+5, e.Y+5)), e)
	}
	v.Icon(kit.IconNone).Name("")
	v.SetValue(150)
	if e := element(t, w, "150"); e.Value != "99+" {
		t.Fatal("count restore")
	}
	v.SetValue(0)
	if roleOfName(w, "150") != "" {
		t.Fatal("zero count visible")
	}
}
