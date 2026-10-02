package el

import (
	"image"
	"math"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// scrollbarState uses the painted thumb geometry for both drawing and hit
// testing, including minimum-size thumbs in small viewports.
type scrollbarState struct {
	drag            gesture.Drag
	hover           gesture.Hover
	active, hovered bool
	grab            float32
}

func scrollbarThumb(track image.Rectangle, horizontal bool, offset, view, total, minimum int) image.Rectangle {
	length := mainOf(track.Size(), horizontal)
	if length <= 0 || total <= view || view <= 0 {
		return image.Rectangle{}
	}
	// Leave travel even when the viewport is shorter than the minimum thumb.
	size := min(length-1, max(length*view/total, min(minimum, max(length/2, 1))))
	start := (length - size) * min(max(offset, 0), total-view) / (total - view)
	thumb := track
	if horizontal {
		thumb.Min.X += start
		thumb.Max.X = thumb.Min.X + size
	} else {
		thumb.Min.Y += start
		thumb.Max.Y = thumb.Min.Y + size
	}
	return thumb
}

func (s *scrollbarState) update(gtx core.C, track image.Rectangle, horizontal bool, offset, view, total, minimum int, focus *elemState) int {
	s.hovered = s.hover.Update(gtx.Source)
	axis := gesture.Vertical
	if horizontal {
		axis = gesture.Horizontal
	}
	for {
		ev, ok := s.drag.Update(gtx.Metric, gtx.Source, axis)
		if !ok {
			break
		}
		thumb := scrollbarThumb(track, horizontal, offset, view, total, minimum)
		if thumb.Empty() {
			s.active = false
			continue
		}
		pos := ev.Position.Y
		if horizontal {
			pos = ev.Position.X
		}
		start := float32(mainOf(thumb.Min, horizontal))
		size := float32(mainOf(thumb.Size(), horizontal))
		switch ev.Kind {
		case pointer.Press:
			if thumb.Empty() {
				continue
			}
			s.active = true
			s.grab = pos - start
			if pos < start || pos >= start+size {
				s.grab = size / 2
			}
			if focus.focusable {
				focus.pointerFocus = true
				gtx.Execute(key.FocusCmd{Tag: focus})
			}
			if pos >= start && pos < start+size {
				continue
			}
		case pointer.Drag:
			if !s.active {
				continue
			}
		case pointer.Release, pointer.Cancel:
			s.active = false
			continue
		default:
			continue
		}
		travel := mainOf(track.Size(), horizontal) - int(size)
		if travel > 0 {
			offset = min(max(int(math.Round(float64((pos-float32(mainOf(track.Min, horizontal))-s.grab)*float32(total-view)/float32(travel)))), 0), total-view)
		}
	}
	if track.Empty() || total <= view {
		s.active = false
	}
	return offset
}

func (s *scrollbarState) paint(gtx core.C, track image.Rectangle, horizontal bool, offset, view, total, minimum int) {
	thumb := scrollbarThumb(track, horizontal, offset, view, total, minimum)
	if thumb.Empty() {
		return
	}
	defer clip.Rect(track).Push(gtx.Ops).Pop()
	s.drag.Add(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	hover := clip.Rect(track).Push(gtx.Ops)
	s.hover.Add(gtx.Ops)
	hover.Pop()
	pass.Pop()
	if s.hovered || s.active {
		paint.FillShape(gtx.Ops, theme.Subtle, clip.Rect(track).Op())
	}
	// The bar overlays content, so at rest it is a thin translucent line that
	// keeps the text beneath readable; hovering or dragging widens it.
	cross := mainOf(track.Size(), !horizontal)
	inset := max(1, cross*3/8)
	color := theme.Muted
	color.A = 0x80
	if s.hovered || s.active {
		inset = max(1, cross/4)
		color.A = 0xff
	}
	if s.active {
		color = theme.Primary
	}
	if horizontal {
		thumb.Min.Y += inset
		thumb.Max.Y -= inset
	} else {
		thumb.Min.X += inset
		thumb.Max.X -= inset
	}
	r := max(1, mainOf(thumb.Size(), !horizontal)/2)
	paint.FillShape(gtx.Ops, color, clip.UniformRRect(thumb, r).Op(gtx.Ops))
}

func (st *elemState) scrollKey(gtx core.C, ev key.Event) bool {
	if ev.Modifiers & ^key.ModShift != 0 {
		return false
	}
	horizontal := ev.Name == key.NameLeftArrow || ev.Name == key.NameRightArrow || ev.Modifiers.Contain(key.ModShift) || !st.scrollableY
	if horizontal && !st.scrollableX || !horizontal && !st.scrollableY {
		return false
	}
	offset, view, total := st.scrollY+st.scrollPending, st.scrollView, st.scrollContent
	if horizontal {
		offset, view, total = st.scrollX+st.scrollPendingX, st.scrollViewX, st.scrollContentX
	}
	target := offset
	step := gtx.Dp(40)
	switch ev.Name {
	case key.NameLeftArrow, key.NameUpArrow:
		target -= step
	case key.NameRightArrow, key.NameDownArrow:
		target += step
	case key.NamePageUp:
		target -= view
	case key.NamePageDown:
		target += view
	case key.NameHome:
		target = 0
	case key.NameEnd:
		target = total - view
	default:
		return false
	}
	target = min(max(target, 0), max(total-view, 0))
	if target == offset {
		return false
	}
	if ev.State == key.Press {
		if horizontal {
			st.scrollPendingX += target - offset
		} else {
			st.scrollPending += target - offset
		}
	}
	return true
}
