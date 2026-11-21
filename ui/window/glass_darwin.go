//go:build darwin && !ios && !nometal

package window

import (
	"errors"
	"image"
	"image/color"
	"sync"
	"sync/atomic"
	"time"

	gioapp "github.com/dyike/keel/gio/app"
	"github.com/dyike/keel/gio/gpu"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/appkit"
)

const mtlPixelFormatRGBA16Float = 115

func platformGlassSupported() bool { return appkit.Class("CAMetalLayer") != 0 }

// Liquid Glass needs macOS 26, where NSGlassEffectView exists.
func platformLiquidGlassSupported() bool { return appkit.Class("NSGlassEffectView") != 0 }

// The native effect is the window root, as in a standard AppKit glass
// window. Its ordinary hit testing routes input to the full-size GioView.
var glassClasses = sync.OnceValue(func() struct{ frosted, liquid, container appkit.ID } {
	return struct{ frosted, liquid, container appkit.ID }{
		frosted:   appkit.Class("NSVisualEffectView"),
		liquid:    appkit.Class("NSGlassEffectView"),
		container: appkit.Class("NSGlassEffectContainerView"),
	}
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
			run:        w.win.Run,
		}
	}
}

// glassContext owns the native objects. ready is 0 while AppKit installs
// them, 1 once installed, -1 on failure; the objects are written before
// ready is set and only released on the main queue after Release.
type glassContext struct {
	ready                                atomic.Int32
	installing                           atomic.Bool
	released                             atomic.Bool
	view, backdrop, layer, device, queue appkit.ID
	windowGlass                          appkit.ID   // retained NSGlassEffectView for liquid glass windows
	controlHost                          appkit.ID   // retained NSView hosting glassContainer (below) + GioView (above)
	glassContainer                       appkit.ID   // retained NSGlassEffectContainerView (or NSView fallback)
	pillsHost                            appkit.ID   // retained NSView set as glassContainer.contentView
	pills                                []appkit.ID // retained NSGlassEffectView pool
	checkedInteractive                   bool
	supportsInteractive                  bool
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
		classes := glassClasses()
		frosted, liquid := classes.frosted, classes.liquid
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
			ctx.windowGlass = glass
		} else {
			appkit.Send(backdrop, "addSubview:", uintptr(view))
		}
		appkit.Send(window, "makeFirstResponder:", uintptr(view))
		ctx.ready.Store(1)
	})
	return ctx
}

type metalGlassRenderer struct {
	// Idle reclamation runs off the event loop. Serialize it with Frame and
	// Release. Frame can wait for Window.Run outside the frame lock; AppKit
	// must never wait for mu (idle trimming uses TryLock).
	mu         sync.Mutex
	idleTimer  *time.Timer
	idleAt     time.Time
	generation uint64
	context    *glassContext
	gpu        gpu.GPU
	invalidate func()
	run        func(func())
	lastSize   image.Point // last successfully presented size, not idle drawableSize
}

func (r *metalGlassRenderer) Frame(ops *op.Ops, size image.Point) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generation++
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
	defer r.scheduleIdle()
	var err error
	appkit.Pool(func() {
		layer := r.context.layer
		resized := size != r.lastSize
		appkit.Send(layer, "setPresentsWithTransaction:", appkit.Bool(resized))
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
			if resized {
				// Gio services Run while displayLayer is waiting for this frame.
				// Present in that main-thread transaction so the new bounds and
				// pixels become visible together. GPU scheduling stays off main.
				r.run(func() { appkit.Send(drawable, "present") })
			} else {
				appkit.Send(drawable, "present")
			}
			r.lastSize = size
		}
	})
	return err
}

// A static glass window retains its presented frame in Core Animation.
// Reclaim the renderer's rebuildable caches and spare full-size drawables,
// just as Gio's ordinary Metal context does. Keep the compiled pipelines
// and triple buffering; the next Frame restores the full drawable size.
func (r *metalGlassRenderer) scheduleIdle() {
	r.idleAt = time.Now().Add(time.Second)
	// Keep one timer across a run of frames; do not allocate one per frame.
	if r.idleTimer == nil {
		r.idleTimer = time.AfterFunc(time.Second, r.trimIdle)
	}
}

