package theme_test

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func TestSystemEmojiFallbackRendersColor(t *testing.T) {
	for _, face := range []struct {
		name string
		mono bool
	}{{"text", false}, {"monospace", true}} {
		t.Run(face.name, func(t *testing.T) {
			label := el.Text("👋 👋🏽 👨‍👩‍👧‍👦").TextSize(24)
			if face.mono {
				label.Mono()
			}
			root := el.Root(el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().P(8).Bg(theme.RGB(0xffffff)).Child(label)
			}))
			file := filepath.Join(t.TempDir(), "emoji.png")
			if err := window.ScreenshotAtScale(root, 240, 60, 2, file); err != nil {
				t.Fatal(err)
			}
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			img, err := png.Decode(f)
			if err != nil {
				t.Fatal(err)
			}
			colored := 0
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
					r, g, b, _ := img.At(x, y).RGBA()
					if r > 30000 && g > b+7000 && r > b+7000 {
						colored++
					}
				}
			}
			if colored < 100 {
				t.Fatalf("emoji fell back to missing/monochrome glyphs: only %d color pixels", colored)
			}
		})
	}
}
