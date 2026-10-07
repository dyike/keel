package window

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"sync"

	"github.com/dyike/keel/internal/appicon"
	"golang.org/x/image/draw"
)

// The app icon while the app runs. A packaged app (keel build) carries its
// icon in the bundle or executable. keel run supplies a finished icon at
// runtime; plain go run can use SetIcon with artwork shaped as keel build does.

var appIcon struct {
	sync.Mutex
	art      image.Image
	finished bool
	runOnce  sync.Once
}

// SetIcon sets the app's icon from full-bleed square PNG artwork, as in a
// keel project's appicon.png: macOS shows it in the Dock with Apple's plate
// and shadow; Windows on every window's title bar and taskbar button; Linux
// X11 on every window (_NET_WM_ICON). Wayland and the browser take their
// icon from the installed .desktop entry and the page, so SetIcon does
// nothing there. It applies to open windows and to those opened later; call
// it before Open, or any time to change the icon.
func SetIcon(artwork []byte) error { return setIcon(artwork, false) }

func setIcon(artwork []byte, finished bool) error {
	img, err := png.Decode(bytes.NewReader(artwork))
	if err != nil {
		return fmt.Errorf("window.SetIcon: not a PNG: %w", err)
	}
	if b := img.Bounds(); b.Dx() != b.Dy() || b.Dx() == 0 {
		return errors.New("window.SetIcon: the artwork must be square")
	}
	appIcon.Lock()
	appIcon.art = img
	appIcon.finished = finished
	appIcon.Unlock()
	if !offScreen() {
		platformSetIcon(img, finished)
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

// Only windows load the CLI icon: helper/server processes must not start AppKit.
func loadRunIcon() {
	appIcon.runOnce.Do(func() {
		path := os.Getenv("KEEL_RUN_ICON")
		if path == "" || currentIcon() != nil {
			return
		}
		if err := readRunIcon(path); err != nil {
			log.Printf("keel: development icon: %v", err)
		}
	})
}

func readRunIcon(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return setIcon(data, true)
}

func currentIconState() (image.Image, bool) {
	appIcon.Lock()
	defer appIcon.Unlock()
	return appIcon.art, appIcon.finished
}

// CLI icons already include the platform outline and shadow. Resize them
// directly instead of applying the platform template a second time.
func renderIcon(art image.Image, shape appicon.Shape, size int, finished bool) *image.NRGBA {
	if !finished {
		return shape.Render(art, size, true)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(dst, dst.Bounds(), art, art.Bounds(), draw.Src, nil)
	return dst
}
