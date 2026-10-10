package el

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
)

// Optional Gio capability: upstream editors keep their native wheel/caret
// scrolling. Keel's fork also exposes that viewport to a draggable scrollbar.
type editorViewport interface {
	ScrollBounds() image.Rectangle
	ScrollOffset() image.Point
	ScrollTo(image.Point)
}

func inputScrollbarWidth(ed *widget.Editor, multiline bool, width int) int {
	if _, ok := any(ed).(editorViewport); !ok || !multiline {
		return 0
	}
	return width
}

func (e *engine) updateInputScrollbar(n *Node, st *elemState, size image.Point, track image.Rectangle) {
	sc, ok := any(&st.editor).(editorViewport)
	if !ok || !n.input.multiline {
		return
	}
	offset := sc.ScrollOffset()
	maximum := max(0, sc.ScrollBounds().Max.Y)
	delta := st.scroll.Update(e.gtx.Metric, e.gtx.Source, e.gtx.Now, gesture.Vertical, pointer.ScrollRange{}, pointer.ScrollRange{Min: -offset.Y, Max: maximum - offset.Y})
	if delta != 0 {
		sc.ScrollTo(image.Pt(offset.X, offset.Y+delta))
		offset = sc.ScrollOffset()
		e.gtx.Execute(op.InvalidateCmd{})
	}
	total := size.Y + maximum
	wasActive := st.scrollbarY.active
	next := st.scrollbarY.update(e.gtx, track, false, offset.Y, size.Y, total, e.dp(24), nil)
	if !wasActive && st.scrollbarY.active {
		e.gtx.Execute(key.FocusCmd{Tag: &st.editor})
	}
	if next != offset.Y {
		sc.ScrollTo(image.Pt(offset.X, next))
		e.gtx.Execute(op.InvalidateCmd{})
	}
}

func (e *engine) paintInputScrollbar(n *Node, st *elemState, size image.Point, track image.Rectangle) {
	sc, ok := any(&st.editor).(editorViewport)
	if !ok || !n.input.multiline {
		return
	}
	offset := sc.ScrollOffset().Y
	if offset != st.scrollY || st.scrollbarY.active {
		st.scrollVisibleUntil = e.gtx.Now.Add(ScrollbarLinger)
	}
	st.scrollY = offset
	st.scrollView = size.Y
	st.scrollContent = size.Y + max(0, sc.ScrollBounds().Max.Y)
	hovered := st.scrollHover.Update(e.gtx.Source)
	mode, wanted, alpha := st.scrollbarVisibility(e.gtx, n, hovered, true)
	defer clip.Rect(image.Rectangle{Max: size}).Push(e.gtx.Ops).Pop()
	if mode == ScrollbarHover {
		pass := pointer.PassOp{}.Push(e.gtx.Ops)
		st.scrollHover.Add(e.gtx.Ops)
		pass.Pop()
	}
	if alpha > 0 || wanted {
		st.scrollbarY.paint(e.gtx, track, false, offset, size.Y, st.scrollContent, e.dp(24), alpha, wanted)
	}
	if st.scrollContent > st.scrollView {
		area := clip.Rect(track).Push(e.gtx.Ops)
		pass := pointer.PassOp{}.Push(e.gtx.Ops)
		st.scroll.Add(e.gtx.Ops)
		pass.Pop()
		area.Pop()
	}
}
