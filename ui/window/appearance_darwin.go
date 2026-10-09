//go:build darwin && !ios

package window

import "github.com/dyike/keel/ui/internal/appkit"

func platformSystemAppearance() Appearance {
	dark := false
	appkit.Pool(func() {
		defaults := appkit.Send(appkit.Class("NSUserDefaults"), "standardUserDefaults")
		dark = appkit.Equal(appkit.Send(defaults, "stringForKey:", uintptr(appkit.String("AppleInterfaceStyle"))), "Dark")
	})
	if dark {
		return AppearanceDark
	}
	return AppearanceLight
}
func platformNativeAppearanceSupported() bool { return true }
func platformSetNativeAppearance(value Appearance) {
	appkit.MainAsync(func() {
		app := appkit.App()
		if app == 0 {
			return
		}
		var appearance appkit.ID
		switch value {
		case AppearanceLight:
			appearance = appkit.Send(appkit.Class("NSAppearance"), "appearanceNamed:", uintptr(appkit.Constant("NSAppearanceNameAqua")))
		case AppearanceDark:
			appearance = appkit.Send(appkit.Class("NSAppearance"), "appearanceNamed:", uintptr(appkit.Constant("NSAppearanceNameDarkAqua")))
		}
		appkit.Send(app, "setAppearance:", uintptr(appearance))
	})
}
