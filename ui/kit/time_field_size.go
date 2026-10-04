package kit

import "github.com/dyike/keel/ui/theme"

// TimeFieldSize selects coordinated frame, text, icon and segment dimensions.
type TimeFieldSize uint8

const (
	TimeFieldSizeMedium TimeFieldSize = iota
	TimeFieldSizeXSmall
	TimeFieldSizeSmall
	TimeFieldSizeLarge
)

// Size changes presentation without committing drafts or changing focus.
func (v *TimeFieldView) Size(size TimeFieldSize) *TimeFieldView {
	if size <= TimeFieldSizeLarge {
		v.size = size
	}
	return v
}

func (v *TimeFieldView) sizeMetrics() (height, font, icon, padding, verticalPadding float32) {
	switch v.size {
	case TimeFieldSizeXSmall:
		return 24, theme.TextXs, 12, 6, 2
	case TimeFieldSizeSmall:
		return 28, theme.TextSm, 14, 8, 3
	case TimeFieldSizeLarge:
		return 40, theme.TextBody, 20, 12, 6
	default:
		return float32(theme.ControlHeight), theme.TextControl, 16, 10, theme.SpaceXs
	}
}
