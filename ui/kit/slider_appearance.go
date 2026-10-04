package kit

import (
	"image/color"

	"github.com/dyike/keel/ui/theme"
)

// SliderAppearance describes the track and both thumbs. Dimensions are dp.
// A callback receives current theme defaults on each render.
type SliderAppearance struct {
	TrackColor, FillColor, ThumbColor, ThumbBorderColor color.NRGBA
	TrackSize, ThumbSize, ThumbBorderWidth              float32
	TrackRadius, ThumbRadius                            float32
}

// Appearance customizes this slider without changing its value or callbacks.
// Nil restores theme defaults. The callback runs during Render and should only
// edit the supplied configuration; both range endpoints share this appearance.
func (v *SliderView) Appearance(fn func(*SliderAppearance)) *SliderView {
	v.appearance = fn
	return v
}

func (v *SliderView) resolveAppearance() SliderAppearance {
	a := SliderAppearance{
		TrackColor: theme.Border, FillColor: theme.Primary,
		ThumbColor: theme.Surface, ThumbBorderColor: theme.Primary,
		TrackSize: 4, ThumbSize: 16, ThumbBorderWidth: 2,
		TrackRadius: theme.RadiusFull, ThumbRadius: theme.RadiusLg,
	}
	defaults := a
	if v.appearance != nil {
		v.appearance(&a)
	}
	for _, p := range []struct {
		value    *float32
		fallback float32
		positive bool
	}{
		{&a.TrackSize, defaults.TrackSize, true}, {&a.ThumbSize, defaults.ThumbSize, true},
		{&a.ThumbBorderWidth, defaults.ThumbBorderWidth, false},
		{&a.TrackRadius, defaults.TrackRadius, false}, {&a.ThumbRadius, defaults.ThumbRadius, false},
	} {
		if !finiteNumber(float64(*p.value)) || *p.value < 0 || p.positive && *p.value == 0 {
			*p.value = p.fallback
		}
	}
	// Bound arithmetic and keep the border inside the thumb.
	a.TrackSize = min(a.TrackSize, 1024)
	a.ThumbSize = min(a.ThumbSize, 1024)
	a.ThumbBorderWidth = min(a.ThumbBorderWidth, a.ThumbSize/2)
	a.TrackRadius = min(a.TrackRadius, 1024)
	a.ThumbRadius = min(a.ThumbRadius, a.ThumbSize/2)
	if v.disabled {
		a.FillColor, a.ThumbBorderColor = theme.Muted, theme.Muted
	}
	return a
}
