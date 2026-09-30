// Package shortcut manages process-wide global hotkeys.
package shortcut

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/driver"
	"github.com/dyike/keel/internal/platform"
)

type Chord = driver.Chord
type Modifiers = driver.Modifiers

const (
	Control = driver.Control
	Alt     = driver.Alt
	Shift   = driver.Shift
	Super   = driver.Super
)

// Parse accepts Ctrl/Control, Alt/Option, Shift, Super/Cmd/Command/Win and one key.
func Parse(s string) (Chord, error) {
	var c Chord
	for _, p := range strings.Split(strings.ToLower(s), "+") {
		p = strings.TrimSpace(p)
		var m Modifiers
		switch p {
		case "ctrl", "control":
			m = Control
		case "alt", "option":
			m = Alt
		case "shift":
			m = Shift
		case "super", "cmd", "command", "win":
			m = Super
		default:
			if p == "" || c.Key != "" {
				return Chord{}, capability.ErrInvalidArgument
			}
			c.Key = p
		}
		if m != 0 {
			if c.Modifiers&m != 0 {
				return Chord{}, capability.ErrInvalidArgument
			}
			c.Modifiers |= m
		}
	}
	if c.Key == "" || c.Modifiers == 0 {
		return Chord{}, fmt.Errorf("%w: shortcut requires a modifier and key", capability.ErrInvalidArgument)
	}
	return c, nil
}

type registrar interface {
	Register(driver.Chord, func()) (func() error, error)
}

type Manager struct {
	mu      sync.Mutex
	backend registrar
	closed  bool
	entries map[Chord]*Registration
}
type Registration struct {
	owner   *Manager
	chord   Chord
	cancel  func() error
	closing bool
}

func New() *Manager { return &Manager{backend: platform.New(), entries: make(map[Chord]*Registration)} }

// Register reserves a global shortcut while the host's native event loop is running.
// Callbacks run off the UI thread; one callback per registration runs at a time.
// Repeated events coalesce while a callback is busy. A panic in a callback is not recovered.
func (m *Manager) Register(accelerator string, handler func()) (*Registration, error) {
	c, err := Parse(accelerator)
	if err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, capability.ErrInvalidArgument
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, capability.ErrClosed
	}
	if _, ok := m.entries[c]; ok {
		m.mu.Unlock()
		return nil, capability.ErrConflict
	}
	r := &Registration{owner: m, chord: c}
	m.entries[c] = r
	m.mu.Unlock()
	// Never hold a Go lock while native code synchronously dispatches to the UI thread.
	cancel, err := m.backend.Register(c, handler)
	m.mu.Lock()
	if err != nil {
		delete(m.entries, c)
		m.mu.Unlock()
		return nil, err
	}
	r.cancel = cancel
	closed := m.closed
	m.mu.Unlock()
	if closed {
		return nil, errors.Join(capability.ErrClosed, r.Close())
	}
	return r, nil
}

// Close unregisters this shortcut. An in-flight callback may finish afterward.
// Concurrent closes return ErrNotReady rather than block the host's UI thread.
func (r *Registration) Close() error {
	m := r.owner
	m.mu.Lock()
	if m.entries[r.chord] != r {
		m.mu.Unlock()
		return nil
	}
	if r.closing || r.cancel == nil {
		m.mu.Unlock()
		return capability.ErrNotReady
	}
	r.closing = true
	cancel := r.cancel
	m.mu.Unlock()
	err := cancel()
	m.mu.Lock()
	r.closing = false
	if err == nil {
		delete(m.entries, r.chord)
	}
	m.mu.Unlock()
	return err
}

// Close forbids new registrations and releases completed registrations. It is retryable on error.
// A Register in flight cleans itself up before returning ErrClosed; callers must wait for
// those Register calls to finish before stopping the native event loop.
func (m *Manager) Close() error {
	m.mu.Lock()
	m.closed = true
	entries := make([]*Registration, 0, len(m.entries))
	for _, r := range m.entries {
		if r.cancel != nil {
			entries = append(entries, r)
		}
	}
	m.mu.Unlock()
	var errs []error
	for _, r := range entries {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
