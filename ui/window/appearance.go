package window

import "fmt"

// Appearance describes system preference or an explicit native chrome style.
// Application palette selection remains independent, e.g. theme.Apply(custom).
type Appearance string

const (
	AppearanceSystem Appearance = "system"
	AppearanceLight  Appearance = "light"
	AppearanceDark   Appearance = "dark"
)

// SystemAppearance reads the platform preference, falling back to Light when
// unavailable. Linux portal updates are cached so rendering never waits on D-Bus.
func SystemAppearance() Appearance { return platformSystemAppearance() }

// NativeAppearanceSupported reports whether SetNativeAppearance can change
// native application chrome. Go-rendered content still uses the application's theme.
func NativeAppearanceSupported() bool { return platformNativeAppearanceSupported() }

// SetNativeAppearance chooses native chrome without replacing a custom Go
// palette. System restores OS-controlled appearance. Unsupported platforms keep
// their existing native chrome; query NativeAppearanceSupported if needed.
func SetNativeAppearance(value Appearance) error {
	switch value {
	case AppearanceSystem, AppearanceLight, AppearanceDark:
	default:
		return fmt.Errorf("window: invalid appearance %q", value)
	}
	if !offScreen() {
		platformSetNativeAppearance(value)
	}
	return nil
}
