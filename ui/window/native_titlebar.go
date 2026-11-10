package window

import "runtime"

// NativeTitleBarInset reports the space reserved for AppKit's native buttons.
// Other platforms, system-framed windows and drawn controls return zero.
func (w *Window) NativeTitleBarInset() float32 {
	if runtime.GOOS != "darwin" || !w.opts.Frameless || !w.opts.NativeTrafficLights {
		return 0
	}
	left, spacing := float32(12), float32(20)
	if l := w.opts.TrafficLightLayout; l != nil {
		left = l.Left
		if l.Spacing > 0 {
			spacing = l.Spacing
		}
	}
	// Reserve 16dp for the system button and 8dp clear space after it.
	return left + 2*spacing + 24
}
