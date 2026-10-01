package theme

import "github.com/dyike/keel/ui/internal/loop"

// ReducedMotion is the application preference, false until configured. Native
// preference discovery is unavailable on unsupported platforms.
var ReducedMotion bool

// SetReducedMotion runs under the UI frame lock, like Apply.
func SetReducedMotion(reduce bool) {
	if ReducedMotion != reduce {
		ReducedMotion = reduce
		loop.InvalidateAll()
	}
}
