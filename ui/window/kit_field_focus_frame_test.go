package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestFieldFocusUsesOuterFrameOnly(t *testing.T) {
	for _, view := range []el.View{kit.DatePicker("Date"), kit.Select("Choice", "A", "B")} {
		w := openTest(t, kitPage(view))
		if err := w.press("Tab"); err != nil {
			t.Fatal(err)
		}
		w.render()
		name := "Date"
		if _, ok := view.(*kit.SelectView); ok {
			name = "Choice"
		}
		field := element(t, w, name)
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		// The empty area across the inner trigger's top edge must not form
		// another primary-colored focus rectangle inside the field frame.
		primary := 0
		for x := field.X + field.Width/4; x < field.X+field.Width/2; x++ {
			c := color.NRGBAModel.Convert(im.At(x, field.Y)).(color.NRGBA)
			if c == theme.Primary {
				primary++
			}
		}
		if primary > 2 {
			t.Fatalf("%s has a nested focus border (%d pixels)", name, primary)
		}
	}
}
