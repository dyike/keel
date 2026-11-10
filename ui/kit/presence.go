package kit

import (
	"time"

	"github.com/dyike/keel/ui/el"
)

// Floating controls share the same finite entry/exit clock and inert exit
// behavior. The root owns state, so removing the component cancels it as well.
func floatingPresence(cx *el.Context, key string, open bool) el.PresenceFrame {
	return cx.Presence(key, open, el.PresenceOptions{Enter: el.Pose{Fade: 1}, Exit: el.Pose{Fade: 1}, Animation: el.Tween(160*time.Millisecond, el.EaseOut)})
}
