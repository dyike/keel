package window

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
)

// The transparent root must preserve clear pixels while still drawing content.
// Native compositing is checked separately in the desktop fixture.
func TestGlassRootAlpha(t *testing.T) {
	win, err := headless.NewWindow(80, 60)
	if err != nil {
		t.Skip(err)
	}
	defer win.Release()
	for _, transparent := range []bool{false, true} {
		var ops op.Ops
		gtx := layout.Context{Ops: &ops, Now: time.Now(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Exact(image.Pt(80, 60))}
		r := root{transparent: transparent}
		r.Layout(gtx, glassTestContent{})
		if err := win.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rect(0, 0, 80, 60))
		if err := win.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		wantAlpha := uint8(255)
		if transparent {
			wantAlpha = 0
		}
		if got := img.RGBAAt(70, 50).A; got != wantAlpha {
			t.Fatalf("transparent=%v: background alpha %d, want %d", transparent, got, wantAlpha)
		}
		if got := img.RGBAAt(5, 5); got != (color.RGBA{R: 255, A: 255}) {
			t.Fatalf("transparent=%v: content pixel %v", transparent, got)
		}
	}
}

type glassTestContent struct{}

func (glassTestContent) FillsWindow() bool { return true }
func (glassTestContent) Layout(gtx core.C) core.D {
	paint.FillShape(gtx.Ops, color.NRGBA{R: 255, A: 255}, clip.Rect(image.Rect(0, 0, 20, 20)).Op())
	return core.D{Size: gtx.Constraints.Max}
}

func TestGlassOptionsAreCopiedAndValidated(t *testing.T) {
	o := &GlassOptions{Style: GlassClear, CornerRadius: 18}
	w := newWindow(Options{Glass: o})
	o.Style, o.CornerRadius = GlassFrosted, 50
	if *w.opts.Glass != (GlassOptions{Style: GlassClear, CornerRadius: 18}) {
		t.Fatal("caller mutated window options")
	}
	for _, o := range []GlassOptions{{Style: GlassStyle(255)}, {CornerRadius: -1}, {CornerRadius: float32(math.NaN())}, {CornerRadius: float32(math.Inf(1))}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted invalid glass options %+v", o)
				}
			}()
			newWindow(Options{Glass: &o})
		}()
	}
}
