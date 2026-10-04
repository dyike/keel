package core

// ClipboardImage is owned encoded image data provided to a paste handler.
type ClipboardImage struct {
	MIME string
	Data []byte
}

// ClipboardData holds the available text, images and file paths in one paste.
// Paths are references only; the UI does not open the files automatically.
type ClipboardData struct {
	Text   string
	Images []ClipboardImage
	Files  []string
}

// ClipboardReader starts an asynchronous read. It must call done once, on any
// goroutine; input components marshal completion back to the UI loop. An error
// falls back to Gio text paste. Applications can adapt native/clipboard.Read.
type ClipboardReader func(done func(ClipboardData, error))
