//go:build darwin && !ios && cgo && !nometal

package window

/*
#cgo LDFLAGS: -framework AppKit -framework QuartzCore -framework Metal
#include "glass_darwin.h"
*/
import "C"

import (
	"errors"
	"image"
	"image/color"

	gioapp "gioui.org/app"
	"gioui.org/gpu"
	"gioui.org/op"
)

func platformGlassSupported() bool       { return C.keel_glass_supported() != 0 }
func platformLiquidGlassSupported() bool { return C.keel_liquid_glass_supported() != 0 }

// View events and rendering run on the window goroutine, outside the frame
// lock. AppKit installation/teardown is asynchronous to avoid blocking its
// event loop while it is delivering a Gio event.
func platformGlassEvent(w *Window, event any) {
	e, ok := event.(gioapp.AppKitViewEvent)
	if !ok || w.opts.Glass == nil || !GlassSupported() {
		return
	}
	if w.glassRenderer != nil {
		w.glassRenderer.Release()
		w.glassRenderer = nil
	}
	if e.View != 0 {
		o := w.opts.Glass
		w.glassRenderer = &metalGlassRenderer{
			context:    C.keel_glass_create(C.uintptr_t(e.View), C.int(o.Style), C.double(o.CornerRadius)),
			invalidate: w.win.Invalidate,
		}
	}
}

type metalGlassRenderer struct {
	context    *C.keel_glass_context
	gpu        gpu.GPU
	invalidate func()
}

func (r *metalGlassRenderer) Frame(ops *op.Ops, size image.Point) error {
	if r.context == nil {
		return errors.New("could not allocate Metal backdrop")
	}
	if r.gpu == nil {
		var api C.keel_glass_api
		switch C.keel_glass_ready(r.context, &api) {
		case 0:
			r.invalidate()
			return nil
		case -1:
			return errors.New("could not initialize Metal backdrop")
		}
		var err error
		r.gpu, err = gpu.New(gpu.Metal{Device: uintptr(api.device), Queue: uintptr(api.queue), PixelFormat: int(api.pixel_format)})
		if err != nil {
			return err
		}
	}
	if size.X <= 0 || size.Y <= 0 {
		return nil
	}
	drawable := C.keel_glass_drawable(r.context, C.int(size.X), C.int(size.Y))
	if drawable == 0 {
		// Occluded/minimized windows can temporarily have no drawable.
		return nil
	}
	r.gpu.Clear(color.NRGBA{})
	err := r.gpu.Frame(ops, gpu.MetalRenderTarget{Texture: uintptr(C.keel_glass_texture(drawable))}, size)
	var present C.int
	if err == nil {
		present = 1
	}
	C.keel_glass_present(r.context, drawable, present)
	return err
}

func (r *metalGlassRenderer) Release() {
	if r.gpu != nil {
		r.gpu.Release()
		r.gpu = nil
	}
	if r.context != nil {
		C.keel_glass_release(r.context)
		r.context = nil
	}
}
