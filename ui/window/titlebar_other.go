//go:build !darwin || ios

package window

func platformWindowEvent(w *Window, e any) { activationWindowEvent(w, e) }
func syncTitleBar(w *Window)               {}
func clearPlatformWindow(w *Window)        {}
