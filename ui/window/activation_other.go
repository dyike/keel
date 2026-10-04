//go:build !linux || android

package window

func activationWindowEvent(w *Window, e any)        {}
func platformActivate(w *Window, token string) bool { return false }
