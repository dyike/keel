// Package permissions checks and explicitly requests desktop permissions.
package permissions

import (
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
	"github.com/dyike/keel/internal/platform"
)

type Kind = driver.Permission
type Status = driver.Status

const (
	Accessibility   = driver.Accessibility
	ScreenRecording = driver.ScreenRecording
	InputMonitoring = driver.InputMonitoring
	NotGranted      = driver.NotGranted
	Granted         = driver.Granted
)

type Manager struct{ backend driver.Backend }

func New() *Manager { return &Manager{platform.New()} }

// Check never requests permission. NotGranted deliberately does not distinguish denial from an unanswered prompt.
func (m *Manager) Check(k Kind) (Status, error) { return m.check(k, false) }

// Request may open a system prompt. Its result reflects the current grant, not future user actions.
func (m *Manager) Request(k Kind) (Status, error) { return m.check(k, true) }
func (m *Manager) check(k Kind, request bool) (Status, error) {
	if k > InputMonitoring {
		return NotGranted, capability.ErrInvalidArgument
	}
	return m.backend.Check(k, request)
}
