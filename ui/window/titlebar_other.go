//go:build !darwin || ios || !cgo

package window

func platformWindowEvent(w *Window, e any) { activationWindowEvent(w, e) }
func syncTitleBar(w *Window)               {}
func clearPlatformWindow(w *Window)        {}
