package window

import (
	"image"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/pointer"
)

// Actions of the platform accessibility bridges, as input in window
// pixels. The bridges feed it to Gio where its router may be touched: the
// main thread inside an AppKit callback on macOS and the window's goroutine
// between events on X11 and Wayland (Window.ProcessInput), or through
// Window.QueueInput from UI Automation's threads on Windows.

// clickEvents click the center of e's visible part, as a screen reader's
// press does. An element scrolled out of view or disabled is not pressed:
// clients scroll it into view first.
func clickEvents(e Element) []event.Event {
	shown := e.box().Intersect(e.viewport)
	if e.Disabled || shown.Empty() {
		return nil
	}
	c := shown.Min.Add(shown.Max).Div(2)
	p := f32.Pt(float32(c.X), float32(c.Y))
	return []event.Event{
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p},
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p},
	}
}

// scrollToEvents scroll the area that holds e until e shows, its top first
// when it is taller than the area: a wheel scroll at the area's center, as a
// person would.
func scrollToEvents(e Element) []event.Event {
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
	return scrollEvents(vp, delta(b.Min.X, b.Max.X, vp.Min.X, vp.Max.X), delta(b.Min.Y, b.Max.Y, vp.Min.Y, vp.Max.Y))
}

// pageEvents scroll a scroll area by most of its size; dx and dy are -1, 0
// or 1.
func pageEvents(area Element, dx, dy int) []event.Event {
	r := area.box().Intersect(area.viewport)
	return scrollEvents(r, dx*r.Dx()*9/10, dy*r.Dy()*9/10)
}

func scrollEvents(area image.Rectangle, dx, dy int) []event.Event {
	if area.Empty() || dx == 0 && dy == 0 {
		return nil
	}
	c := area.Min.Add(area.Max).Div(2)
	return []event.Event{pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(float32(c.X), float32(c.Y)), Scroll: f32.Pt(float32(dx), float32(dy))}}
}

// processInput feeds events to the window's router directly; see above for
// where that may happen.
func processInput(w *Window, events []event.Event) bool {
	if w.win == nil || len(events) == 0 {
		return false
	}
	for _, e := range events {
		w.win.ProcessInput(e)
	}
	return true
}
