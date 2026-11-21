package core

import "image/color"

// WindowControls is what UI code may ask of the window it is drawn in: a
// custom title bar uses it for its buttons. ui/window implements it.
type WindowControls interface {
	// Frameless reports whether the window draws its own title bar.
	Frameless() bool
	// Focused reports native window activation, not an individual control focus.
	Focused() bool
	// TitleBarArea registers the current draggable title region in window dp.
	// The window clears it each frame; controls must be outside this rectangle.
	TitleBarArea(x, y, width, height float32)
	Minimize()
	// ToggleMaximize maximizes the window, or restores it when maximized.
	ToggleMaximize()
	Maximized() bool
	Close()
}

var currentWindow WindowControls // guarded by the frame lock

// CurrentWindow returns the window being laid out, or nil outside one (a
// screenshot, a test harness). Read it in Render or Layout, under the frame
// lock, and keep it for callbacks; callbacks run within the same window.
func CurrentWindow() WindowControls { return currentWindow }

// SetCurrentWindow marks w as the window being laid out and returns a func
// that restores the previous one. Only ui/window calls it.
func SetCurrentWindow(w WindowControls) (restore func()) {
	prev := currentWindow
	currentWindow = w
	return func() { currentWindow = prev }
}

// NativeTitleBarControls is implemented by windows with OS-owned titlebar
// buttons. The inset is the reserved extent from the window's left edge in dp;
// zero means the application should draw its own controls.
type NativeTitleBarControls interface {
	NativeTitleBarInset() float32
}

// GlassRegionStyle selects the native Liquid Glass surface style for a control region.
type GlassRegionStyle uint8

const (
	GlassRegionRegular GlassRegionStyle = iota
	GlassRegionClear
)

// GlassRegion describes a control-level native Liquid Glass surface in window dp.
type GlassRegion struct {
	X, Y, Width, Height float32
	CornerRadius        float32
	Style               GlassRegionStyle
	Tint                color.NRGBA
	Interactive         bool
	Spacing             float32 // NSGlassEffectContainerView merge proximity in dp (0 = default)
}

// NativeGlassControls is implemented by windows that render control-level
// Liquid Glass with native compositor views (e.g. macOS 26+ NSGlassEffectView
// and NSGlassEffectContainerView). Regions are cleared each frame before layout.
type NativeGlassControls interface {
	// NativeGlassEnabled reports whether native control-level Liquid Glass is active.
	NativeGlassEnabled() bool
	// AddGlassRegion registers a native glass surface for the current frame.
	AddGlassRegion(r GlassRegion)
}
