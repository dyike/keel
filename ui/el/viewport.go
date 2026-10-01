package el

import (
	"gioui.org/op"
	"image"
)

// PaintGeometry reports the current element's origin and clipped viewport in
// root coordinates. Use it only from Decorate, while the element is painted.
// It lets input handlers keep a pointer stationary as ancestors scroll.
func (cx *Context) PaintGeometry() (origin image.Point, viewport image.Rectangle) {
	e := &cx.root.e
	return e.origin, e.visible
}

// ScrollBy schedules a pixel delta on the nearest enclosing ScrollY container
// and returns the amount it can scroll. Call during Decorate. The next frame
// applies it before drawing children, so their input and painting agree.
func (cx *Context) ScrollBy(dy int) int {
	e := &cx.root.e
	if len(e.scrollParents) == 0 {
		return 0
	}
	st := e.scrollParents[len(e.scrollParents)-1]
	before := st.scrollY + st.scrollPending
	after := min(max(before+dy, 0), st.scrollMax)
	st.scrollPending += after - before
	if after != before {
		e.gtx.Execute(op.InvalidateCmd{})
	}
	return after - before
}
