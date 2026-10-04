//go:build windows

package window

import (
	"sync"

	"golang.org/x/sys/windows/registry"

	"github.com/dyike/keel/ui/el"
)

var preferencesOnce sync.Once

// watchSystemPreferences reads "Automatically hide scroll bars" (Settings,
// Accessibility, Visual effects) once; Windows 11 turns it on by default.
func watchSystemPreferences() {
	preferencesOnce.Do(func() {
		k, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Accessibility`, registry.QUERY_VALUE)
		if err != nil {
			return
		}
		defer k.Close()
		if v, _, err := k.GetIntegerValue("DynamicScrollbars"); err == nil && v != 0 {
			el.SetSystemScrollbars(el.ScrollbarScrolling)
		}
	})
}
