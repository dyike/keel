//go:build darwin && !ios

package window

import "github.com/dyike/keel/ui/internal/appkit"

const (
	nsHapticFeedbackPatternGeneric     = 0
	nsHapticFeedbackPatternAlignment   = 1
	nsHapticFeedbackPatternLevelChange = 2
	nsHapticFeedbackPerformanceTimeNow = 1
)

func platformHaptic(style HapticStyle) {
	appkit.MainAsync(func() {
		mgr := appkit.Class("NSHapticFeedbackManager")
		if mgr == 0 {
			return
		}
		performer := appkit.Send(mgr, "defaultPerformer")
		if performer == 0 {
			return
		}
		pattern := uintptr(nsHapticFeedbackPatternGeneric)
		switch style {
		case HapticSelection, HapticLight:
			pattern = nsHapticFeedbackPatternAlignment
		case HapticHeavy:
			pattern = nsHapticFeedbackPatternLevelChange
		}
		appkit.Send(performer, "performFeedbackPattern:performanceTime:", pattern, nsHapticFeedbackPerformanceTimeNow)
	})
}
