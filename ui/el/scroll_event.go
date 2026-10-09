package el

import (
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/unit"
	"math"
)

// ScrollRange is the signed range of scroll deltas accepted in dp. A zero
// range passes that axis to enclosing scroll handlers.
type ScrollRange struct{ Min, Max float32 }

// ScrollEvent reports scroll deltas in dp. Gio does not identify discrete
// wheels versus trackpads or expose a gesture-end phase on this event.
type ScrollEvent struct{ X, Y float32 }

type scrollHandler struct {
	x, y ScrollRange
	fn   func(ScrollEvent)
}

// OnScroll receives scroll input within the supplied axis ranges. Excess is
// routed by Gio to enclosing handlers. Descendant handlers have priority.
// Nil removes the handler; disabled/hidden ancestors suppress delivery.
func (s *Styled[T]) OnScroll(x, y ScrollRange, fn func(ScrollEvent)) *T {
	if fn == nil {
		s.n.onScroll = nil
		return s.self
	}
	s.n.onScroll = &scrollHandler{x: x, y: y, fn: fn}
	return s.self
}

func (h *scrollHandler) filter(tag any, m unit.Metric) pointer.Filter {
	convert := func(r ScrollRange) pointer.ScrollRange {
		bound := func(v float32) int {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return 0
			}
			return m.Dp(unit.Dp(min(max(v, -1e6), 1e6)))
		}
		return pointer.ScrollRange{Min: min(bound(r.Min), 0), Max: max(bound(r.Max), 0)}
	}
	return pointer.Filter{Target: tag, Kinds: pointer.Scroll, ScrollX: convert(h.x), ScrollY: convert(h.y)}
}
