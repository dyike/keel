//go:build ios && cgo

package window

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework UIKit
int keel_prepareIOSScene(void);
*/
import "C"

func init() {
	if C.keel_prepareIOSScene() != 0 {
		panic("window: Gio's iOS application delegate is incompatible with KeelSceneDelegate")
	}
}
