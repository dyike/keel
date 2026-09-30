package keel

import (
	"github.com/go-gui-org/go-gui/gui"
	"github.com/dyike/keel/capability"
	"github.com/dyike/keel/ui"
	"github.com/dyike/keel/window"
)

type Vibrancy = window.Vibrancy

const (
	VibrancyNone    = window.VibrancyNone
	VibrancySidebar = window.VibrancySidebar
	VibrancyHUD     = window.VibrancyHUD
)

// OverlayOptions defines an initially hidden utility window. Closing hides it.
// AlwaysOnTop is currently unsupported by the Go-Gui backend.
type OverlayOptions struct {
	Title                               string
	Width, Height                       int
	UI                                  *ui.Page
	AlwaysOnTop, Frameless, Transparent bool
	Vibrancy                            Vibrancy
	OnReady                             func(*window.Window)
	OnError                             func(error)
}

func (k *Kit) NewOverlay(o OverlayOptions) (*window.Window, error) {
	if o.AlwaysOnTop {
		return nil, capability.ErrUnsupported
	}
	if o.Vibrancy > VibrancyHUD {
		return nil, capability.ErrInvalidArgument
	}
	w, err := k.Window.New(window.Options{Title: o.Title, Width: o.Width, Height: o.Height, UI: o.UI, Hidden: true, Frameless: o.Frameless, Transparent: o.Transparent, GUI: gui.WindowCfg{OnCloseRequest: func(host *gui.Window) { host.Hide() }}})
	if err != nil {
		return nil, err
	}
	_ = w.OnReady(func() {
		if o.Vibrancy != VibrancyNone {
			if err := w.SetVibrancy(o.Vibrancy); err != nil && o.OnError != nil {
				o.OnError(err)
			}
		}
		if o.OnReady != nil {
			o.OnReady(w)
		}
	})
	return w, nil
}
