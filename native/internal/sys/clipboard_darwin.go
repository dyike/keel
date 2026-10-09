//go:build darwin && !ios

package sys

import "encoding/json"

// ClipboardRead snapshots NSPasteboard on the main queue; done runs on its
// own goroutine.
func ClipboardRead(done func([]byte, error)) {
	mainAsync(func() {
		data, code := clipboardSnapshotDarwin()
		go done(data, status(code))
	})
}

// Main thread only, inside an autorelease pool.
func clipboardSnapshotDarwin() ([]byte, int) {
	const failed = 7
	board := send(class("NSPasteboard"), "generalPasteboard")
	version := send(board, "changeCount")
	items := send(board, "pasteboardItems")
	count := uintptr(send(items, "count"))
	if count > clipboardItemLimit {
		return nil, failed
	}
	var out clipboardSnapshot
	out.Text = goString(send(board, "stringForType:", uintptr(constant("NSPasteboardTypeString"))))
	out.Images, out.Files = []clipboardImage{}, []string{}
	size := len(out.Text)
	fileURL, png, tiff := constant("NSPasteboardTypeFileURL"), constant("NSPasteboardTypePNG"), constant("NSPasteboardTypeTIFF")
	for i := uintptr(0); i < count; i++ {
		item := send(items, "objectAtIndex:", i)
		if text := send(item, "stringForType:", uintptr(fileURL)); text != 0 {
			url := send(class("NSURL"), "URLWithString:", uintptr(text))
			if url != 0 && sendBool(url, "isFileURL") {
				if path := goString(send(url, "path")); path != "" {
					size += len(path)
					out.Files = append(out.Files, path)
					// Finder may attach an image preview; preserve the file identity.
					continue
				}
			}
		}
		mime, data := "image/png", send(item, "dataForType:", uintptr(png))
		if data == 0 {
			mime, data = "image/tiff", send(item, "dataForType:", uintptr(tiff))
		}
		if data != 0 {
			size += int(send(data, "length"))
			if size > clipboardByteLimit {
				return nil, failed
			}
			out.Images = append(out.Images, clipboardImage{MIME: mime, Data: goBytes(data)})
		}
	}
	if size > clipboardByteLimit || version != send(board, "changeCount") {
		return nil, failed
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, failed
	}
	return raw, 0
}
