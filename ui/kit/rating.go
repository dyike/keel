package kit

import (
	"gioui.org/op/clip"
	"github.com/dyike/keel/ui/core"
	"image"
	"math"
	"strconv"

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
}

// Rating creates a rating out of max stars (5 if max < 1).
func Rating(label string, max int) *RatingView {
	if max < 1 {
		max = 5
	}
	return &RatingView{label: label, max: max}
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
			shown = float64(i) // preview the rating under the pointer
		}
	}
	name := v.label
	if name == "" {
		name = v.name
	}
	row := el.Div().ID(id).Role("slider").Name(name).Value(strconv.FormatFloat(v.value, 'f', -1, 64) + "/" + strconv.Itoa(v.max)).
		Row().Gap(2).Rounded(4).Disabled(v.disabled).Focusable(interactive).
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
	for i := 1; i <= v.max; i++ {
		i := i
		portion := float32(min(1, max(0, shown-float64(i-1))))
		star := el.Div().ID(id + "/" + strconv.Itoa(i)).Size(el.Dp(22)).Child(Icon(IconStarOutline).Size(22).Color(theme.Muted).Render(cx))
		if portion > 0 {
			filled := el.Div().Absolute().Top(0).Left(0).Size(el.Dp(22)).Child(Icon(IconStar).Size(22).Color(theme.Warning).Render(cx))
			filled.Decorate(func(gtx core.C, draw func()) {
				bounds := gtx.Constraints.Max
				bounds.X = int(math.Round(float64(bounds.X) * float64(portion)))
				defer clip.Rect(image.Rectangle{Max: bounds}).Push(gtx.Ops).Pop()
				draw()
			})
			star.Child(filled)
		}
		if interactive {
			star.CursorPointer().Focusable(false).OnClick(func() { v.set(i) })
		}
		row.Child(star)
	}
	return labelled(v.label, el.Div().Items(el.Start).Child(row), "")
}

func (v *RatingView) setName(s string) { v.name = s }
