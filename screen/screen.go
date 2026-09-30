// Package screen enumerates displays and captures PNG screenshots.
package screen

import (
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
	"github.com/dyike/keel/internal/platform"
)

type Display = driver.Display
type Rect = driver.Rect
type Manager struct{ backend driver.Backend }

func New() *Manager { return &Manager{platform.New()} }

// Displays reports global desktop bounds in logical points (origin at primary display top left).
func (m *Manager) Displays() ([]Display, error) { return m.backend.Displays() }

// Capture returns a native-resolution PNG. Call from a background goroutine, after permission is granted.
// The native operation times out after 10 seconds; it never prompts for permission.
func (m *Manager) Capture(displayID uint32) ([]byte, error) {
	if displayID == 0 {
		return nil, capability.ErrInvalidArgument
	}
	return m.backend.Capture(displayID)
}
