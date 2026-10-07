package el

import (
	"testing"
	"time"

	"github.com/dyike/keel/ui/theme"
)

func TestScrollbarFadeAndModes(t *testing.T) {
	old := theme.ReducedMotion
	defer func() { theme.ReducedMotion = old }()
	theme.ReducedMotion = false
	t0 := time.Unix(100, 0)
	if a := scrollbarAlpha(true, t0.Add(60*time.Millisecond), t0, t0); a < .4 || a > .6 {
		t.Fatal("halfway through fading in", a)
	}
	if a := scrollbarAlpha(true, t0.Add(time.Second), t0, t0); a != 1 {
		t.Fatal("faded in", a)
	}
	if a := scrollbarAlpha(false, t0.Add(100*time.Millisecond), t0, t0); a < .4 || a > .6 {
		t.Fatal("halfway through fading out", a)
	}
	if a := scrollbarAlpha(false, t0.Add(time.Second), t0, t0); a != 0 {
		t.Fatal("faded out", a)
	}
	theme.ReducedMotion = true
	if scrollbarAlpha(false, t0.Add(time.Millisecond), t0, t0) != 0 || scrollbarAlpha(true, t0, t0, t0) != 1 {
		t.Fatal("reduced motion does not fade")
	}

	previousDefault := ScrollbarMode(scrollbarDefault.Load())
	previousAutoHide := theme.SystemScrollbarsAutoHide()
	defer SetScrollbarDefault(previousDefault)
	defer theme.SetSystemScrollbarsAutoHide(previousAutoHide)
	theme.SetSystemScrollbarsAutoHide(false)
	if resolveScrollbars(0, false) != ScrollbarAlways {
		t.Fatal("system default with persistent bars is Always")
	}
	theme.SetSystemScrollbarsAutoHide(true)
	if resolveScrollbars(0, false) != ScrollbarScrolling {
		t.Fatal("system default follows the platform")
	}
	if resolveScrollbars(ScrollbarHover, true) != ScrollbarHover || resolveScrollbars(ScrollbarAlways, true) != ScrollbarAlways {
		t.Fatal("an element's own mode wins")
	}
	SetScrollbarDefault(ScrollbarAlways)
	if resolveScrollbars(0, false) != ScrollbarAlways {
		t.Fatal("explicit global default wins over the platform")
	}
}
