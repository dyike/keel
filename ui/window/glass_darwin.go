//go:build darwin && !ios && !nometal

package window

import (
	"errors"
	"image"
	"image/color"
	"sync"
	"sync/atomic"

	gioapp "github.com/dyike/keel/gio/app"
	"github.com/dyike/keel/gio/gpu"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/ui/internal/appkit"
)

const mtlPixelFormatRGBA16Float = 115

func platformGlassSupported() bool { return appkit.Class("CAMetalLayer") != 0 }

// Liquid Glass needs macOS 26, where NSGlassEffectView exists.
func platformLiquidGlassSupported() bool { return appkit.Class("NSGlassEffectView") != 0 }

// The native effect is the window root, as in a standard AppKit glass
// window. Its ordinary hit testing routes input to the full-size GioView.
var glassClasses = sync.OnceValues(func() (frosted, liquid appkit.ID) {
	return appkit.Class("NSVisualEffectView"), appkit.Class("NSGlassEffectView")
})

// View events and rendering run on the window goroutine, outside the frame
// lock. AppKit installation/teardown is asynchronous to avoid blocking its
// event loop while it is delivering a Gio event.
func platformGlassEvent(w *Window, event any) {
	e, ok := event.(gioapp.AppKitViewEvent)
	if !ok || w.opts.Glass == nil || !GlassSupported() {
		return
	}
	if renderer, ok := w.glassRenderer.(*metalGlassRenderer); ok && renderer.context != nil {
		ctx := renderer.context
		// Moving the existing GioView into its native wrapper emits temporary
		// detach/attach events. These do not replace the view or its renderer.
		if ctx.installing.Load() || (ctx.ready.Load() == 1 && e.View == uintptr(ctx.view) && e.Layer == uintptr(ctx.layer)) {
			return
		}
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
	installing                           atomic.Bool
	view, backdrop, layer, device, queue appkit.ID
}

const (
	nsViewWidthSizable     = 1 << 1
	nsViewHeightSizable    = 1 << 4
	nsMaterialSidebar      = 7
	nsBlendingBehindWindow = 0
	nsStateActive          = 1
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
		ctx.layer = appkit.Retain(appkit.Send(view, "layer"))
		appkit.Send(ctx.layer, "setDevice:", uintptr(ctx.device))
		// Gio outputs linear premultiplied RGB. Let Core Animation convert the
		// color space after compositing; an sRGB target encodes RGB before the
		// alpha divide and washes translucent content out to white.
		// Half floats retain dark-tone precision in this linear surface.
		appkit.Send(ctx.layer, "setPixelFormat:", mtlPixelFormatRGBA16Float)
		space := appkit.ID(appkit.Call(appkit.Sym("CGColorSpaceCreateWithName"), uintptr(appkit.Constant("kCGColorSpaceExtendedLinearSRGB"))))
		if space == 0 {
			ctx.ready.Store(-1)
			return
		}
		appkit.Send(ctx.layer, "setColorspace:", uintptr(space))
		appkit.Call(appkit.Sym("CGColorSpaceRelease"), uintptr(space))
		appkit.Send(ctx.layer, "setOpaque:", 0)
		appkit.Send(ctx.layer, "setPresentsWithTransaction:", 0)
		appkit.MsgSetFloat(ctx.layer, appkit.Sel("setContentsScale:"), appkit.MsgFloat(window, appkit.Sel("backingScaleFactor")))
		appkit.Send(appkit.Send(view, "layer"), "setOpaque:", 0)
		appkit.Send(window, "setOpaque:", 0)
		appkit.Send(window, "setBackgroundColor:", uintptr(appkit.Send(appkit.Class("NSColor"), "clearColor")))

		bounds := appkit.MsgRect(view, appkit.Sel("bounds"))
		initWithFrame := appkit.Sel("initWithFrame:")
		const resize = nsViewWidthSizable | nsViewHeightSizable
		// Behind-window sampling must live at the native window root, outside
		// the Metal surface that draws the transparent application content.
		frosted, liquid := glassClasses()
		backdrop := appkit.MsgInitRect(appkit.Send(frosted, "alloc"), initWithFrame, bounds)
		appkit.Send(backdrop, "setMaterial:", nsMaterialSidebar)
		appkit.Send(backdrop, "setBlendingMode:", nsBlendingBehindWindow)
		// The material is a persistent compositing surface, not inactive window chrome.
		appkit.Send(backdrop, "setState:", nsStateActive)
		appkit.Send(backdrop, "setWantsLayer:", 1)
		if radius > 0 {
			layer := appkit.Send(backdrop, "layer")
			appkit.MsgSetFloat(layer, appkit.Sel("setCornerRadius:"), radius)
			appkit.Send(layer, "setMasksToBounds:", 1)
		}
		ctx.backdrop = backdrop
		ctx.installing.Store(true)
		defer ctx.installing.Store(false)
		// NSWindow retains the root; our retained view survives the temporary
		// detach before it is installed as the effect's transparent content.
		appkit.Send(window, "setContentView:", uintptr(backdrop))
		appkit.MsgSetRect(view, appkit.Sel("setFrame:"), bounds)
		appkit.Send(view, "setAutoresizingMask:", resize)
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
			appkit.Send(glass, "setContentView:", uintptr(view))
			appkit.Send(glass, "setAutoresizingMask:", resize)
			appkit.Send(backdrop, "addSubview:", uintptr(glass))
			appkit.Release(glass)
		} else {
			appkit.Send(backdrop, "addSubview:", uintptr(view))
		}
		appkit.Send(window, "makeFirstResponder:", uintptr(view))
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
		r.gpu, err = gpu.New(gpu.Metal{Device: uintptr(r.context.device), Queue: uintptr(r.context.queue), PixelFormat: mtlPixelFormatRGBA16Float})
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