func (r *metalGlassRenderer) stopIdle() {
	r.generation++
	if r.idleTimer != nil {
		r.idleTimer.Stop()
		r.idleTimer = nil
	}
}

func (r *metalGlassRenderer) trimIdle() {
	r.mu.Lock()
	if r.context == nil || r.gpu == nil {
		r.mu.Unlock()
		return
	}
	if delay := time.Until(r.idleAt); delay > 0 {
		r.idleTimer.Reset(delay)
		r.mu.Unlock()
		return
	}
	generation := r.generation
	r.idleTimer = nil
	appkit.Pool(func() {
		if g, ok := r.gpu.(interface{ Trim() }); ok {
			g.Trim()
		}
	})
	r.mu.Unlock()
	appkit.MainAsync(func() {
		// Never block AppKit on a frame (nextDrawable may itself be waiting for
		// the compositor). A new frame cancels this generation in any case.
		if !r.mu.TryLock() {
			return
		}
		defer r.mu.Unlock()
		if generation != r.generation || r.context == nil {
			return
		}
		tx := appkit.Class("CATransaction")
		appkit.Send(tx, "begin")
		appkit.Send(tx, "setDisableActions:", 1)
		appkit.MsgSetSize(r.context.layer, appkit.Sel("setDrawableSize:"), appkit.Size{Width: 1, Height: 1})
		appkit.Send(tx, "commit")
	})
}

func (r *metalGlassRenderer) SyncRegions(regions []core.GlassRegion, windowHeight float32) {
	r.mu.Lock()
	ctx := r.context
	r.mu.Unlock()
	if ctx == nil || ctx.released.Load() {
		return
	}
	appkit.MainAsync(func() {
		ctx.applyRegions(regions, windowHeight)
	})
}

