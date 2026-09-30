// Package screen lists displays and captures screenshots.
package screen

import (
	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

// Display bounds are logical points; the origin is the primary display's top left.
type Display = sys.Display

func Displays() ([]Display, error) { return sys.Displays() }

// Capture returns a PNG of the display at native resolution. It needs the
// ScreenRecording permission, never prompts, and times out after 10 seconds;
// call it off the UI goroutine.
func Capture(displayID uint32) ([]byte, error) {
	if displayID == 0 {
		return nil, native.ErrInvalidArgument
	}
	return sys.Capture(displayID)
}
