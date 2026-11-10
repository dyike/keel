package el

import (
	"image"
	"time"

	"github.com/dyike/keel/gio/f32"
)

// Snapshot only drawing state at the last displayed frame. The previous node,
// callbacks and semantics never survive the handoff.
func (s *sharedLayoutState) handoffContent(n *Node) {
	s.contentFrom = nil
	s.contentHandoff = false
	if n.content != nil && s.content != nil && s.contentSize.X > 0 && s.contentSize.Y > 0 {
		s.contentHandoff = true
		if len(s.content.tracks) == 0 {
			s.contentFrom = s.content.last.shape()
		} else {
			p := min(max(s.content.progress.Get(), 0), 1)
			for _, track := range s.content.tracks {
				g := track.at(p, s.contentTime)
				if g.alpha > 1.0/255 {
					s.contentFrom = append(s.contentFrom, g)
				}
			}
		}
		ratio := f32.Pt(float32(n.size.X)/float32(s.contentSize.X), float32(n.size.Y)/float32(s.contentSize.Y))
		for i := range s.contentFrom {
			g := &s.contentFrom[i]
			if !g.rasterSet {
				g.raster, g.rasterSet = g.position, true
			}
			g.position = f32.Pt(g.position.X*ratio.X, g.position.Y*ratio.Y)
			g.stretch = f32.Pt((1+g.stretch.X)*ratio.X-1, (1+g.stretch.Y)*ratio.Y-1)
			g.viewport = affineBounds(f32.AffineId().Scale(f32.Point{}, ratio), g.viewport)
		}
	}
	s.content = nil
}

func (e *engine) sharedContent(n *Node) *sharedLayoutState {
	if n.sharedLayout == nil {
		return nil
	}
	s := e.sharedLayouts[n.sharedLayout.identity]
	if s == nil || s.owner != n.key {
		return nil
	}
	return s
}

func (s *sharedLayoutState) recordContent(content *textContentState, size image.Point, now time.Time) {
	s.content, s.contentSize, s.contentTime = content, size, now
	s.contentFrom, s.contentHandoff = nil, false
}
