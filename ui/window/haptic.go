package window

// HapticStyle selects a tactile feedback intensity on platforms with haptic
// hardware (iOS Taptic Engine, macOS Force Touch trackpads).
type HapticStyle uint8

const (
	HapticSelection HapticStyle = iota
	HapticLight
	HapticMedium
	HapticHeavy
)

// Haptic triggers native tactile feedback when supported by the current
// platform and hardware. It is safe to call from any goroutine and is a no-op
// on platforms without haptic feedback.
func Haptic(style HapticStyle) {
	platformHaptic(style)
}
