package window

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"time"

	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/unit"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
)

// Screenshot renders content as a window would, off-screen at 2× scale, and
// writes a PNG. Width and height are in dp.
func Screenshot(content core.Widget, width, height int, path string) error {
	return ScreenshotAtScale(content, width, height, 2, path)
}

// ScreenshotAtScale renders the first frame at an explicit pixel density.
// Dimensions are in dp; scale must be finite and positive.
func ScreenshotAtScale(content core.Widget, width, height int, scale float32, path string) error {
	size, err := screenshotSize(width, height, scale)
	if err != nil {
		return err
	}
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

func screenshotSize(width, height int, scale float32) (image.Point, error) {
	x, y := math.Round(float64(width)*float64(scale)), math.Round(float64(height)*float64(scale))
	maxInt := float64(int(^uint(0) >> 1))
	if width <= 0 || height <= 0 || scale <= 0 || math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || x < 1 || y < 1 || x >= maxInt || y >= maxInt {
		return image.Point{}, fmt.Errorf("window: invalid screenshot dimensions %dx%d at %g scale", width, height, scale)
	}
	return image.Pt(int(x), int(y)), nil
}
