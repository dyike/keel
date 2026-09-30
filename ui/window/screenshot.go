package window

import (
	"image"
	"image/png"
	"os"
	"time"

	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
)

// Screenshot renders content as a window would, off-screen at 2× scale, and
// writes a PNG. Width and height are in dp.
func Screenshot(content core.Widget, width, height int, path string) error {
	const scale = 2
	size := image.Pt(width*scale, height*scale)
	win, err := headless.NewWindow(size.X, size.Y)
	if err != nil {
		return err
	}
	defer win.Release()
	var (
		ops    op.Ops
		r      root
		router input.Router // a real source keeps widgets in their enabled look
	)
	gtx := layout.Context{Ops: &ops, Now: time.Now(), Source: router.Source(),
		Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale}}
	loop.Lock()
	r.Layout(gtx, content)
	loop.Unlock()
	if err := win.Frame(&ops); err != nil {
		return err
	}
	img := image.NewRGBA(image.Rectangle{Max: size})
	if err := win.Screenshot(img); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
