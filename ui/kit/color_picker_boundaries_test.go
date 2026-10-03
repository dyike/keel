package kit

import (
	"image/color"
	"math"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

func TestColorPickerNarrowKeyboardAndOwnedSwatches(t *testing.T) {
	colors := []color.NRGBA{{R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255}, {R: 255, G: 255, A: 255}, {R: 255, B: 255, A: 255}, {G: 255, B: 255, A: 255}}
	p := ColorPicker().Alpha().Swatches(colors...)
	colors[0] = color.NRGBA{A: 255}
	for _, scale := range []int{1, 2} {
		h := renderView(p, 160, scale)
		for _, name := range []string{"饱和度与亮度", "色相", "不透明度", "#FF0000FF", "#00FFFFFF"} {
			b := bounds(h, name)
			if b.Dx() == 0 || b.Min.X < 0 || b.Max.X > 160*scale {
				t.Fatalf("narrow %s: %v", name, b)
			}
		}
		click(t, h, "色相")
		h.Key(key.NameHome, 0)
		h.Key(key.NameRightArrow, key.ModShift)
		if p.h != 20 {
			t.Fatal("hue keyboard", p.h)
		}
		h.Key(key.NameUpArrow, 0)
		if p.h != 22 {
			t.Fatal("vertical arrow", p.h)
		}
		click(t, h, "不透明度")
		h.Key(key.NameHome, 0)
		if p.a != 0 {
			t.Fatal("opacity home")
		}
		h.Key(key.NameEnd, 0)
		if p.a != 1 {
			t.Fatal("opacity end")
		}
	}
}
func TestColorPickerCancelAndNonFiniteInput(t *testing.T) {
	p := ColorPicker()
	h := page(p)
	b := bounds(h, "饱和度与亮度")
	at := f32.Pt(float32(b.Min.X+40), float32(b.Min.Y+50))
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: at})
	h.Frame()
	before := p.Value()
	h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
	h.Frame()
	if p.Value() != before {
		t.Fatal("cancel coordinates changed color")
	}
	p.set(math.NaN(), 1, 1, 1)
	if p.Value() != before {
		t.Fatal("nonfinite color accepted")
	}
	if contrastingText(color.NRGBA{R: 255, G: 255, B: 255, A: 255}) != (color.NRGBA{A: 255}) {
		t.Fatal("light check contrast")
	}
	if contrastingText(color.NRGBA{A: 255}) != (color.NRGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatal("dark check contrast")
	}
}
