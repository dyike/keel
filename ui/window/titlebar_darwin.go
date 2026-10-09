//go:build darwin && !ios

package window

import (
	"sync"

	gioapp "gioui.org/app"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/ebitengine/purego/objc"
)

var nativeWindows sync.Map // uintptr NSView -> *Window; never pass Go pointers to native code
func platformWindowEvent(w *Window, event any) {
	if e, ok := event.(gioapp.AppKitViewEvent); ok {
		loop.Lock()
		defer loop.Unlock()
		clearPlatformWindow(w)
		w.nativeView = e.View
		w.nativeTrafficLightsDirty = true
		if e.View != 0 {
			nativeWindows.Store(e.View, w)
		}
	}
}
func clearPlatformWindow(w *Window) {
	if w.nativeView != 0 {
		nativeWindows.Delete(w.nativeView)
		setTitleBarArea(w.nativeView, appkit.Rect{})
		w.nativeView, w.lastTitleView = 0, 0
	}
}
func syncTitleBar(w *Window) {
	if w.nativeView != 0 && w.opts.Frameless && w.opts.NativeTrafficLights && w.nativeTrafficLightsDirty {
		w.nativeTrafficLightsDirty = false
		var layout TrafficLightLayout
		custom := w.opts.TrafficLightLayout != nil
		if custom {
			layout = *w.opts.TrafficLightLayout
		}
		nativeTrafficLights(w.nativeView, custom, layout)
	}
	if w.nativeView == 0 || w.nativeView == w.lastTitleView && w.titleArea == w.lastTitleArea {
		return
	}
	w.lastTitleView, w.lastTitleArea = w.nativeView, w.titleArea
	a := w.titleArea
	setTitleBarArea(w.nativeView, appkit.Rect{Origin: appkit.Point{X: float64(a[0]), Y: float64(a[1])}, Size: appkit.Size{Width: float64(a[2]), Height: float64(a[3])}})
}

func postTitleBarAction(view uintptr, action int) {
	if value, ok := nativeWindows.Load(view); ok {
		w := value.(*Window)
		// AppKit invokes this on the main thread. Gio's Invalidate wakes and
		// flushes events inline there, re-entering its own invalidation mutex.
		// Post from another goroutine so the native callback returns first.
		go core.Update(func() {
			if w.closed {
				return
			}
			if action == 1 {
				w.Minimize()
			} else if action == 2 {
				w.ToggleMaximize()
			}
		})
	}
}

const nsEventMaskLeftMouseDown = 1 << 1

// Title bar regions of frameless windows, by view (retained). Main thread only.
var titleBars struct {
	regions map[appkit.ID]appkit.Rect
	monitor appkit.ID
}

// setTitleBarArea makes double-clicks in area (view coordinates, top-left
// origin) do what the system's "double-click a window's title bar" setting
// asks; an empty area removes it. Gio keeps the view valid until the next
// view event: retain it before the asynchronous dispatch so frame code never
// waits for the main thread.
func setTitleBarArea(handle uintptr, area appkit.Rect) {
	view := appkit.Retain(appkit.ID(handle))
	appkit.MainAsync(func() {
		defer appkit.Release(view)
		if titleBars.regions == nil {
			titleBars.regions = map[appkit.ID]appkit.Rect{}
			handler := appkit.NewBlock(func(_ objc.Block, event appkit.ID) appkit.ID { return titleBarMouseDown(event) })
			defer handler.Release()
			titleBars.monitor = appkit.Retain(appkit.Send(appkit.Class("NSEvent"), "addLocalMonitorForEventsMatchingMask:handler:", nsEventMaskLeftMouseDown, uintptr(handler)))
		}
		_, had := titleBars.regions[view]
		if area.Size.Width > 0 && area.Size.Height > 0 {
			if !had {
				appkit.Retain(view)
			}
			titleBars.regions[view] = area
		} else if had {
			delete(titleBars.regions, view)
			appkit.Release(view)
		}
	})
}

