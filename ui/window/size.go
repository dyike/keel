package window

import (
	"fmt"
	"image"
	"math"

	gioapp "gioui.org/app"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
)

// Size returns the most recently observed client-area width and height in dp.
// Before the first frame it returns the requested initial size. It is safe from
// callbacks and background goroutines, and includes window menus/decorations
// drawn by Keel. The operating system's external frame is excluded.
func (w *Window) Size() (width, height int) {
	value := w.size.Load()
	return int(value >> 32), int(uint32(value))
}
func (w *Window) storeSize(width, height int) {
	w.size.Store(uint64(uint32(width))<<32 | uint64(uint32(height)))
}

// Resize requests a client-area size in dp. It is safe from callbacks and
// background goroutines. The platform may constrain the request; Size and
// OnResize report the actual size on the next frame. Closed windows are ignored.
func (w *Window) Resize(width, height int) error {
	if width <= 0 || height <= 0 || uint64(width) > math.MaxInt32 || uint64(height) > math.MaxInt32 {
		return fmt.Errorf("window: invalid size %dx%d", width, height)
	}
	width = max(width, w.opts.MinWidth)
	height = max(height, w.opts.MinHeight)
	version := w.resizeVersion.Add(1)
	core.Update(func() {
		if w.closed || w.resizeVersion.Load() != version {
			return
		}
		if w.win == nil {
			if w.virt != nil {
				w.virt.setSize(image.Pt(width, height))
			}
			return
		}
		// Gio may wait for its native thread; never call Option under the frame lock.
		go func() {
			w.resizeSerial.Lock()
			defer w.resizeSerial.Unlock()
			if w.resizeVersion.Load() != version || w.isClosed() {
				return
			}
			w.win.Option(gioapp.Size(unit.Dp(width), unit.Dp(height)))
		}()
	})
	return nil
}
func (w *Window) observeSize(gtx core.C) {
	size := image.Pt(int(math.Round(float64(gtx.Metric.PxToDp(gtx.Constraints.Max.X)))), int(math.Round(float64(gtx.Metric.PxToDp(gtx.Constraints.Max.Y)))))
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	w.storeSize(size.X, size.Y)
	if size == w.lastSize {
		return
	}
	w.lastSize = size
	if callback := w.opts.OnResize; callback != nil {
		core.Call(gtx, func() { callback(size.X, size.Y) })
	}
}
