// Package window creates Go-Gui windows and adapts host-owned native windows.
package window

import (
	"unsafe"

	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/internal/platform"
)

type Vibrancy uint8

const (
	VibrancyNone Vibrancy = iota
	VibrancySidebar
	VibrancyHUD
)

// Driver isolates window operations from the application.
type Driver interface {
	SetAlwaysOnTop(bool) error
	SetClickThrough(bool) error
	SetVibrancy(Vibrancy) error
}
type Window struct{ driver Driver }

func New(d Driver) (*Window, error) {
	if d == nil {
		return nil, capability.ErrInvalidArgument
	}
	return &Window{d}, nil
}
func (w *Window) SetAlwaysOnTop(v bool) error  { return w.driver.SetAlwaysOnTop(v) }
func (w *Window) SetClickThrough(v bool) error { return w.driver.SetClickThrough(v) }
func (w *Window) SetVibrancy(v Vibrancy) error {
	if v > VibrancyHUD {
		return capability.ErrInvalidArgument
	}
	return w.driver.SetVibrancy(v)
}

// NativeDriver adapts a host-owned native handle. Dispatch must execute synchronously
// on the UI thread, including when called from that thread. Handle is resolved inside Dispatch.
// Do not call after the host event loop stops. No Go pointer may be returned by Handle.
type NativeDriver struct {
	Handle   func() unsafe.Pointer
	Dispatch func(func())
}

func (d NativeDriver) apply(op string, value int) error {
	if d.Handle == nil || d.Dispatch == nil {
		return capability.ErrInvalidArgument
	}
	var err error
	d.Dispatch(func() {
		h := d.Handle()
		if h == nil {
			err = capability.ErrNotReady
			return
		}
		err = platform.New().Window(h, op, value)
	})
	return err
}
func (d NativeDriver) SetAlwaysOnTop(v bool) error  { return d.apply("top", boolean(v)) }
func (d NativeDriver) SetClickThrough(v bool) error { return d.apply("click", boolean(v)) }
func (d NativeDriver) SetVibrancy(v Vibrancy) error {
	if v > VibrancyHUD {
		return capability.ErrInvalidArgument
	}
	return d.apply("vibrancy", int(v))
}
func boolean(v bool) int {
	if v {
		return 1
	}
	return 0
}

// Show, Hide and Toggle dispatch through the host on its UI thread.
func (w *Window) Show() error {
	d, ok := w.driver.(interface{ Show() error })
	if !ok {
		return capability.ErrUnsupported
	}
	return d.Show()
}
func (w *Window) Hide() error {
	d, ok := w.driver.(interface{ Hide() error })
	if !ok {
		return capability.ErrUnsupported
	}
	return d.Hide()
}
func (w *Window) Toggle() error {
	d, ok := w.driver.(interface{ Toggle() error })
	if !ok {
		return capability.ErrUnsupported
	}
	return d.Toggle()
}
