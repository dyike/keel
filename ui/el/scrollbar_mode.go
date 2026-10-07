package el

import (
	"sync/atomic"
	"time"

	"github.com/dyike/keel/ui/theme"
)

// ScrollbarMode controls when an overflowing scroll container shows its bars.
// It applies to both axes and does not change the content's layout or offset.
type ScrollbarMode uint8

const (
	ScrollbarAlways    ScrollbarMode = iota // visible whenever content overflows
	ScrollbarHover                          // visible while the pointer is inside the viewport or dragging a bar
	ScrollbarScrolling                      // visible during offset changes and briefly afterward
	// ScrollbarSystem follows the platform's setting: macOS "Show scroll
	// bars", Windows "Automatically hide scroll bars". Elsewhere it is Always.
	ScrollbarSystem
)

// ScrollbarLinger is the idle delay before Scrolling mode hides the bars.
const ScrollbarLinger = 900 * time.Millisecond

// Bars fade in and out over these durations, instantly with reduced motion.
const (
	scrollbarFadeIn  = 120 * time.Millisecond
	scrollbarFadeOut = 200 * time.Millisecond
)

// Scrollbars selects the display mode for this ScrollX/ScrollY element. Invalid
// modes are ignored. Hidden bars have no pointer hit area; scrolling and
// keyboard navigation remain available. ScrollOffset still hides all bars.
// Elements without a mode use SetScrollbarDefault's.
func (s *Styled[T]) Scrollbars(mode ScrollbarMode) *T {
	if mode <= ScrollbarSystem {
		s.n.style.scrollbarMode = mode
		s.n.style.scrollbarModeSet = true
	}
	return s.self
}

var scrollbarDefault atomic.Uint32 // a ScrollbarMode

func init() { scrollbarDefault.Store(uint32(ScrollbarSystem)) }

// SetScrollbarDefault sets the mode of scroll containers that do not choose
// one, ScrollbarSystem by default. Pass ScrollbarAlways to keep overflowing
// bars visible. Windows redraw on their next frame.
func SetScrollbarDefault(mode ScrollbarMode) {
	if mode <= ScrollbarSystem {
		scrollbarDefault.Store(uint32(mode))
	}
}

// SystemScrollbars is the platform's preference as last read: Scrolling
// where bars hide at rest, otherwise Always.
// ui/window reads the setting into theme.SystemScrollbarsAutoHide.
func SystemScrollbars() ScrollbarMode {
	if theme.SystemScrollbarsAutoHide() {
		return ScrollbarScrolling
	}
	return ScrollbarAlways
}

// resolveScrollbars is the mode an element shows its bars by.
func resolveScrollbars(mode ScrollbarMode, set bool) ScrollbarMode {
	if !set {
		mode = ScrollbarMode(scrollbarDefault.Load())
	}
	if mode == ScrollbarSystem {
		mode = SystemScrollbars()
	}
	return mode
}

// scrollbarAlpha fades bars in after shownAt and out after lastWanted.
func scrollbarAlpha(want bool, now, shownAt, lastWanted time.Time) float32 {
	if ReducedMotion() {
		if want {
			return 1
		}
		return 0
	}
	if want {
		return min(1, float32(now.Sub(shownAt))/float32(scrollbarFadeIn))
	}
	return max(0, 1-float32(now.Sub(lastWanted))/float32(scrollbarFadeOut))
}
