//go:build !darwin || ios

package window

import "gioui.org/io/system"

func centerNewWindow(w *Window) {
	w.perform(system.ActionCenter)
}
