package kit

import (
	"image/color"
	"testing"

	"github.com/dyike/keel/ui/theme"
)

func TestButtonVisualCacheIsBoundedAndAllocationFreeWhenWarm(t *testing.T) {
	clear(buttonStyleCache)
	a := ButtonAppearance{Focus: theme.Primary, Hover: theme.Subtle, Active: theme.SubtleHover, HoverForeground: theme.Text, ActiveForeground: theme.Text}
	resolveButtonStyles(a, theme.Gradient{})
	if got := testing.AllocsPerRun(20, func() { resolveButtonStyles(a, theme.Gradient{}) }); got != 0 {
		t.Fatalf("unchanged visual callbacks allocated %.0f times", got)
	}
	for i := range 600 {
		a.Focus = color.NRGBA{R: uint8(i), G: uint8(i >> 8), A: 255}
		resolveButtonStyles(a, theme.Gradient{})
		if len(buttonStyleCache) > 256 {
			t.Fatal("custom appearance cache grew without a bound")
		}
	}
}