// titleBarMouseDown runs on the main thread for every left mouse down.
func titleBarMouseDown(event appkit.ID) appkit.ID {
	if len(titleBars.regions) == 0 || int(appkit.Send(event, "clickCount")) != 2 {
		return event
	}
	window := appkit.Send(event, "window")
	for view, area := range titleBars.regions {
		if w := appkit.Send(view, "window"); w == 0 || w != window {
			continue
		}
		p := appkit.MsgPointFromView(view, appkit.Sel("convertPoint:fromView:"), appkit.MsgPoint(event, appkit.Sel("locationInWindow")), 0)
		if !appkit.SendBool(view, "isFlipped") {
			p.Y = appkit.MsgRect(view, appkit.Sel("bounds")).Size.Height - p.Y
		}
		if !area.Contains(p) {
			continue
		}
		var command int
		appkit.Pool(func() {
			defaults := appkit.Send(appkit.Class("NSUserDefaults"), "standardUserDefaults")
			action := appkit.Send(defaults, "stringForKey:", uintptr(appkit.String("AppleActionOnDoubleClick")))
			switch {
			case appkit.Equal(action, "None"):
			case appkit.Equal(action, "Minimize"):
				command = 1
			default:
				command = 2
			}
		})
		if command != 0 {
			postTitleBarAction(uintptr(view), command)
		}
		return 0 // Do not let Gio start another native window drag.
	}
	return event
}

// trafficLights is the geometry applied to one view's window buttons. Main
// thread only, keyed by the KeelTrafficLightLayout object observing the window.
type trafficLights struct {
	view                                                    appkit.ID
	height, left, offsetY, spacing, systemSpacing, minimumH float64
}

var trafficLayouts = map[appkit.ID]*trafficLights{}

var trafficLayoutClass = sync.OnceValue(func() appkit.ID {
	return appkit.RegisterClass("KeelTrafficLightLayout", "NSObject", nil, []objc.MethodDef{
		appkit.Method("updated:", func(self appkit.ID, _ objc.SEL, _ appkit.ID) {
			if l := trafficLayouts[self]; l != nil {
				l.apply()
			}
		}),
		appkit.Method("dealloc", func(self appkit.ID, _ objc.SEL) {
			appkit.Send(appkit.Send(appkit.Class("NSNotificationCenter"), "defaultCenter"), "removeObserver:", uintptr(self))
			delete(trafficLayouts, self)
			self.SendSuper(appkit.Sel("dealloc"))
		}),
	})
})

const (
	nsWindowCloseButton            = 0
	nsWindowMiniaturizeButton      = 1
	nsWindowZoomButton             = 2
	nsWindowTitleHidden            = 1
	nsTitlebarSeparatorStyleNone   = 1
	objcAssociationRetainNonatomic = 1
)

// AppKit lays out close/minimize again when it updates the window. Apply the
// requested geometry after that pass, and after resize/fullscreen transitions.
func (l *trafficLights) apply() {
	window := appkit.Send(l.view, "window")
	if window == 0 || uintptr(appkit.Send(window, "styleMask"))&nsWindowStyleMaskFullScreen != 0 {
		return
	}
	frameSel, setFrameSel := appkit.Sel("frame"), appkit.Sel("setFrame:")
	setFrame := func(obj appkit.ID, r appkit.Rect) {
		if appkit.MsgRect(obj, frameSel) != r {
			appkit.MsgSetRect(obj, setFrameSel, r)
		}
	}
	if l.height > 0 {
		bar := appkit.Send(appkit.Send(window, "standardWindowButton:", nsWindowCloseButton), "superview")
		container := appkit.Send(bar, "superview")
		barHeight := max(l.height, l.minimumH)
		frame := appkit.MsgRect(container, frameSel)
		frame.Origin.Y = frame.MaxY() - barHeight
		frame.Size.Height = barHeight
		setFrame(container, frame)
		frame = appkit.MsgRect(bar, frameSel)
		frame.Origin.Y = 0
		frame.Size.Height = barHeight
		setFrame(bar, frame)
	}
	flipped := appkit.SendBool(l.view, "isFlipped")
	for kind := uintptr(nsWindowCloseButton); kind <= nsWindowZoomButton; kind++ {
		button := appkit.Send(window, "standardWindowButton:", kind)
		appkit.Send(button, "setHidden:", 0)
		if l.height <= 0 {
			continue
		}
		gap := l.spacing
		if gap <= 0 {
			gap = l.systemSpacing
		}
		frame := appkit.MsgRect(button, frameSel)
		center := appkit.Point{X: l.left + frame.Size.Width/2 + float64(kind)*gap, Y: l.height/2 + l.offsetY}
		if !flipped {
			center.Y = appkit.MsgRect(l.view, appkit.Sel("bounds")).Size.Height - center.Y
		}
		center = appkit.MsgPointFromView(appkit.Send(button, "superview"), appkit.Sel("convertPoint:fromView:"), center, l.view)
		frame.Origin = appkit.Point{X: center.X - frame.Size.Width/2, Y: center.Y - frame.Size.Height/2}
		setFrame(button, frame)
	}
}

