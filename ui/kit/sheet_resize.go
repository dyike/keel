package kit

import (
	"strconv"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// Resizable enables the inner-edge handle, enabled by default.
func (v *SheetView) Resizable(on bool) *SheetView {
	v.resizeOff = !on
	if !on {
		v.resizing = false
	}
	return v
}

// OnResize reports user size changes in dp; Size does not invoke it.
func (v *SheetView) OnResize(fn func(float32)) *SheetView { v.onResize = fn; return v }

// PanelSize returns the requested width or height in dp, including user resizing.
func (v *SheetView) PanelSize() float32 { return v.size }

func (v *SheetView) resizeTo(size float32) {
	if !v.open || v.disabled || v.resizeOff {
		return
	}
	size = min(max(size, min(float32(80), v.resizeLimit)), v.resizeLimit)
	if size == v.size {
		return
	}
	v.size = size
	if v.onResize != nil {
		v.onResize(size)
	}
}

func (v *SheetView) resizeHandle(id string, cx *el.Context) el.Element {
	horizontal := v.side == el.Left || v.side == el.Right
	w, h := cx.ViewportSize()
	v.resizeLimit = w
	if !horizontal {
		v.resizeLimit = max(0, h-v.marginTop)
	}
	sign := float32(1)
	if v.side == el.Right || v.side == el.Bottom {
		sign = -1
	}
	cursor := pointer.CursorColResize
	if !horizontal {
		cursor = pointer.CursorRowResize
	}
	handle := el.Div().ID(id + "/resize").Absolute().Role("separator").Name(locale.Current().Resize).
		Value(strconv.FormatFloat(float64(v.size), 'f', -1, 32)).Focusable(true).Cursor(cursor).
		Hover(func(s *el.Style) { s.Bg(theme.Primary) }).
		FocusStyle(func(s *el.Style) { s.Bg(theme.Primary) }).
		OnDrag(func(e el.DragEvent) {
			along := e.X
			if !horizontal {
				along = e.Y
			}
			if e.Kind == el.DragStart {
				v.resizing = true
				v.resizeGrab = along
				v.resizeStart = v.size
				return
			}
			if !v.resizing {
				return
			}
			if e.Canceled {
				v.resizing = false
				v.resizeTo(v.resizeStart)
				return
			}
			v.resizeTo(v.resizePainted + sign*(along-v.resizeGrab))
			if e.Kind == el.DragEnd {
				v.resizing = false
			}
		}).OnKey(func(e el.KeyEvent) bool {
		size := v.resizePainted
		switch key.Name(e.Name) {
		case key.NameHome:
			size = 0
		case key.NameEnd:
			size = v.resizeLimit
		case key.NameLeftArrow:
			if !horizontal {
				return false
			}
			size -= 16 * sign
		case key.NameRightArrow:
			if !horizontal {
				return false
			}
			size += 16 * sign
		case key.NameUpArrow:
			if horizontal {
				return false
			}
			size -= 16 * sign
		case key.NameDownArrow:
			if horizontal {
				return false
			}
			size += 16 * sign
		default:
			return false
		}
		if e.State == el.KeyPress {
			v.resizeTo(size)
		}
		return true
	})
	switch v.side {
	case el.Left:
		handle.Right(0).Top(0).Bottom(0).W(el.Dp(handleSize))
	case el.Right:
		handle.Left(0).Top(0).Bottom(0).W(el.Dp(handleSize))
	case el.Top:
		handle.Bottom(0).Left(0).Right(0).H(el.Dp(handleSize))
	case el.Bottom:
		handle.Top(0).Left(0).Right(0).H(el.Dp(handleSize))
	}
	return handle
}
