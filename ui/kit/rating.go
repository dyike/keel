package kit

import (
	"image"
	"image/color"
	"math"
	"strconv"

	"gioui.org/op/clip"
	"github.com/dyike/keel/ui/core"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// RatingView picks a whole number of stars from 0 to max. Click a star, or
// use ← → (Home / End for the ends) once it has focus.
type RatingView struct {
	name               string // accessible name from a Form row when label is empty
	label              string
	value              float64
	max                int
	readOnly, disabled bool
	onChange           func(int)
	size               float32
	color              *color.NRGBA
}

// Rating creates a rating out of max stars (5 if max < 1).
func Rating(label string, max int) *RatingView {
	if max < 1 {
		max = 5
	}
	return &RatingView{label: label, max: max}
}

// Size sets each star's diameter in dp (8–128); invalid values are ignored.
func (v *RatingView) Size(dp float32) *RatingView {
	if dp >= 8 && dp <= 128 {
		v.size = dp
	}
	return v
}

// Color sets the filled star color; outlines continue to use theme.Muted.
func (v *RatingView) Color(c color.NRGBA) *RatingView { v.color = &c; return v }
func (v *RatingView) clickValue(i int) int {
	if v.value >= float64(i) {
		return i - 1
	}
	return i
}
func (v *RatingView) ReadOnly() *RatingView             { v.readOnly = true; return v }
func (v *RatingView) OnChange(fn func(int)) *RatingView { v.onChange = fn; return v }
func (v *RatingView) Value() int                        { return int(v.value) }
func (v *RatingView) SetValue(n int)                    { v.SetScore(float64(n)) }
func (v *RatingView) SetDisabled(on bool)               { v.disabled = on }

// Score returns the exact value, including fractions used for display.
func (v *RatingView) Score() float64 { return v.value }

// SetScore sets a fractional score without callbacks. User editing remains
// whole-star; use ReadOnly for averages. NaN becomes zero, infinities clamp.
func (v *RatingView) SetScore(n float64) {
	if math.IsNaN(n) {
		n = 0
	}
	v.value = min(float64(v.max), max(0, n))
}

func (v *RatingView) set(n int) {
	n = min(max(n, 0), v.max)
	if float64(n) == v.value || v.readOnly || v.disabled {
		return
	}
	v.value = float64(n)
	if v.onChange != nil {
		v.onChange(n)
	}
}

func (v *RatingView) Render(cx *el.Context) el.Element {
	id := autoID("rating", v)
	interactive := !v.readOnly && !v.disabled
	shown := v.value
	for i := 1; i <= v.max && interactive; i++ {
		if cx.Hovered(id + "/" + strconv.Itoa(i)) {
			shown = float64(v.clickValue(i)) // preview the score this click would choose
		}
	}
	name := v.label
	if name == "" {
		name = v.name
	}
	row := el.Div().ID(id).Role("slider").Name(name).Value(strconv.FormatFloat(v.value, 'f', -1, 64) + "/" + strconv.Itoa(v.max)).
		Row().Gap(theme.SpaceXxs).Rounded(theme.RadiusSm).Disabled(v.disabled).Focusable(interactive).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			d := 0
			switch key.Name(e.Name) {
			case key.NameLeftArrow, key.NameDownArrow:
				d = -1
			case key.NameRightArrow, key.NameUpArrow:
				d = 1
			case key.NameHome:
				d = -v.max
			case key.NameEnd:
				d = v.max
			default:
				return false
			}
			if e.State == el.KeyPress {
				v.set(int(v.value) + d)
			}
			return true
		})
	size := v.size
	if size == 0 {
		size = 22
	}
	active := theme.Warning
	if v.color != nil {
		active = *v.color
	}
	for i := 1; i <= v.max; i++ {
		i := i
		portion := float32(min(1, max(0, shown-float64(i-1))))
		star := el.Div().ID(id + "/" + strconv.Itoa(i)).Size(el.Dp(size)).Child(Icon(IconStarOutline).Size(size).Color(theme.Muted).Render(cx))
		if portion > 0 {
			filled := el.Div().Absolute().Top(0).Left(0).Size(el.Dp(size)).Child(Icon(IconStar).Size(size).Color(active).Render(cx))
			filled.Decorate(func(gtx core.C, draw func()) {
				bounds := gtx.Constraints.Max
				bounds.X = int(math.Round(float64(bounds.X) * float64(portion)))
				defer clip.Rect(image.Rectangle{Max: bounds}).Push(gtx.Ops).Pop()
				draw()
			})
			star.Child(filled)
		}
		if interactive {
			star.CursorPointer().Focusable(false).OnClick(func() { v.set(v.clickValue(i)) })
		}
		row.Child(star)
	}
	return labelled(v.label, el.Div().Items(el.Start).Child(row), "")
}

func (v *RatingView) setName(s string) { v.name = s }

func (v *RatingView) FocusID() string {
	if v.readOnly {
		return ""
	}
	return autoID("rating", v)
}
