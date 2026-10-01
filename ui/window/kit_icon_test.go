package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/png"
	"testing"
)

func TestKitIconThemePixels(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	w := openTest(t, Options{Width: 120, Height: 120, Content: el.Embed(kit.Icon(kit.IconInfo).Size(32))})
	for _, p := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(p) })
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for y := 0; y < im.Bounds().Max.Y; y++ {
			for x := 0; x < im.Bounds().Max.X; x++ {
				r, g, b, _ := im.At(x, y).RGBA()
				if uint8(r>>8) == p.Text.R && uint8(g>>8) == p.Text.G && uint8(b>>8) == p.Text.B {
					count++
				}
			}
		}
		if count < 20 {
			t.Fatalf("icon does not follow palette: %d", count)
		}
		for _, e := range w.snapshot() {
			if e.Role == "image" || e.Role == "button" {
				t.Fatal("decorative icon exposed")
			}
		}
	}
}
