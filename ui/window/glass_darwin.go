//go:build darwin && !ios && !nometal

package window

import (
	"errors"
	"image"
	"image/color"
	"sync"
	"sync/atomic"

	gioapp "gioui.org/app"
	"gioui.org/gpu"
	"gioui.org/op"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/ebitengine/purego/objc"
)

const mtlPixelFormatBGRA8UnormSRGB = 81

func platformGlassSupported() bool { return appkit.Class("CAMetalLayer") != 0 }

// Liquid Glass needs macOS 26, where NSGlassEffectView exists.
func platformLiquidGlassSupported() bool { return appkit.Class("NSGlassEffectView") != 0 }

// All input stays with GioView, including clicks on empty glass. Native
// material views are only used for compositing, never as input responders.
var glassClasses = sync.OnceValues(func() (frosted, liquid appkit.ID) {
	hitTest := appkit.Method("hitTest:", func(_ appkit.ID, _ objc.SEL, _ appkit.Point) appkit.ID { return 0 })
	frosted = appkit.RegisterClass("KeelFrostedBackdrop", "NSVisualEffectView", nil, []objc.MethodDef{hitTest})
	if platformLiquidGlassSupported() {
		liquid = appkit.RegisterClass("KeelLiquidBackdrop", "NSGlassEffectView", nil, []objc.MethodDef{hitTest})
	}
	return frosted, liquid
})

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
			context:    newGlassContext(e.View, o.Style, float64(o.CornerRadius)),
			invalidate: w.win.Invalidate,
		}
	}
}

// glassContext owns the native objects. ready is 0 while AppKit installs
// them, 1 once installed, -1 on failure; the objects are written before
// ready is set and only released on the main queue after Release.
type glassContext struct {
	ready                                atomic.Int32
	view, backdrop, layer, device, queue appkit.ID
}

const (
	nsViewWidthSizable     = 1 << 1
	nsViewHeightSizable    = 1 << 4
	nsMaterialSidebar      = 7
	nsBlendingBehindWindow = 0
	nsStateFollowsWindow   = 0
	nsGlassEffectRegular   = 0
	nsGlassEffectClear     = 1
)

func newGlassContext(handle uintptr, style GlassStyle, radius float64) *glassContext {
	ctx := &glassContext{view: appkit.Retain(appkit.ID(handle))}
	appkit.MainAsync(func() {
		view := ctx.view
		window := appkit.Send(view, "window")
		if window == 0 || !appkit.SendBool(appkit.Send(view, "layer"), "isKindOfClass:", uintptr(appkit.Class("CAMetalLayer"))) {
			ctx.ready.Store(-1)
			return
		}
		ctx.device = appkit.ID(appkit.Call(appkit.Sym("MTLCreateSystemDefaultDevice")))
		if ctx.device != 0 {
			ctx.queue = appkit.Send(ctx.device, "newCommandQueue")
		}
		if ctx.device == 0 || ctx.queue == 0 {
			ctx.ready.Store(-1)
			return
		}
		ctx.layer = appkit.Send(appkit.Send(appkit.Class("CAMetalLayer"), "alloc"), "init")
		appkit.Send(ctx.layer, "setDevice:", uintptr(ctx.device))
		appkit.Send(ctx.layer, "setPixelFormat:", mtlPixelFormatBGRA8UnormSRGB)
		appkit.Send(ctx.layer, "setOpaque:", 0)
		appkit.Send(ctx.layer, "setPresentsWithTransaction:", 0)
		appkit.MsgSetFloat(ctx.layer, appkit.Sel("setContentsScale:"), appkit.MsgFloat(window, appkit.Sel("backingScaleFactor")))
		appkit.Send(appkit.Send(view, "layer"), "setOpaque:", 0)
		appkit.Send(window, "setOpaque:", 0)
		appkit.Send(window, "setBackgroundColor:", uintptr(appkit.Send(appkit.Class("NSColor"), "clearColor")))

		bounds := appkit.MsgRect(view, appkit.Sel("bounds"))
		initWithFrame := appkit.Sel("initWithFrame:")
		const resize = nsViewWidthSizable | nsViewHeightSizable
		content := appkit.MsgInitRect(appkit.Send(appkit.Class("NSView"), "alloc"), initWithFrame, bounds)
		appkit.Send(content, "setWantsLayer:", 1)
		appkit.Send(content, "setLayer:", uintptr(ctx.layer))
		appkit.Send(content, "setAutoresizingMask:", resize)

		// Behind-window vibrancy supplies the desktop sampling surface.
		// Liquid Glass then renders its optical treatment above that
		// surface, with Gio inside its documented contentView slot.
		frosted, liquid := glassClasses()
		backdrop := appkit.MsgInitRect(appkit.Send(frosted, "alloc"), initWithFrame, bounds)
		appkit.Send(backdrop, "setMaterial:", nsMaterialSidebar)
		appkit.Send(backdrop, "setBlendingMode:", nsBlendingBehindWindow)
		appkit.Send(backdrop, "setState:", nsStateFollowsWindow)
		appkit.Send(backdrop, "setWantsLayer:", 1)
		if radius > 0 {
			layer := appkit.Send(backdrop, "layer")
			appkit.MsgSetFloat(layer, appkit.Sel("setCornerRadius:"), radius)
			appkit.Send(layer, "setMasksToBounds:", 1)
		}
		if liquid != 0 && style != GlassFrosted {
			glass := appkit.MsgInitRect(appkit.Send(liquid, "alloc"), initWithFrame, bounds)
			s := uintptr(nsGlassEffectRegular)
			if style == GlassClear {
				s = nsGlassEffectClear
			}
			appkit.Send(glass, "setStyle:", s)
			if radius > 0 {
				appkit.MsgSetFloat(glass, appkit.Sel("setCornerRadius:"), radius)
			}
			appkit.Send(glass, "setContentView:", uintptr(content))
			appkit.Send(glass, "setAutoresizingMask:", resize)
			appkit.Send(backdrop, "addSubview:", uintptr(glass))
			appkit.Release(glass)
		} else {
			appkit.Send(backdrop, "addSubview:", uintptr(content))
		}
		appkit.Release(content)
		appkit.Send(backdrop, "setAutoresizingMask:", resize)
		// Keep GioView as NSWindow.contentView: Gio's delegates rely on
		// that identity for input, focus, close and fullscreen events.
		appkit.Send(view, "addSubview:", uintptr(backdrop))
		ctx.backdrop = backdrop
		ctx.ready.Store(1)
	})
	return ctx
}

