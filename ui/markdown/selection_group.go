package markdown

import (
	"image"
	"slices"

	"gioui.org/io/event"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// SelectionGroup joins separately rendered documents into one selectable text
// flow. Keep it across frames. The documents may live in embedded roots or a
// virtual list; SetDocuments supplies their reading order, including offscreen
// documents. Collapsed content should be omitted from that order.
type SelectionGroup struct {
	docs               []*Doc
	starts             []int
	lengths            []int
	revisions          []uint64
	index              map[*Doc]int
	selection          documentSelection
	visible            []groupPart
	painting, scrolled bool
	origin             image.Point
}

type groupPart struct {
	doc  *Doc
	part selectionPart
}

// SetDocuments sets the reading order. Appending or prepending history retains
// selection in existing documents; removing a selected document clears it.
func (g *SelectionGroup) SetDocuments(docs ...*Doc) {
	if slices.Equal(g.docs, docs) {
		return
	}
	// Capture endpoints by document identity before inserting new history.
	type endpoint struct {
		doc    *Doc
		offset int
	}
	locate := func(pos int) endpoint {
		for i := len(g.starts) - 1; i >= 0; i-- {
			if pos >= g.starts[i] {
				return endpoint{g.docs[i], min(pos-g.starts[i], g.lengths[i])}
			}
		}
		return endpoint{}
	}
	ends := [4]endpoint{locate(g.selection.sel[0]), locate(g.selection.sel[1]), locate(g.selection.anchor[0]), locate(g.selection.anchor[1])}
	for _, d := range g.docs {
		if d.selectionGroup == g {
			d.selectionGroup = nil
			d.selection.clear()
		}
	}
	clear(g.visible)
	g.visible = g.visible[:0]
	g.docs = append(g.docs[:0], docs...)
	g.index = make(map[*Doc]int, len(docs))
	g.starts, g.lengths, g.revisions = nil, nil, nil
	for i, d := range g.docs {
		d.selectionGroup = g
		g.index[d] = i
	}
	g.collect()
	for i, e := range ends {
		n, ok := g.index[e.doc]
		if !ok {
			g.selection.clear()
			break
		}
		pos := g.starts[n] + min(e.offset, g.lengths[n])
		if i < 2 {
			g.selection.sel[i] = pos
		} else {
			g.selection.anchor[i-2] = pos
		}
	}
	g.distribute()
}

// Decorate installs the group's input handler on a common ancestor of its
// documents. Call once per frame, after adding children to root. Keep scroll
// controls inside this ancestor so pointer capture survives virtual row removal.
func (g *SelectionGroup) Decorate(cx *el.Context, root *el.DivEl) *el.DivEl {
	return root.Decorate(func(gtx core.C, draw func()) {
		if !gtx.Enabled() {
			draw()
			return
		}
		s := &g.selection
		s.bounded = true
		// Resolve text for unmounted rows only while selecting. Normal scrolling
		// keeps virtualization; Ctrl+A/Copy must still include unmounted messages.
		if s.dragging || s.sel[0] != s.sel[1] {
			var measureOps op.Ops
			for _, d := range g.docs {
				if !d.ranges.ready || d.ranges.source != d.src {
					measure := el.Embed(d)
					m := gtx.Disabled()
					measureOps.Reset()
					m.Ops = &measureOps
					m.Constraints.Min = image.Point{}
					measure.Layout(m)
				}
			}
		}
		inputOrigin := g.origin
		g.origin, _ = cx.PaintGeometry()
		s.origin = g.origin
		g.visible = g.visible[:0]
		g.painting, g.scrolled = true, false
		area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
		event.Op(gtx.Ops, s)
		draw()
		area.Pop()
		g.painting = false
		g.collect()
		if s.all {
			s.sel = [2]int{0, len(s.text)}
			s.all = false
			g.distribute()
		}
		before, moved, dragging := s.sel, s.moved, s.dragging
		s.update(gtx, inputOrigin)
		if s.dragging && s.dragged {
			s.extend(s.pointerRoot.Sub(g.origin))
		}
		if before != s.sel || moved != s.moved || dragging != s.dragging {
			g.distribute()
			gtx.Execute(op.InvalidateCmd{})
		}
	})
}

func (d *Doc) paintSelection(gtx core.C, cx *el.Context, root el.Element, draw func()) {
	g := d.selectionGroup
	if g == nil {
		d.selection.paint(gtx, cx, root, draw)
		return
	}
	d.selection.collect(root, gtx)
	if gtx.Enabled() && g.painting {
		origin, viewport := cx.PaintGeometry()
		for _, part := range d.selection.parts {
			part.rect = part.rect.Add(origin.Sub(g.origin))
			part.visible = part.visible.Add(origin.Sub(g.origin))
			g.visible = append(g.visible, groupPart{d, part})
		}
		if g.selection.dragging && g.selection.dragged && !g.scrolled {
			g.selection.autoScroll(gtx, cx, viewport)
			g.scrolled = true
		}
	}
	draw()
}

func (g *SelectionGroup) collect() {
	s := &g.selection
	// Preserve endpoints relative to their document when earlier text grows or
	// offscreen documents are measured for the first time.
	remap := func(pos int, starts, lengths []int) int {
		for i := len(g.starts) - 1; i >= 0; i-- {
			if pos >= g.starts[i] {
				return starts[i] + min(pos-g.starts[i], lengths[i])
			}
		}
		return pos
	}
	changed := g.starts == nil || len(g.revisions) != len(g.docs)
	if !changed {
		for i, d := range g.docs {
			if g.revisions[i] != d.selection.revision {
				changed = true
				break
			}
		}
	}
	if changed {
		starts, lengths := make([]int, len(g.docs)), make([]int, len(g.docs))
		revisions := make([]uint64, len(g.docs))
		var text []rune
		for i, d := range g.docs {
			if i > 0 {
				text = append(text, '\n', '\n')
			}
			starts[i], lengths[i], revisions[i] = len(text), len(d.selection.text), d.selection.revision
			text = append(text, d.selection.text...)
		}
		if len(g.starts) == len(starts) {
			for i := range s.sel {
				s.sel[i] = remap(s.sel[i], starts, lengths)
				s.anchor[i] = remap(s.anchor[i], starts, lengths)
			}
		}
		g.starts, g.lengths, g.revisions = starts, lengths, revisions
		s.text = text
		if s.all {
			s.sel = [2]int{0, len(text)}
		}
	}
	clear(s.parts)
	s.parts = s.parts[:0]
	for _, p := range g.visible {
		i, ok := g.index[p.doc]
		if !ok {
			continue
		}
		part := p.part
		part.start += g.starts[i]
		part.end += g.starts[i]
		s.parts = append(s.parts, part)
	}
	// Paint order normally equals reading order; nested layers need not.
	slices.SortStableFunc(s.parts, func(a, b selectionPart) int { return a.start - b.start })
	g.distribute()
}

func (g *SelectionGroup) distribute() {
	for i, d := range g.docs {
		if i >= len(g.starts) {
			break
		}
		for j, pos := range g.selection.sel {
			d.selection.sel[j] = max(0, min(g.lengths[i], pos-g.starts[i]))
		}
		d.selection.moved = g.selection.moved
	}
}
