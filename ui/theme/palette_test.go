package theme

import (
	"github.com/dyike/keel/ui/internal/loop"
	"image/color"
	"math"
	"testing"
)

func TestApplyPaletteAndInvalidateAllWindows(t *testing.T) {
	loop.Lock()
	defer loop.Unlock()
	original := Current()
	defer Apply(original)
	if original != Light() {
		t.Fatal("default colors differ from light preset")
	}
	material, shaper := Material, Material.Shaper
	counts := [2]int{}
	first, second := new(int), new(int)
	loop.Register(first, func() { counts[0]++ })
	loop.Register(second, func() { counts[1]++ })
	defer loop.Unregister(first)
	defer loop.Unregister(second)
	for i, palette := range []Palette{Dark(), Light()} {
		before := Revision()
		Apply(palette)
		if Current() != palette || Revision() != before+1 {
			t.Fatal("palette or cache revision did not update")
		}
		if Material != material || Material.Shaper != shaper {
			t.Fatal("theme switch replaced font resources")
		}
		if Material.Fg != palette.Text || Material.Bg != palette.Surface || Material.ContrastBg != palette.Primary || Material.ContrastFg != palette.OnColor {
			t.Fatal("Gio palette is stale")
		}
		if counts != [2]int{i + 1, i + 1} {
			t.Fatalf("windows were not both invalidated: %v", counts)
		}
	}
	custom := Dark()
	custom.Success = RGB(0xabcdef)
	Apply(custom)
	custom.Success = RGB(0)
	if Success != RGB(0xabcdef) || Dark().Success == Success {
		t.Fatal("palette does not have value semantics")
	}
}

func luminance(c color.NRGBA) float64 {
	linear := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	return .2126*linear(c.R) + .7152*linear(c.G) + .0722*linear(c.B)
}
func TestSelectedTextContrast(t *testing.T) {
	for _, p := range []Palette{Light(), Dark()} {
		a, b := luminance(p.PrimaryText), luminance(p.Highlight)
		if a < b {
			a, b = b, a
		}
		if c := (a + .05) / (b + .05); c < 4.5 {
			t.Fatalf("contrast %.2f", c)
		}
	}
}

func TestSemanticTextContrast(t *testing.T) {
	for _, p := range []Palette{Light(), Dark()} {
		for _, c := range []color.NRGBA{p.Success, p.Warning, p.Info, p.DangerText} {
			a, b := luminance(c), luminance(p.Surface)
			if a < b {
				a, b = b, a
			}
			if ratio := (a + .05) / (b + .05); ratio < 4.5 {
				t.Fatalf("semantic contrast %.2f", ratio)
			}
		}
	}
}
