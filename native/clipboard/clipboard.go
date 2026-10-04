// Package clipboard reads a bounded snapshot of text, encoded images and file
// paths without linking to the UI. macOS uses AppKit; Windows uses Win32; Linux
// uses Wayland when given a display (UseWaylandDisplay), else X11.
package clipboard

import (
	"encoding/json"
	"fmt"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
)

// Image contains owned encoded image bytes, without decoding pixel data.
type Image struct {
	MIME string `json:"mime"`
	Data []byte `json:"data"`
}

// Data is one clipboard snapshot. Files are paths, not opened file contents.
type Data struct {
	Text   string   `json:"text"`
	Images []Image  `json:"images"`
	Files  []string `json:"files"`
}

// Read asynchronously reads the clipboard on the platform's required thread.
// done runs once on a background goroutine. UI callers must dispatch UI changes.
// macOS supports PNG/TIFF and file URLs, up to 128 items and 16MiB encoded data.
// Windows supports Unicode text, PNG or packed DIB (returned as image/bmp),
// and file paths, with the same limits. Linux reads text, file URIs and
// PNG/JPEG/TIFF/BMP/WebP over Wayland or X11. Other platforms return
// native.ErrUnsupported. Nil done is ignored.
func Read(done func(Data, error)) {
	if done == nil {
		return
	}
	finish := func(raw []byte, err error) {
		if err != nil {
			done(Data{}, err)
			return
		}
		var data Data
		if err = json.Unmarshal(raw, &data); err != nil {
			done(Data{}, fmt.Errorf("%w: clipboard snapshot: %v", native.ErrFailed, err))
			return
		}
		done(data, nil)
	}
	if waylandDisplay.Load() != nil {
		go func() {
			if raw, ok, err := readWayland(); ok {
				finish(raw, err)
			} else {
				sys.ClipboardRead(finish)
			}
		}()
		return
	}
	sys.ClipboardRead(finish)
}
