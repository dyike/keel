// Package clipboard reads a bounded snapshot of text, encoded images and file
// paths without linking to the UI. macOS uses AppKit's general pasteboard.
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
// Other platforms currently return native.ErrUnsupported. Nil done is ignored.
func Read(done func(Data, error)) {
	if done == nil {
		return
	}
	sys.ClipboardRead(func(raw []byte, err error) {
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
	})
}