type metalGlassRenderer struct {
	context    *glassContext
	gpu        gpu.GPU
	invalidate func()
}

func (r *metalGlassRenderer) Frame(ops *op.Ops, size image.Point) error {
	if r.context == nil {
		return errors.New("could not allocate Metal backdrop")
	}
	if r.gpu == nil {
		switch r.context.ready.Load() {
		case 0:
			r.invalidate()
			return nil
		case -1:
			return errors.New("could not initialize Metal backdrop")
		}
		var err error
		r.gpu, err = gpu.New(gpu.Metal{Device: uintptr(r.context.device), Queue: uintptr(r.context.queue), PixelFormat: mtlPixelFormatBGRA8UnormSRGB})
		if err != nil {
			return err
		}
	}
	if size.X <= 0 || size.Y <= 0 {
		return nil
	}
	var err error
	appkit.Pool(func() {
		layer := r.context.layer
		appkit.MsgSetSize(layer, appkit.Sel("setDrawableSize:"), appkit.Size{Width: float64(size.X), Height: float64(size.Y)})
		drawable := appkit.Send(layer, "nextDrawable")
		if drawable == 0 {
			// Occluded/minimized windows can temporarily have no drawable.
			return
		}
		r.gpu.Clear(color.NRGBA{})
		err = r.gpu.Frame(ops, gpu.MetalRenderTarget{Texture: uintptr(appkit.Send(drawable, "texture"))}, size)
		if err == nil {
			buffer := appkit.Send(r.context.queue, "commandBuffer")
			appkit.Send(buffer, "commit")
			appkit.Send(buffer, "waitUntilScheduled")
			appkit.Send(drawable, "present")
		}
	})
	return err
}

func (r *metalGlassRenderer) Release() {
	if r.gpu != nil {
		r.gpu.Release()
		r.gpu = nil
	}
	if ctx := r.context; ctx != nil {
		r.context = nil
		// Enqueued after installation, including when the window closes
		// before installation finishes. Never wait for AppKit with the frame
		// lock held.
		appkit.MainAsync(func() {
			if ctx.backdrop != 0 {
				appkit.Send(ctx.backdrop, "removeFromSuperview")
			}
			for _, obj := range []appkit.ID{ctx.backdrop, ctx.layer, ctx.queue, ctx.device, ctx.view} {
				appkit.Release(obj)
			}
		})
	}
}
