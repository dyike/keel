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

// ScrollState reports, in dp, the scroll offset, viewport height and content
// height of the ScrollY element with id as last painted; all zero before its
// first frame. Virtual lists use it to build only the rows that show.
func (cx *Context) ScrollState(id string) (offset, viewport, content float32) {
	st := cx.scrollElem(id)
	if st == nil {
		return 0, 0, 0
	}
	px := cx.root.e.m.PxPerDp
	if px == 0 {
		px = 1
	}
	return float32(st.scrollY+st.scrollPending) / px, float32(st.scrollView) / px, float32(st.scrollContent) / px
}

// ScrollIntoView scrolls the ScrollY element with id as little as needed to
// show [top, bottom], in dp of its content. It applies when the element is
// next painted, so call it from Render or a callback.
func (cx *Context) ScrollIntoView(id string, top, bottom float32) {
	st := cx.scrollElem(id)
	if st == nil {
		return
	}
	px := cx.root.e.m.PxPerDp
	if px == 0 {
		px = 1
	}
	y := st.scrollY + st.scrollPending
	t, b := int(top*px), int(bottom*px+0.5)
	switch {
	case t < y:
		st.scrollPending += t - y
	case b > y+st.scrollView:
		st.scrollPending += min(b-(y+st.scrollView), t-y)
	}
}

func (cx *Context) scrollElem(id string) *elemState {
	for _, st := range cx.root.store.states {
		if st.id == id && st.scrolled {
			return st
		}
	}
	return nil
}
