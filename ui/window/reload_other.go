//go:build !darwin || ios

package window

// windowOrigin reports nothing: other platforms place new windows themselves.
func windowOrigin(view uintptr) *[2]float64 { return nil }
