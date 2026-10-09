package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"gioui.org/font"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestKitCodeDecorationWeightPixels(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	core.Update(func() { theme.Apply(theme.Light()) })
	ed := kit.CodeEditor("MMMMMMMMMMMM\nMMMMMMMMMMMM").Name("weighted source")
	red, green := color.NRGBA{R: 220, G: 30, B: 30, A: 255}, color.NRGBA{R: 30, G: 200, B: 30, A: 255}
	ed.Decorations(
		kit.CodeDecoration{Range: kit.CodeRange{Line: 0, EndLine: 0, EndCol: 12}, Style: kit.CodeDecorationText, Color: &red, Weight: font.Bold},
		kit.CodeDecoration{Range: kit.CodeRange{Line: 1, EndLine: 1, EndCol: 12}, Style: kit.CodeDecorationText, Color: &green},
	)
	w := openTest(t, Options{Width: 500, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(16).Items(el.Stretch).Child(ed.Render(cx)) }))})
	w.render()
	w.render()
	shot, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(shot))
	if err != nil {
		t.Fatal(err)
	}
	var bold, regular int
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
			r, g, b := int(c.R), int(c.G), int(c.B)
			switch {
			case r > g+60 && r > b+60:
				bold++
			case g > r+60 && g > b+60:
				regular++
			}
		}
	}
	if regular < 20 || bold < regular*11/10 {
		t.Fatalf("bold glyphs should cover more pixels: bold %d, regular %d", bold, regular)
	}
}
