package core

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
