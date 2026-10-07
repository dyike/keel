//go:build windows

package window

import "golang.org/x/sys/windows/registry"

func platformSystemAppearance() Appearance {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return AppearanceLight
	}
	defer key.Close()
	if light, _, err := key.GetIntegerValue("AppsUseLightTheme"); err == nil && light == 0 {
		return AppearanceDark
	}
	return AppearanceLight
}
func platformNativeAppearanceSupported() bool { return false }
func platformSetNativeAppearance(Appearance)  {}
