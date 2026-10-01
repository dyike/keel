package kit

import (
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
	value, max         int
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
func (v *RatingView) Value() int                        { return v.value }
func (v *RatingView) SetValue(n int)                    { v.value = min(max(n, 0), v.max) }
func (v *RatingView) SetDisabled(on bool)               { v.disabled = on }

func (v *RatingView) set(n int) {
	n = min(max(n, 0), v.max)
	if n == v.value || v.readOnly || v.disabled {
		return
	}
	v.value = n
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
			shown = i // preview the rating under the pointer
		}
	}
	name := v.label
	if name == "" {
		name = v.name
	}
	row := el.Div().ID(id).Role("slider").Name(name).Value(strconv.Itoa(v.value) + "/" + strconv.Itoa(v.max)).
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
				v.set(v.value + d)
			}
			return true
		})
	for i := 1; i <= v.max; i++ {
		i := i
		icon, color := IconStarOutline, theme.Muted
		if i <= shown {
			icon, color = IconStar, theme.Warning
		}
		star := el.Div().ID(id + "/" + strconv.Itoa(i)).Child(Icon(icon).Size(22).Color(color).Render(cx))
		if interactive {
			star.CursorPointer().Focusable(false).OnClick(func() { v.set(i) })
		}
		row.Child(star)
	}
	return labelled(v.label, el.Div().Items(el.Start).Child(row), "")
}

func (v *RatingView) setName(s string) { v.name = s }
