package clipboard

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/dyike/keel/native"
	"github.com/dyike/keel/native/internal/sys"
	"github.com/dyike/keel/native/internal/wlclip"
)

var waylandDisplay atomic.Pointer[byte]

// UseWaylandDisplay makes Read use a Wayland connection, the wl_display of
// the application's focused window (window.Window.WaylandDisplay). Wayland
// shows the clipboard only to the focused client, so Read cannot open a
// connection of its own. Nil, or a display without a data device, falls
// back to X11 (XWayland). Pass the display again when focus moves to another
// window; it must stay open until Read completes.
func UseWaylandDisplay(display unsafe.Pointer) { waylandDisplay.Store((*byte)(display)) }

// readWayland reports ok false when Read should fall back to X11.
func readWayland() (raw []byte, ok bool, err error) {
	display := unsafe.Pointer(waylandDisplay.Load())
	if display == nil {
		return nil, false, nil
	}
	sel, err := wlclip.Open(display)
	if errors.Is(err, native.ErrUnsupported) {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, fmt.Errorf("%w: Wayland clipboard", native.ErrFailed)
	}
	defer sel.Close()
	targets := map[string]bool{}
	for _, mime := range sel.MIMEs() {
		targets[mime] = true
	}
	deadline := time.Now().Add(5 * time.Second)
	raw, err = sys.CollectMIMEClipboard(targets, func(mime string) ([]byte, error) {
		return sel.Read(mime, sys.ClipboardByteLimit, deadline)
	})
	return raw, true, err
}
