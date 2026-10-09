package window

import (
	gioapp "gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"image"
	"testing"
	"time"
)

func TestWindowSizeReportsClientDPAndChanges(t *testing.T) {
	var sizes []image.Point
	w := newWindow(Options{Width: 800, Height: 600, MenuDisplay: MenuDisplayWindow, OnResize: func(width, height int) { sizes = append(sizes, image.Pt(width, height)) }})
	if width, height := w.Size(); width != 800 || height != 600 {
		t.Fatal(width, height)
	}
	var ops op.Ops
	frame := func(width, height int, scale float32) {
		ops.Reset()
		w.layout(layout.Context{Ops: &ops, Now: time.Now(), Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale}, Constraints: layout.Exact(image.Pt(width, height))})
	}
	frame(1600, 1200, 2)
	frame(1600, 1200, 2)
	frame(1800, 1000, 2)
	if width, height := w.Size(); width != 900 || height != 500 {
		t.Fatal(width, height)
	}
	if len(sizes) != 2 || sizes[0] != image.Pt(800, 600) || sizes[1] != image.Pt(900, 500) {
		t.Fatal(sizes)
	}
}
func TestResizeVirtualWindowAndIgnoreClosed(t *testing.T) {
	w := openTest(t, Options{Width: 800, Height: 600, Content: core.Func(func(gtx core.C) core.D { return core.D{Size: gtx.Constraints.Max} })})
	w.render()
	if err := w.Resize(900, 500); err != nil {
		t.Fatal(err)
	}
	if err := w.Resize(960, 540); err != nil {
		t.Fatal(err)
	}
	w.render()
	if width, height := w.Size(); width != 960 || height != 540 {
		t.Fatal(width, height)
	}
	if err := w.Resize(0, 540); err == nil {
		t.Fatal("invalid size accepted")
	}
	w.closed = true
	if err := w.Resize(1000, 800); err != nil {
		t.Fatal(err)
	}
	w.render()
	if width, height := w.Size(); width != 960 || height != 540 {
		t.Fatal("closed window resized", width, height)
	}
}

func TestWindowMinimumSizeClampsInitialAndResize(t *testing.T) {
	w := openTest(t, Options{Width: 300, Height: 200, MinWidth: 800, MinHeight: 480})
	if width, height := w.Size(); width != 800 || height != 480 {
		t.Fatal("initial size below minimum", width, height)
	}
	// Verify the native driver receives the limits at both common pixel scales.
	for _, scale := range []float32{1, 2} {
		var config gioapp.Config
		metric := unit.Metric{PxPerDp: scale, PxPerSp: scale}
		for _, option := range w.nativeOptions() {
			option(metric, &config)
		}
		if config.MinSize != image.Pt(int(800*scale), int(480*scale)) {
			t.Fatal("native minimum", config.MinSize)
		}
	}
	if err := w.Resize(250, 150); err != nil {
		t.Fatal(err)
	}
	w.render()
	if width, height := w.Size(); width != 800 || height != 480 {
		t.Fatal("resize below minimum", width, height)
	}
	if err := w.Resize(1000, 700); err != nil {
		t.Fatal(err)
	}
	w.render()
	if width, height := w.Size(); width != 1000 || height != 700 {
		t.Fatal("valid resize constrained", width, height)
	}
}
func TestWindowMinimumSingleAxis(t *testing.T) {
	for _, options := range []Options{{MinWidth: 800}, {MinHeight: 600}} {
		w := newWindow(options)
		var config gioapp.Config
		for _, option := range w.nativeOptions() {
			option(unit.Metric{PxPerDp: 1, PxPerSp: 1}, &config)
		}
		if config.MinSize.X != max(1, options.MinWidth) || config.MinSize.Y != max(1, options.MinHeight) {
			t.Fatal(config.MinSize)
		}
	}
}
