package window

import (
	"image"
	"math"

	"gioui.org/op"
)

// GlassStyle selects the macOS backdrop material.
type GlassStyle uint8

const (
	GlassRegular GlassStyle = iota // Liquid Glass on macOS 26+, vibrancy on older systems.
	GlassClear                     // More transparent Liquid Glass; vibrancy on older systems.
	GlassFrosted                   // NSVisualEffectView on all supported macOS versions.
)

// GlassOptions configures a glass backdrop covering the window content area.
// The backdrop samples content behind the window, not pixels drawn by Gio.
// Opaque content hides it; transparent content exposes it.
type GlassOptions struct {
	Style GlassStyle
	// CornerRadius is in dp. Zero uses the system's default glass curvature.
	CornerRadius float32
}

func (o GlassOptions) validate() {
	if o.Style > GlassFrosted {
		panic("window: invalid glass style")
	}
	if o.CornerRadius < 0 || math.IsNaN(float64(o.CornerRadius)) || math.IsInf(float64(o.CornerRadius), 0) {
		panic("window: invalid glass corner radius")
	}
}

// GlassSupported reports whether this build supports a native glass backdrop.
// Headless screenshots cannot include the native compositor's effects.
func GlassSupported() bool { return platformGlassSupported() }

// LiquidGlassSupported reports whether this build and OS support
// NSGlassEffectView. Otherwise Glass uses NSVisualEffectView on macOS.
func LiquidGlassSupported() bool { return platformLiquidGlassSupported() }

type glassRenderer interface {
	Frame(*op.Ops, image.Point) error
	Release()
}
