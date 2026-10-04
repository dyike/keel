package el

import "time"

// ScrollbarMode controls when an overflowing scroll container shows its bars.
// It applies to both axes and does not change the content's layout or offset.
type ScrollbarMode uint8

const (
	ScrollbarAlways    ScrollbarMode = iota // default: visible whenever content overflows
	ScrollbarHover                          // visible while the pointer is inside the viewport or dragging a bar
	ScrollbarScrolling                      // visible during offset changes and briefly afterward
)

// ScrollbarLinger is the idle delay before Scrolling mode hides the bars.
const ScrollbarLinger = 900 * time.Millisecond

// Scrollbars selects the display mode for this ScrollX/ScrollY element. Invalid
// modes are ignored. Hidden bars have no pointer hit area; scrolling and
// keyboard navigation remain available. ScrollOffset still hides all bars.
func (s *Styled[T]) Scrollbars(mode ScrollbarMode) *T {
	if mode <= ScrollbarScrolling {
		s.n.style.scrollbarMode = mode
	}
	return s.self
}
