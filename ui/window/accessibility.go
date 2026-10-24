package window

import (
	"image"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
)

// Actions of the platform accessibility bridges, in window pixels. They feed
// Gio's router (Window.ProcessInput), so they run where the router may be
// touched: the main thread inside an AppKit callback on macOS, the window's
// goroutine between events on X11 and Wayland, where Gio's driver runs
// inside Event.

// pressElement clicks the center of e's visible part, as a screen reader's
// press does. An element scrolled out of view is not pressed: clients scroll
// it into view first.
func pressElement(w *Window, e Element) bool {
	shown := e.box().Intersect(e.viewport)
	if w.win == nil || e.Disabled || shown.Empty() {
		return false
	}
	c := shown.Min.Add(shown.Max).Div(2)
	p := f32.Pt(float32(c.X), float32(c.Y))
	w.win.ProcessInput(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p})
	w.win.ProcessInput(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p})
	w.win.ProcessInput(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
	return true
}

// scrollToElement scrolls the area that holds e until e shows, its top first
// when it is taller than the area: a wheel scroll at the area's center, as a
// person would.
func scrollToElement(w *Window, e Element) {
	b, vp := e.box(), e.viewport
	delta := func(lo, hi, vlo, vhi int) int {
		switch {
		case lo < vlo:
			return lo - vlo
		case hi > vhi:
			return min(hi-vhi, lo-vlo)
		}
		return 0
	}
	scrollArea(w, vp, delta(b.Min.X, b.Max.X, vp.Min.X, vp.Max.X), delta(b.Min.Y, b.Max.Y, vp.Min.Y, vp.Max.Y))
}

// scrollPage scrolls a scroll area by most of its size; dx and dy are -1, 0
// or 1.
func scrollPage(w *Window, area Element, dx, dy int) {
	r := area.box().Intersect(area.viewport)
	scrollArea(w, r, dx*r.Dx()*9/10, dy*r.Dy()*9/10)
}

func scrollArea(w *Window, area image.Rectangle, dx, dy int) {
	if w.win == nil || area.Empty() || dx == 0 && dy == 0 {
		return
	}
	c := area.Min.Add(area.Max).Div(2)
	w.win.ProcessInput(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(float32(c.X), float32(c.Y)), Scroll: f32.Pt(float32(dx), float32(dy))})
}
