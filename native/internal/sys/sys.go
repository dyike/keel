// Package sys binds the platform code. Public packages validate arguments and
// call it; it never selects behavior by OS outside build tags.
package sys

// Display bounds are logical points; the origin is the primary display's top left.
type Display struct {
	ID                      uint32
	X, Y, Width, Height     float64
	PixelWidth, PixelHeight int
	Primary                 bool
}

// Modifier bits understood by the native hotkey code.
const (
	ModCtrl = 1 << iota
	ModAlt
	ModShift
	ModCmd
)
