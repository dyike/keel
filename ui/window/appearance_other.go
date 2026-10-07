//go:build !(darwin && !ios && cgo) && !windows && !(linux && !android)

package window

func platformSystemAppearance() Appearance    { return AppearanceLight }
func platformNativeAppearanceSupported() bool { return false }
func platformSetNativeAppearance(Appearance)  {}
