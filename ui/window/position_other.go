//go:build !darwin || ios

package window

import "github.com/dyike/keel/third_party/gio/io/system"

func centerNewWindow(w *Window) {
	w.perform(system.ActionCenter)
}
