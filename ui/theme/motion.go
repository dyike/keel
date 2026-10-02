package theme

import "github.com/dyike/keel/ui/internal/loop"

// ReducedMotion is the effective preference. It follows the system by default;
// SetReducedMotion gives the application an explicit override.
var ReducedMotion bool
var systemReducedMotion, motionOverride bool

// SetReducedMotion overrides the system preference. Call under the UI frame lock.
func SetReducedMotion(reduce bool) {
	motionOverride = true
	applyMotion(reduce)
}

// FollowSystemMotion removes the application override and applies the latest
// system value. Unsupported platforms default to allowing motion.
func FollowSystemMotion() { motionOverride = false; applyMotion(systemReducedMotion) }

// SetSystemReducedMotion is the window backend's preference bridge. Native
// changes are remembered while an application override is active.
func SetSystemReducedMotion(reduce bool) {
	systemReducedMotion = reduce
	if !motionOverride {
		applyMotion(reduce)
	}
}
func applyMotion(reduce bool) {
	if ReducedMotion != reduce {
		ReducedMotion = reduce
		loop.InvalidateAll()
	}
}
