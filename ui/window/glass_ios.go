//go:build ios && !nometal

package window

import (
	gioapp "github.com/dyike/keel/gio/app"
	"github.com/ebitengine/purego/objc"
)

const (
	uiViewAutoresizingFlexibleWidth  = 1 << 1
	uiViewAutoresizingFlexibleHeight = 1 << 4
	uiBlurEffectStyleSystemMaterial  = 8
	uiGlassEffectStyleRegular        = 0
	uiGlassEffectStyleClear          = 1
)

func platformGlassSupported() bool {
	return objc.GetClass("UIVisualEffectView") != 0
}

// Liquid Glass on iOS 26+ uses UIGlassEffect inside UIVisualEffectView.
func platformLiquidGlassSupported() bool {
	return objc.GetClass("UIGlassEffect") != 0
}

func platformGlassEvent(w *Window, event any) {
	e, ok := event.(gioapp.UIKitViewEvent)
	if !ok || w.opts.Glass == nil || e.ViewController == 0 || !platformGlassSupported() {
		return
	}
	style := w.opts.Glass.Style
	radius := float64(w.opts.Glass.CornerRadius)
	vc := objc.ID(e.ViewController)
	iosMainAsync(func() {
		pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(sel("new"))
		defer pool.Send(sel("drain"))
		view := vc.Send(sel("view"))
		if view == 0 {
			return
		}
		clearColor := objc.ID(objc.GetClass("UIColor")).Send(sel("clearColor"))
		view.Send(sel("setOpaque:"), false)
		view.Send(sel("setBackgroundColor:"), clearColor)
		if layer := view.Send(sel("layer")); layer != 0 {
			layer.Send(sel("setOpaque:"), false)
		}

		var effect objc.ID
		if glassCls := objc.GetClass("UIGlassEffect"); glassCls != 0 && style != GlassFrosted {
			s := uintptr(uiGlassEffectStyleRegular)
			if style == GlassClear {
				s = uiGlassEffectStyleClear
			}
			effect = objc.ID(glassCls).Send(sel("effectWithStyle:"), s)
		} else if blurCls := objc.GetClass("UIBlurEffect"); blurCls != 0 {
			effect = objc.ID(blurCls).Send(sel("effectWithStyle:"), uintptr(uiBlurEffectStyleSystemMaterial))
		}
		if effect == 0 {
			return
		}
		effectView := objc.ID(objc.GetClass("UIVisualEffectView")).Send(sel("alloc")).Send(sel("initWithEffect:"), effect)
		if effectView == 0 {
			return
		}
		effectView.Send(sel("setAutoresizingMask:"), uintptr(uiViewAutoresizingFlexibleWidth|uiViewAutoresizingFlexibleHeight))
		effectView.Send(sel("setUserInteractionEnabled:"), false)
		if radius > 0 {
			if layer := effectView.Send(sel("layer")); layer != 0 {
				layer.Send(sel("setCornerRadius:"), radius)
				layer.Send(sel("setMasksToBounds:"), true)
			}
		}
		view.Send(sel("insertSubview:atIndex:"), effectView, uintptr(0))
		effectView.Send(sel("release"))
	})
}
