//go:build (!darwin || ios || !cgo) && !windows

package window

func watchSystemPreferences() {}
