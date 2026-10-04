package kit

import "github.com/dyike/keel/ui/theme"

// CalendarSize controls day cells, navigation and text. Medium preserves defaults.
type CalendarSize uint8

const (
	CalendarSizeMedium CalendarSize = iota
	CalendarSizeXSmall
	CalendarSizeSmall
	CalendarSizeLarge
)

func (v *CalendarView) Size(size CalendarSize) *CalendarView {
	if size <= CalendarSizeLarge {
		v.size = size
	}
	return v
}

type calendarMetrics struct{ width, height, weekday, nav, font float32 }

func (v *CalendarView) metrics() calendarMetrics {
	switch v.size {
	case CalendarSizeXSmall:
		return calendarMetrics{28, 24, 20, 24, theme.TextXs}
	case CalendarSizeSmall:
		return calendarMetrics{32, 28, 22, 26, theme.TextSm}
	case CalendarSizeLarge:
		return calendarMetrics{44, 40, 28, 36, theme.TextBody}
	default:
		return calendarMetrics{36, 32, 24, 28, theme.TextMd}
	}
}
