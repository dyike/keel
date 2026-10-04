package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"math"
	"testing"
)

func TestSliderAppearanceDragGeometry(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			releases := 0
			v := Slider("Custom", 0, 100).Step(1).Appearance(func(a *SliderAppearance) { a.ThumbSize = 32; a.TrackSize = 10 }).OnRelease(func(float64) { releases++ })
			if vertical {
				v.Vertical(232)
			}
			h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(232)).Child(v.Render(cx)) }), 300, scale)
			n, ok := semanticNode(h, "slider:0")
			if !ok {
				t.Fatal("missing slider")
			}
			b := n.Desc.Bounds
			// For single sliders the accessible bounds describe the full track.
			x, y := float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2)
			h.Click(x, y)
			if v.Value() != 50 || releases != 1 {
				t.Fatalf("midpoint scale=%d vertical=%v: %v %d bounds %v", scale, vertical, v.Value(), releases, b)
			}
			if vertical {
				h.Drag(x, y, x, float32(b.Min.Y+16*scale))
			} else {
				h.Drag(x, y, float32(b.Max.X-16*scale), y)
			}
			if v.Value() != 100 || releases != 2 {
				t.Fatal("custom thumb endpoint", v.Value(), releases)
			}
			h.Key(key.NameHome, 0)
			if v.Value() != 0 {
				t.Fatal("keyboard endpoint")
			}
			v.SetDisabled(true)
			h.Frame()
			h.Click(x, y)
			if v.Value() != 0 {
				t.Fatal("disabled custom slider")
			}
		}
	}
}

func TestRangeSliderAppearanceGeometryAndReset(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, vertical := range []bool{false, true} {
			v := RangeSlider("Range", 0, 100).Step(1).Appearance(func(a *SliderAppearance) { a.ThumbSize = 32 })
			if vertical {
				v.Vertical(232)
			}
			v.SetValues(20, 80)
			h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(232)).Child(v.Render(cx)) }), 300, scale)
			b := bounds(h, "上限 Range")
			if b.Dx() != 32*scale || b.Dy() != 32*scale {
				t.Fatal("thumb size", b)
			}
			x, y := center(b)
			if vertical {
				h.Drag(x, y, x, y+120*float32(scale))
			} else {
				h.Drag(x, y, x-120*float32(scale), y)
			}
			if lo, hi := v.Values(); lo != 20 || hi != 20 {
				t.Fatal("range endpoint mapping", lo, hi)
			}
			v.Appearance(nil)
			h.Frame()
			b = bounds(h, "上限 Range")
			if b.Dx() != 16*scale {
				t.Fatal("appearance reset", b)
			}
			if lo, hi := v.Values(); lo != 20 || hi != 20 {
				t.Fatal("style reset changed values")
			}
		}
	}
}

func TestSliderAppearanceValidationAndTheme(t *testing.T) {
	old := theme.Current()
	defer theme.Apply(old)
	v := Slider("", 0, 100).Appearance(func(a *SliderAppearance) {
		a.TrackSize = float32(math.NaN())
		a.ThumbSize = -1
		a.ThumbBorderWidth = 100
		a.TrackRadius = float32(math.Inf(1))
		a.ThumbRadius = -1
	})
	a := v.resolveAppearance()
	if a.TrackSize != 4 || a.ThumbSize != 16 || a.ThumbBorderWidth != 8 || a.ThumbRadius != theme.RadiusLg {
		t.Fatal(a)
	}
	v.Appearance(nil)
	theme.Apply(theme.Dark())
	a = v.resolveAppearance()
	if a.TrackColor != theme.Border || a.ThumbColor != theme.Surface {
		t.Fatal("stale theme colors")
	}
}
