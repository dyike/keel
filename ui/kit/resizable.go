package kit

import (
	"strconv"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ResizableView puts two panes side by side (or stacked with Vertical) with a
// handle between them. Drag the handle, or focus it and use the arrow keys
// (16dp a press, Home / End for the limits). The first pane keeps its size
// when the window resizes; the second takes the rest.
type ResizableView struct {
	first, second  el.View
	vertical       bool
	size           float32 // first pane, dp
	min1, min2     float32
	total, painted float32 // container and first pane size as last painted
	grab           float32
	onChange       func(float32)
}

func Resizable(first, second el.View) *ResizableView {
	return &ResizableView{first: first, second: second, size: 240, min1: 80, min2: 80}
}

// Vertical stacks the panes, first on top.
func (v *ResizableView) Vertical() *ResizableView { v.vertical = true; return v }

// Min sets the smallest sizes of the two panes in dp (80 each by default).
func (v *ResizableView) Min(first, second float32) *ResizableView {
	v.min1, v.min2 = max(first, 0), max(second, 0)
	return v
}
func (v *ResizableView) OnChange(fn func(size float32)) *ResizableView { v.onChange = fn; return v }

// Value is the first pane's size in dp.
func (v *ResizableView) Value() float32 { return v.size }

// SetValue sets the first pane's size without calling OnChange.
func (v *ResizableView) SetValue(dp float32) { v.size = v.clamp(dp) }

const handleSize = 6

func (v *ResizableView) clamp(dp float32) float32 {
	dp = max(dp, v.min1)
	if v.total > 0 {
		dp = min(dp, v.total-handleSize-v.min2)
	}
	return max(dp, 0)
}

func (v *ResizableView) set(dp float32) {
	dp = v.clamp(dp)
	if dp == v.size {
		return
	}
	v.size = dp
	if v.onChange != nil {
		v.onChange(dp)
	}
}

func (v *ResizableView) Render(cx *el.Context) el.Element {
	v.size = v.clamp(v.size)
	v.painted = v.size
	along := func(e el.DragEvent) float32 {
		if v.vertical {
			return e.Y
		}
		return e.X
	}
	cursor := pointer.CursorColResize
	if v.vertical {
		cursor = pointer.CursorRowResize
	}
	handle := el.Div().Role("separator").Name(locale.Current().Resize).Value(strconv.Itoa(int(v.size))).
		NoShrink().Bg(theme.Border).Cursor(cursor).Focusable(true).
		FocusStyle(func(s *el.Style) { s.Bg(theme.Primary).BorderColor(theme.Primary) }).
		Hover(func(s *el.Style) { s.Bg(theme.Primary) }).
		OnDrag(func(e el.DragEvent) {
			if e.Kind == el.DragStart {
				v.grab = along(e)
				return
			}
			// The handle sits where the last paint put it, so the new size
			// is that size plus how far the pointer moved inside it.
			v.set(v.painted + along(e) - v.grab)
		}).
		OnKey(func(e el.KeyEvent) bool {
			d, ok := float32(0), true
			switch key.Name(e.Name) {
			case key.NameLeftArrow, key.NameUpArrow:
				d = -16
			case key.NameRightArrow, key.NameDownArrow:
				d = 16
			case key.NameHome:
				d = -1e6
			case key.NameEnd:
				d = 1e6
			default:
				ok = false
			}
			if ok && e.State == el.KeyPress {
				v.set(v.size + d)
			}
			return ok
		})
	one := el.Div().NoShrink().Items(el.Stretch)
	two := el.Div().Grow().Items(el.Stretch)
	box := el.Div().Items(el.Stretch).Grow()
	if v.vertical {
		one.H(el.Dp(v.size))
		two.H(el.Dp(0))
		handle.H(el.Dp(handleSize))
	} else {
		box.Row()
		one.W(el.Dp(v.size))
		two.W(el.Dp(0))
		handle.W(el.Dp(handleSize))
	}
	if v.first != nil {
		one.Child(v.first.Render(cx))
	}
	if v.second != nil {
		two.Child(v.second.Render(cx))
	}
	return box.Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			if v.vertical {
				v.total = float32(gtx.Constraints.Max.Y) / px
			} else {
				v.total = float32(gtx.Constraints.Max.X) / px
			}
		}
		draw()
	}).Child(one, handle, two)
}
