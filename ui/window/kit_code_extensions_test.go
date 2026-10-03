package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestKitCodeDecorationsPixelsAndSearch(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	for _, dark := range []bool{false, true} {
		palette := theme.Light()
		if dark {
			palette = theme.Dark()
		}
		core.Update(func() { theme.Apply(palette) })
		ed := kit.CodeEditor("filled text\nframed text\ncolored text\nunderlined text\nsearch target").Name("annotated source")
		colors := []color.NRGBA{{R: 221, G: 33, B: 66, A: 255}, {R: 33, G: 199, B: 88, A: 255}, {R: 41, G: 77, B: 231, A: 255}, {R: 199, G: 77, B: 201, A: 255}}
		styles := []kit.CodeDecorationStyle{kit.CodeDecorationFill, kit.CodeDecorationFrame, kit.CodeDecorationText, kit.CodeDecorationUnderline}
		var ds []kit.CodeDecoration
		for i, style := range styles {
			ds = append(ds, kit.CodeDecoration{Range: kit.CodeRange{Line: i, Col: 0, EndLine: i, EndCol: 8}, Style: style, Color: &colors[i]})
		}
		owner := ed.Decorations(ds...)
		w := openTest(t, Options{Width: 600, Height: 420, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(16).Items(el.Stretch).Child(ed.Render(cx)) }))})
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
		counts := make([]int, len(colors))
		for y := 0; y < im.Bounds().Dy(); y++ {
			for x := 0; x < im.Bounds().Dx(); x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				r, g, b := int(c.R), int(c.G), int(c.B)
				switch {
				case r > g+30 && r > b+30:
					counts[0]++
				case g > r+30 && g > b+30:
					counts[1]++
				case b > r+30 && b > g+30:
					counts[2]++
				case r > g+30 && b > g+30:
					counts[3]++
				}
			}
		}
		for i, n := range counts {
			if n < 5 {
				t.Fatalf("dark %t style %d missing pixels: %v", dark, styles[i], counts)
			}
		}
		e := element(t, w, "annotated source")
		if e.Value != ed.Value() {
			t.Fatal("decoration changed accessible text")
		}
		core.Update(func() {
			ed.SetSearchQuery("target", kit.CodeSearchOptions{})
			ed.Searchable(false)
			ed.SelectSearchMatch(0)
		})
		w.render()
		if ed.Selection() != "target" || ed.SearchSession().PanelOpen {
			t.Fatal("custom search")
		}
		core.Update(func() { owner.Dispose(); ed.CloseSearch() })
		w.render()
		w.click(element(t, w, "annotated source").center())
		if err := w.typeText("x"); err != nil {
			t.Fatal(err)
		}
		w.snapshot()
		if ed.Value() == e.Value {
			t.Fatal("annotation intercepted input")
		}
	}
}