// nativeTrafficLights reuses NSWindow's own buttons in their existing
// hierarchy: no replacement cells, tracking areas, actions or accessibility
// implementation.
func nativeTrafficLights(handle uintptr, custom bool, layout TrafficLightLayout) {
	view := appkit.Retain(appkit.ID(handle))
	appkit.MainAsync(func() {
		defer appkit.Release(view)
		window := appkit.Send(view, "window")
		if window == 0 {
			return
		}
		appkit.Send(window, "setTitleVisibility:", nsWindowTitleHidden)
		appkit.Send(window, "setTitlebarAppearsTransparent:", 1)
		if appkit.SendBool(window, "respondsToSelector:", uintptr(appkit.Sel("setTitlebarSeparatorStyle:"))) {
			appkit.Send(window, "setTitlebarSeparatorStyle:", nsTitlebarSeparatorStyleNone)
		}
		// The layout object belongs to the view (an associated object), so
		// it goes away with it; the selector's address is a stable key.
		key := uintptr(appkit.Sel("keelTrafficLightLayout"))
		obj := appkit.ID(appkit.Call(appkit.Sym("objc_getAssociatedObject"), uintptr(view), key))
		l := trafficLayouts[obj]
		if obj == 0 || l == nil {
			obj = appkit.Send(appkit.Send(trafficLayoutClass(), "alloc"), "init")
			close := appkit.Send(window, "standardWindowButton:", nsWindowCloseButton)
			mini := appkit.Send(window, "standardWindowButton:", nsWindowMiniaturizeButton)
			l = &trafficLights{
				view:          view, // not retained: the view owns the layout
				systemSpacing: appkit.MsgRect(mini, appkit.Sel("frame")).Origin.X - appkit.MsgRect(close, appkit.Sel("frame")).Origin.X,
				minimumH:      appkit.MsgRect(appkit.Send(close, "superview"), appkit.Sel("frame")).Size.Height,
			}
			trafficLayouts[obj] = l
			appkit.Call(appkit.Sym("objc_setAssociatedObject"), uintptr(view), key, uintptr(obj), objcAssociationRetainNonatomic)
			center := appkit.Send(appkit.Class("NSNotificationCenter"), "defaultCenter")
			for _, name := range []string{"NSWindowDidUpdateNotification", "NSWindowDidResizeNotification", "NSWindowDidExitFullScreenNotification"} {
				appkit.Send(center, "addObserver:selector:name:object:", uintptr(obj), uintptr(appkit.Sel("updated:")), uintptr(appkit.Constant(name)), uintptr(window))
			}
			appkit.Release(obj)
		}
		l.height = 0
		if custom {
			l.height = float64(layout.Height)
		}
		l.left, l.offsetY, l.spacing = float64(layout.Left), float64(layout.OffsetY), float64(layout.Spacing)
		l.apply()
	})
}