func (ctx *glassContext) applyRegions(regions []core.GlassRegion, windowHeight float32) {
	if ctx.released.Load() || ctx.ready.Load() != 1 || ctx.windowGlass == 0 {
		return
	}
	classes := glassClasses()
	liquid, containerCls := classes.liquid, classes.container
	if liquid == 0 {
		return
	}
	if len(regions) == 0 && ctx.controlHost == 0 {
		return
	}
	tx := appkit.Class("CATransaction")
	appkit.Send(tx, "begin")
	appkit.Send(tx, "setDisableActions:", 1)
	defer appkit.Send(tx, "commit")

	initWithFrame := appkit.Sel("initWithFrame:")
	const resize = nsViewWidthSizable | nsViewHeightSizable
	if ctx.controlHost == 0 {
		bounds := appkit.MsgRect(ctx.windowGlass, appkit.Sel("bounds"))
		nsView := appkit.Class("NSView")

		controlHost := appkit.MsgInitRect(appkit.Send(nsView, "alloc"), initWithFrame, bounds)
		appkit.Send(controlHost, "setAutoresizingMask:", resize)

		var glassContainer appkit.ID
		if containerCls != 0 {
			glassContainer = appkit.MsgInitRect(appkit.Send(containerCls, "alloc"), initWithFrame, bounds)
		} else {
			glassContainer = appkit.MsgInitRect(appkit.Send(nsView, "alloc"), initWithFrame, bounds)
		}
		appkit.Send(glassContainer, "setAutoresizingMask:", resize)

		pillsHost := appkit.MsgInitRect(appkit.Send(nsView, "alloc"), initWithFrame, bounds)
		appkit.Send(pillsHost, "setAutoresizingMask:", resize)
		if containerCls != 0 {
			appkit.Send(glassContainer, "setContentView:", uintptr(pillsHost))
		} else {
			appkit.Send(glassContainer, "addSubview:", uintptr(pillsHost))
		}
		appkit.Send(controlHost, "addSubview:", uintptr(glassContainer))

		window := appkit.Send(ctx.view, "window")
		ctx.installing.Store(true)
		appkit.Send(ctx.windowGlass, "setContentView:", uintptr(controlHost))
		appkit.MsgSetRect(ctx.view, appkit.Sel("setFrame:"), bounds)
		appkit.Send(ctx.view, "setAutoresizingMask:", resize)
		// GioView stays above glassContainer so it receives all input and draws
		// crisp labels/icons over the native NSGlassEffectView capsules.
		appkit.Send(controlHost, "addSubview:", uintptr(ctx.view))
		if window != 0 {
			appkit.Send(window, "makeFirstResponder:", uintptr(ctx.view))
		}
		ctx.installing.Store(false)

		ctx.controlHost = controlHost
		ctx.glassContainer = glassContainer
		ctx.pillsHost = pillsHost
	}

	bounds := appkit.MsgRect(ctx.controlHost, appkit.Sel("bounds"))
	height := bounds.Size.Height
	if height <= 0 {
		height = float64(windowHeight)
	}

	var maxSpacing float64
	for _, reg := range regions {
		if float64(reg.Spacing) > maxSpacing {
			maxSpacing = float64(reg.Spacing)
		}
	}
	if containerCls != 0 {
		appkit.MsgSetFloat(ctx.glassContainer, appkit.Sel("setSpacing:"), maxSpacing)
	}

	for len(ctx.pills) < len(regions) {
		pill := appkit.MsgInitRect(appkit.Send(liquid, "alloc"), initWithFrame, appkit.Rect{})
		if !ctx.checkedInteractive {
			ctx.checkedInteractive = true
			ctx.supportsInteractive = appkit.SendBool(pill, "respondsToSelector:", uintptr(appkit.Sel("setEffectIsInteractive:")))
		}
		appkit.Send(ctx.pillsHost, "addSubview:", uintptr(pill))
		ctx.pills = append(ctx.pills, pill)
	}

	nsColor := appkit.Class("NSColor")
	colorSel := appkit.Sel("colorWithSRGBRed:green:blue:alpha:")
	setFrameSel := appkit.Sel("setFrame:")
	setRadiusSel := appkit.Sel("setCornerRadius:")
	for i, reg := range regions {
		pill := ctx.pills[i]
		frame := appkit.Rect{
			Origin: appkit.Point{
				X: float64(reg.X),
				Y: height - float64(reg.Y) - float64(reg.Height),
			},
			Size: appkit.Size{
				Width:  float64(reg.Width),
				Height: float64(reg.Height),
			},
		}
		appkit.MsgSetRect(pill, setFrameSel, frame)
		appkit.MsgSetFloat(pill, setRadiusSel, float64(reg.CornerRadius))
		s := uintptr(nsGlassEffectRegular)
		if reg.Style == core.GlassRegionClear {
			s = nsGlassEffectClear
		}
		appkit.Send(pill, "setStyle:", s)
		if reg.Tint.A > 0 {
			c := appkit.MsgColorRGBA(nsColor, colorSel,
				float64(reg.Tint.R)/255.0,
				float64(reg.Tint.G)/255.0,
				float64(reg.Tint.B)/255.0,
				float64(reg.Tint.A)/255.0,
			)
			appkit.Send(pill, "setTintColor:", uintptr(c))
		} else {
			appkit.Send(pill, "setTintColor:", 0)
		}
		if ctx.supportsInteractive {
			appkit.Send(pill, "setEffectIsInteractive:", appkit.Bool(reg.Interactive))
		}
		appkit.Send(pill, "setHidden:", 0)
	}
	for i := len(regions); i < len(ctx.pills); i++ {
		appkit.Send(ctx.pills[i], "setHidden:", 1)
	}
}

func (r *metalGlassRenderer) Release() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stopIdle()
	if r.gpu != nil {
		r.gpu.Release()
		r.gpu = nil
	}
	if ctx := r.context; ctx != nil {
		r.context = nil
		ctx.released.Store(true)
		// Enqueued after installation, including when the window closes
		// before installation finishes. Never wait for AppKit with the frame
		// lock held.
		appkit.MainAsync(func() {
			for _, pill := range ctx.pills {
				appkit.Send(pill, "removeFromSuperview")
				appkit.Release(pill)
			}
			ctx.pills = nil
			if ctx.backdrop != 0 {
				appkit.Send(ctx.backdrop, "removeFromSuperview")
			}
			for _, obj := range []appkit.ID{ctx.pillsHost, ctx.glassContainer, ctx.controlHost, ctx.windowGlass, ctx.backdrop, ctx.layer, ctx.queue, ctx.device, ctx.view} {
				appkit.Release(obj)
			}
		})
	}
}
