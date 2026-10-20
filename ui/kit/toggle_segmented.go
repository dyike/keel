package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Segmented groups adjacent buttons with a 1dp background gap. Explicit Gap
// overrides that separation.
// Selection behavior is unchanged: single by default, multiple with Multiple.
func (v *ToggleGroupView) Segmented(on bool) *ToggleGroupView { v.segmented = on; return v }

// Gap sets finite nonnegative spacing in dp. Segmented defaults to 1dp;
// ordinary groups default to theme.SpaceXs. ResetGap restores that default.
func (v *ToggleGroupView) Gap(dp float32) *ToggleGroupView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.gap = &dp
	}
	return v
}
func (v *ToggleGroupView) ResetGap() *ToggleGroupView { v.gap = nil; return v }

// Unselected segments get a fill so the group reads as one control without a border.
func decorateToggleSegment(b *el.DivEl, first, last, on bool) {
	if !on {
		b.Bg(theme.Subtle)
	}
	left, right := float32(0), float32(0)
	if first {
		left = theme.RadiusMd
	}
	if last {
		right = theme.RadiusMd
	}
	b.RoundedCorners(left, right, right, left)
}
