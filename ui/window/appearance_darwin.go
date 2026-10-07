//go:build darwin && !ios && cgo

package window

/*
#cgo LDFLAGS: -framework AppKit
int keel_system_dark(void);
void keel_native_appearance(int mode);
*/
import "C"

func platformSystemAppearance() Appearance {
	if C.keel_system_dark() != 0 {
		return AppearanceDark
	}
	return AppearanceLight
}
func platformNativeAppearanceSupported() bool { return true }
func platformSetNativeAppearance(value Appearance) {
	mode := 0
	if value == AppearanceLight {
		mode = 1
	}
	if value == AppearanceDark {
		mode = 2
	}
	C.keel_native_appearance(C.int(mode))
}
