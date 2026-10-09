package kit

import (
	"image/color"
	"testing"

	"gioui.org/io/key"

	"github.com/dyike/keel/ui/el"
)

func TestFormatColorAndHSL(t *testing.T) {
	c := color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xcc}
	for f, want := range map[ColorFormat]string{ColorHex: "#2563EB", ColorRGB: "rgb(37, 99, 235)", ColorHSL: "hsl(221, 83%, 53%)"} {
		if got := FormatColor(c, f, false); got != want {
			t.Errorf("format %d = %q, want %q", f, got, want)
		}
	}
	if got := FormatColor(c, ColorRGB, true); got != "rgba(37, 99, 235, 0.8)" {
		t.Errorf("rgba = %q", got)
	}
	// HSL round-trips every primary and a few mixes exactly.
	for _, x := range []color.NRGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}, {128, 128, 128, 255}, {37, 99, 235, 255}, {250, 200, 10, 255}} {
		h, s, l := rgbToHSL(x.R, x.G, x.B)
		r, g, b := hslToRGB(h, s, l)
		if r != x.R || g != x.G || b != x.B {
			t.Errorf("round trip %v -> %d %d %d", x, r, g, b)
		}
	}
}

// Typing in the RGB and HSL fields sets the color; switching format keeps it;
// a popup trigger shows the value in the chosen format.
func TestColorPickerFormatFields(t *testing.T) {
	var got []color.NRGBA
	p := ColorPicker().Format(ColorRGB).OnChange(func(c color.NRGBA) { got = append(got, c) })
	h := renderView(p, 320, 1)
	click(t, h, "G")
	clearField(h)
	h.Type("200")
	h.Key(key.NameReturn, 0)
	h.Frame()
	if v := p.Value(); v.G != 200 || v.R != 0x25 || len(got) != 1 {
		t.Fatalf("RGB entry: %v, %d changes", v, len(got))
	}
	click(t, h, "HSL 颜色格式")
	h.Frame()
	if p.CurrentFormat() != ColorHSL || p.parts[0] == "" {
		t.Fatal("format switch")
	}
	before := p.Value()
	click(t, h, "L")
	clearField(h)
	h.Type("90")
	h.Key(key.NameReturn, 0)
	h.Frame()
	if v := p.Value(); v == before || int(v.R)+int(v.G)+int(v.B) < 600 {
		t.Fatalf("HSL lightness 90%% did not lighten: %v", v)
	}
	click(t, h, "H")
	clearField(h)
	h.Type("abc")
	h.Key(key.NameReturn, 0)
	if p.parts[0] == "" {
		t.Fatal("non-digits were not filtered or restored")
	}

	popup := ColorPicker().Popup(true).Format(ColorRGB).Placement(el.Top, el.End)
	popup.SetValue(color.NRGBA{R: 1, G: 2, B: 3, A: 255})
	ph := page(popup)
	ph.Frame()
	if _, ok := node(ph, "rgb(1, 2, 3)"); !ok {
		t.Fatal("trigger does not show the RGB text")
	}
	if popup.popover.side != el.Top || popup.popover.align != el.End {
		t.Fatal("placement not passed to the popup")
	}
}

// clearField empties the focused field: to the end, then backspace.
func clearField(h interface {
	Key(key.Name, key.Modifiers)
}) {
	h.Key(key.NameEnd, 0)
	for range 4 {
		h.Key(key.NameDeleteBackward, 0)
	}
}
