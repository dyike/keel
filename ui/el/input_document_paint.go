package el

import (
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/inputcontent"
	"github.com/dyike/keel/ui/theme"
	"image"
)

type inputTokenHit struct {
	span             InputTokenSpan
	rects            []image.Rectangle
	pressed, dragged bool
	origin           f32.Point
}

func (e *engine) paintInputTokens(n *Node, st *elemState, g layout.Context) {
	spans := n.input.document.Content().Presentation()
	prepareInputTokenHits(st, len(spans.Spans))
	for i, span := range spans.Spans {
		hit := st.inputTokenHits[i]
		actual := InputTokenSpan{Range: span.Source, Token: span.Token}
		if hit.span != actual {
			hit.pressed = false
			hit.span = actual
		}
		hit.rects = hit.rects[:0]
		runes, err := inputcontent.RuneRange(spans.Text, span.Display)
		if err != nil {
			continue
		}
		for _, region := range st.editor.Regions(runes.Start, runes.End, nil) {
			r := region.Bounds.Intersect(image.Rectangle{Max: g.Constraints.Max})
			if r.Empty() {
				continue
			}
			hit.rects = append(hit.rects, r)
			radius := g.Dp(3)
			paint.FillShape(g.Ops, theme.Subtle, clip.RRect{Rect: r, NE: radius, NW: radius, SE: radius, SW: radius}.Op(g.Ops))

		}
	}
}
func (e *engine) inputTokenClicks(n *Node, st *elemState) {
	for _, hit := range st.inputTokenHits {
		for {
			ev, ok := e.gtx.Event(pointer.Filter{Target: hit, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
			if !ok {
				break
			}
			p, ok := ev.(pointer.Event)
			if !ok {
				continue
			}
			switch p.Kind {
			case pointer.Press:
				hit.pressed = e.gtx.Enabled() && p.Buttons == pointer.ButtonPrimary && p.Modifiers == 0
				hit.dragged = false
				hit.origin = p.Position
			case pointer.Drag:
				delta := p.Position.Sub(hit.origin)
				threshold := float32(e.gtx.Dp(4))
				if delta.X*delta.X+delta.Y*delta.Y > threshold*threshold {
					hit.dragged = true
				}
			case pointer.Cancel:
				hit.pressed = false
			case pointer.Release:
				activate := hit.pressed && !hit.dragged && p.Modifiers == 0
				hit.pressed = false
				inside := false
				for _, r := range hit.rects {
					if p.Position.Round().In(r) {
						inside = true
					}
				}
				if !activate || !inside || !e.gtx.Enabled() {
					continue
				}
				d := n.input.document
				found := false
				for _, span := range d.Content().Tokens() {
					if span == hit.span {
						found = true
						break
					}
				}
				if !found {
					continue
				}
				_ = d.Select(hit.span.Range)
				documentSyncEditor(st, d)
				if fn := n.input.onTokenActivate; fn != nil {
					token := hit.span.Token
					core.Call(e.gtx, func() { fn(token) })
				}
			}
		}
	}
}

func inputTokenHitAreas(st *elemState, g layout.Context) {
	for _, hit := range st.inputTokenHits {
		for _, r := range hit.rects {
			stack := clip.Rect(r).Push(g.Ops)
			pass := pointer.PassOp{}.Push(g.Ops)
			event.Op(g.Ops, hit)
			semantic.LabelOp(hit.span.Token.Display()).Add(g.Ops)
			pass.Pop()
			stack.Pop()
		}
	}
}

func prepareInputTokenHits(st *elemState, count int) {
	if len(st.inputTokenHits) > count {
		st.inputTokenHits = st.inputTokenHits[:count]
	}
	for len(st.inputTokenHits) < count {
		st.inputTokenHits = append(st.inputTokenHits, &inputTokenHit{})
	}
}
