//go:build !darwin || ios

package window

import "github.com/dyike/keel/gio/io/system"

func centerNewWindow(w *Window) {
	w.perform(system.ActionCenter)
}
