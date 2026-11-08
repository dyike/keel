// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !nox11

package app

import (
	"image"
	"unsafe"

	"github.com/dyike/keel/gio/io/system"
)

// XChangeProperty(format=32) consumes an array of native Xlib longs, including
// on LP64. The decorations flag must be present even when its value is zero.
func x11DecorationHints(decorated bool) [5]uintptr {
	hints := [5]uintptr{2, 0, 0, 0, 0}
	if decorated {
		hints[2] = 1
	}
	return hints
}

func x11MoveResizeDirection(size image.Point, edge int, at image.Point, action system.Action, windowed bool) (int, bool) {
	if at.X < 0 || at.Y < 0 || at.X >= size.X || at.Y >= size.Y {
		return 0, false
	}
	if windowed {
		left, right := at.X < edge, at.X >= size.X-edge
		top, bottom := at.Y < edge, at.Y >= size.Y-edge
		switch {
		case top && left:
			return 0, true
		case top && right:
			return 2, true
		case bottom && right:
			return 4, true
		case bottom && left:
			return 6, true
		case top:
			return 1, true
		case right:
			return 3, true
		case bottom:
			return 5, true
		case left:
			return 7, true
		}
	}
	if action == system.ActionMove {
		return 8, true
	}
	return 0, false
}

func (w *x11Window) beginMoveResize(button *xButtonEvent, direction int) {
	rootPosition := image.Pt(int(button.XRoot), int(button.YRoot))
	if direction == 8 {
		delta := rootPosition.Sub(w.titleClickPosition)
		if w.titleClickTime != 0 && button.Time-w.titleClickTime < 400 && delta.X*delta.X+delta.Y*delta.Y < 25 {
			w.titleClickTime = 0
			mode := Maximized
			if w.config.Mode == Maximized {
				mode = Windowed
			}
			w.Configure([]Option{mode.Option()})
			return
		}
		w.titleClickTime, w.titleClickPosition = button.Time, rootPosition
	}
	x11UngrabPointer(w.x, button.Time)
	xev := new(xEvent)
	ev := (*xClientMessageEvent)(unsafe.Pointer(xev))
	*ev = xClientMessageEvent{Type: xClientMessage, Display: w.x, Window: w.xw, MessageType: w.atoms.moveResize, Format: 32}
	ev.Data = [5]int{int(button.XRoot), int(button.YRoot), direction, 1, 1}
	xSendEvent(w.x, xDefaultRootWindow(w.x), xFalse, xSubstructureNotifyMask|xSubstructureRedirectMask, xev)
	xFlush(w.x)
}
