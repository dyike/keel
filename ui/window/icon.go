package window

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"sync"
)

// The app icon while the app runs. A packaged app (keel build) carries its
// icon in the bundle or executable, but go run and keel run start a bare
// executable that the system shows with a generic icon. SetIcon fills that
// gap at run time, cutting the artwork into each platform's shape the same
// way keel build does (internal/appicon).

var appIcon struct {
	sync.Mutex
	art image.Image
}

// SetIcon sets the app's icon from full-bleed square PNG artwork, as in a
// keel project's appicon.png: macOS shows it in the Dock with Apple's plate
// and shadow; Windows on every window's title bar and taskbar button; Linux
// X11 on every window (_NET_WM_ICON). Wayland and the browser take their
// icon from the installed .desktop entry and the page, so SetIcon does
// nothing there. It applies to open windows and to those opened later; call
// it before Open, or any time to change the icon.
func SetIcon(artwork []byte) error {
	img, err := png.Decode(bytes.NewReader(artwork))
	if err != nil {
		return fmt.Errorf("window.SetIcon: not a PNG: %w", err)
	}
	if b := img.Bounds(); b.Dx() != b.Dy() || b.Dx() == 0 {
		return errors.New("window.SetIcon: the artwork must be square")
	}
	appIcon.Lock()
	appIcon.art = img
	appIcon.Unlock()
	if !offScreen() {
		platformSetIcon(img)
	}
	return nil
}

func currentIcon() image.Image {
	appIcon.Lock()
	defer appIcon.Unlock()
	return appIcon.art
}

func encodePNG(img image.Image) []byte {
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}
