package markdown

import (
	"image"
	"io"
	"strings"
	"unicode/utf8"

	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op/clip"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// documentSelection owns keyboard focus and the rune range for all text in
// one Doc. Bounds are in the document's coordinate system, including text
// outside the viewport, so selection also works in a scrolling chat.
type documentSelection struct {
	parts    []selectionPart
	text     []rune
	sel      [2]int // anchor, focus
	dragging bool
	moved    bool // suppress link activation after a selection drag
	pointer  pointer.ID
	press    image.Point
}

type selectionPart struct {
	r       *richBlock
	rect    image.Rectangle
	visible image.Rectangle
	start   int
	end     int
}

func (s *documentSelection) paint(gtx core.C, root el.Element, draw func()) {
	s.collect(root, gtx)
	s.update(gtx)
	// Keep the area open around children: it receives their presses as an
	// ancestor, including presses on links. A drag then grabs the pointer to
	// cancel link clicks and retain input outside the document's bounds.
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, s)
	draw()
	area.Pop()
}

func (s *documentSelection) collect(root el.Element, gtx core.C) {
	oldLen, count := len(s.parts), 0
	changed := false
	el.VisitWidgets(root, gtx.Metric, func(w core.Widget, rect image.Rectangle) {
		var r *richBlock
		visible := rect
		switch w := w.(type) {
		case *richBlock:
			r = w
		case *codeBody:
			r = w.view.rich
			rect, visible = w.selectionBounds(rect)
		}
		if r == nil || r.decoration {
			return
		}
		if count == len(s.parts) {
			s.parts = append(s.parts, selectionPart{})
		}
		p := &s.parts[count]
		if p.r != r {
			changed = true
		}
		p.r, p.rect, p.visible = r, rect, visible
		r.rt.document = s
		count++
	})
	changed = changed || oldLen != count
	// Drop references to replaced blocks so streaming tails can be collected.
	clear(s.parts[count:])
	s.parts = s.parts[:count]
	if !changed {
		return
	}
	var text strings.Builder
	pos := 0
	for i := range s.parts {
		p := &s.parts[i]
		if i > 0 {
			sep := p.r.separator
			if sep == "" {
				sep = "\n\n"
			}
			text.WriteString(sep)
			pos += utf8.RuneCountInString(sep)
		}
		p.start = pos
		text.WriteString(p.r.plain)
		pos += utf8.RuneCountInString(p.r.plain)
		p.end = pos
		p.r.rt.offset, p.r.rt.length = p.start, p.end-p.start
	}
	s.text = []rune(text.String())
	s.sel[0], s.sel[1] = min(s.sel[0], pos), min(s.sel[1], pos)
}

func (s *documentSelection) update(gtx core.C) {
	for {
		ev, ok := gtx.Event(
			pointer.Filter{Target: s, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel},
			key.FocusFilter{Target: s},
			key.Filter{Focus: s, Name: "C", Required: key.ModShortcut},
			key.Filter{Focus: s, Name: "A", Required: key.ModShortcut},
		)
		if !ok {
			break
		}
		switch e := ev.(type) {
		case pointer.Event:
			pt := e.Position.Round()
			switch e.Kind {
			case pointer.Press:
				if !e.Buttons.Contain(pointer.ButtonPrimary) || s.dragging || !s.contains(pt) {
					break
				}
				i := s.hit(pt)
				s.sel = [2]int{i, i}
				s.dragging, s.moved = true, false
				s.pointer, s.press = e.PointerID, pt
				gtx.Execute(key.FocusCmd{Tag: s})
			case pointer.Drag:
				if !s.dragging || e.PointerID != s.pointer {
					break
				}
				s.sel[1] = s.hit(pt)
				dist := pt.Sub(s.press)
				if !s.moved && (s.sel[0] != s.sel[1] || dist.X*dist.X+dist.Y*dist.Y > 9) {
					s.moved = true
					gtx.Execute(pointer.GrabCmd{Tag: s, ID: e.PointerID})
				}
			case pointer.Release:
				if s.dragging && e.PointerID == s.pointer {
					s.sel[1] = s.hit(pt)
					s.dragging = false
				}
			case pointer.Cancel:
				s.dragging = false
			}
		case key.FocusEvent:
			// The root blurs on every press before we claim focus again.
			// Ignore that queued blur if this press already refocused us.
			if !e.Focus && !gtx.Focused(s) {
				s.clear()
			}
		case key.Event:
			if e.State != key.Press {
				break
			}
			switch e.Name {
			case "A":
				s.sel = [2]int{0, len(s.text)}
			case "C":
				if text := s.selectedText(); text != "" {
					gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
				}
			}
		}
	}
}

func (s *documentSelection) clear() {
	s.sel = [2]int{}
	s.dragging, s.moved = false, false
}

func (s *documentSelection) selRange() (int, int) {
	return min(s.sel[0], s.sel[1]), max(s.sel[0], s.sel[1])
}

func (s *documentSelection) selectedText() string {
	lo, hi := s.selRange()
	return string(s.text[min(lo, len(s.text)):min(hi, len(s.text))])
}

func (s *documentSelection) contains(pt image.Point) bool {
	for _, p := range s.parts {
		if pt.In(p.visible) {
			return true
		}
	}
	return false
}

// hit finds the nearest text block, then maps its local glyph coordinates to
// the document's rune indexes. Rectangle distance handles both vertical gaps
// and table cells beside each other.
func (s *documentSelection) hit(pt image.Point) int {
	if len(s.parts) == 0 {
		return 0
	}
	nearest, distance := 0, int64(1<<62)
	top, bottom := s.parts[0].visible.Min.Y, s.parts[0].visible.Max.Y
	for i, p := range s.parts {
		top, bottom = min(top, p.visible.Min.Y), max(bottom, p.visible.Max.Y)
		dx := int64(max(p.visible.Min.X-pt.X, pt.X-p.visible.Max.X, 0))
		dy := int64(max(p.visible.Min.Y-pt.Y, pt.Y-p.visible.Max.Y, 0))
		if d := dx*dx + dy*dy; d < distance {
			nearest, distance = i, d
		}
	}
	if pt.Y < top {
		return 0
	}
	if pt.Y >= bottom {
		return len(s.text)
	}
	p := s.parts[nearest]
	return p.start + min(p.r.rt.hit(pt.Sub(p.rect.Min)), p.end-p.start)
}
