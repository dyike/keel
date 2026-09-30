package el

import (
	"io"
	"strings"

	"gioui.org/io/clipboard"

	"github.com/dyike/keel/ui/core"
)

// pendingClipboard holds text to copy at the next frame. Guarded by the UI
// lock: handlers and Render run under it.
var pendingClipboard *string

// WriteClipboard copies text to the system clipboard. Call it from a handler,
// e.g. a copy button's OnClick; it takes effect in the current frame.
func WriteClipboard(text string) { pendingClipboard = &text }

func flushClipboard(gtx core.C) {
	if pendingClipboard == nil {
		return
	}
	text := *pendingClipboard
	pendingClipboard = nil
	gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
}
