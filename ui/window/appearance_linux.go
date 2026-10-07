//go:build linux && !android

package window

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/godbus/dbus/v5"
)

var portalAppearance struct {
	once sync.Once
	dark atomic.Bool
}

func platformSystemAppearance() Appearance {
	portalAppearance.once.Do(func() { go watchPortalAppearance() })
	if portalAppearance.dark.Load() {
		return AppearanceDark
	}
	return AppearanceLight
}
func watchPortalAppearance() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		if conn, err := dbus.SessionBus(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			var value dbus.Variant
			err = conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop").CallWithContext(ctx, "org.freedesktop.portal.Settings.Read", 0, "org.freedesktop.appearance", "color-scheme").Store(&value)
			cancel()
			if err == nil {
				portalAppearance.dark.Store(portalPrefersDark(value.Value()))
			}
		}
		<-ticker.C
	}
}
func portalPrefersDark(value any) bool {
	if variant, ok := value.(dbus.Variant); ok {
		value = variant.Value()
	}
	scheme, ok := value.(uint32)
	return ok && scheme == 1
}
func platformNativeAppearanceSupported() bool { return false }
func platformSetNativeAppearance(Appearance)  {}
