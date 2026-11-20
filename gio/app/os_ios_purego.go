// SPDX-License-Identifier: Unlicense OR MIT

//go:build ios

package app

import (
	"runtime"
	"structs"
	"sync"
	"unsafe"

	"github.com/dyike/keel/gio/io/key"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The UIKit binding of os_ios.go without cgo: os_ios.m's classes
// (_gioAppDelegate, GioViewController, GioView) registered at run time with
// the same names (ui/window/scene_ios.go looks them up by name), plus
// CADisplayLink, clipboard, touches and UIApplicationMain. Keel patch.

type uiEdgeInsets struct {
	_                        structs.HostLayout
	Top, Left, Bottom, Right float64
}

type iosDrawParams struct {
	dpi, sdpi                float64
	width, height            float64
	top, right, bottom, left float64
}

var (
	msgF64F64        func(id, sel uintptr, v float64) float64
	msgPointID       func(id, sel, view uintptr) nsPoint
	msgEdgeInsets    func(id, sel uintptr) uiEdgeInsets
	msgSetEdgeInsets func(id, sel uintptr, insets uiEdgeInsets)
)

var iosRT struct {
	applicationMain uintptr
}

var (
	selView                            = sel("view")
	selSetView                         = sel("setView:")
	selSubviews                        = sel("subviews")
	selAddSubview                      = sel("addSubview:")
	selSetLayoutMargins                = sel("setLayoutMargins:")
	selLayoutMargins                   = sel("layoutMargins")
	selSafeAreaInsets                  = sel("safeAreaInsets")
	selSetMultipleTouchEnabled         = sel("setMultipleTouchEnabled:")
	selSetPreservesSuperviewLayoutMarg = sel("setPreservesSuperviewLayoutMargins:")
	selSetNeedsLayout                  = sel("setNeedsLayout")
	selUserInfo                        = sel("userInfo")
	selCGRectValue                     = sel("CGRectValue")
	selContentScaleFactor              = sel("contentScaleFactor")
	selSetContentScaleFactor           = sel("setContentScaleFactor:")
	selNativeScale                     = sel("nativeScale")
	selWindowScene                     = sel("windowScene")
	selInstancesRespondToSel           = sel("instancesRespondToSelector:")
	selRemoveObserver                  = sel("removeObserver:")
	selRemoveObserverNameObject        = sel("removeObserver:name:object:")
	selIsFirstResponder                = sel("isFirstResponder")
	selAllObjects                      = sel("allObjects")
	selCoalescedTouchesForTouch        = sel("coalescedTouchesForTouch:")
	selLocationInView                  = sel("locationInView:")
	selPhase                           = sel("phase")
	selKeyCommandWithInput             = sel("keyCommandWithInput:modifierFlags:action:")
	selInitWithObjectsCount            = sel("initWithObjects:count:")
	selDefaultMetrics                  = sel("defaultMetrics")
	selScaledValueForValue             = sel("scaledValueForValue:")
	selSetString                       = sel("setString:")
	selSetWindow                       = sel("setWindow:")
	selInitWithNibName                 = sel("initWithNibName:bundle:")
	selSetRootViewController           = sel("setRootViewController:")
	selMakeKeyAndVisible               = sel("makeKeyAndVisible")
	selViewDidDisappear                = sel("viewDidDisappear:")
	selViewDidLayoutSubviews           = sel("viewDidLayoutSubviews")
	selDidReceiveMemoryWarning         = sel("didReceiveMemoryWarning")
	selKeyboardWillChange              = sel("keyboardWillChange:")
	selKeyboardWillHide                = sel("keyboardWillHide:")
	selAppDidEnterBackground           = sel("applicationDidEnterBackground:")
	selAppWillEnterForeground          = sel("applicationWillEnterForeground:")
	selOnSceneDidActivate              = sel("onSceneDidActivate:")
	selOnSceneWillDeactivate           = sel("onSceneWillDeactivate:")
	selOnWindowDidBecomeKey            = sel("onWindowDidBecomeKey:")
	selOnWindowDidResignKey            = sel("onWindowDidResignKey:")
	selOnUpArrow                       = sel("onUpArrow")
	selOnDownArrow                     = sel("onDownArrow")
	selOnLeftArrow                     = sel("onLeftArrow")
	selOnRightArrow                    = sel("onRightArrow")
	selOnFrameCallback                 = sel("onFrameCallback:")
	selLayerClass                      = sel("layerClass")
	selDisplayLinkWithTarget           = sel("displayLinkWithTarget:selector:")
	selSetPaused                       = sel("setPaused:")
	selMainRunLoop                     = sel("mainRunLoop")
	selCurrentMode                     = sel("currentMode")
	selAddToRunLoopForMode             = sel("addToRunLoop:forMode:")
	selInvalidate                      = sel("invalidate")
)

var (
	classGioAppDelegate    uintptr
	classGioViewController uintptr
	classGioView           uintptr

	classUIView        uintptr
	classUIWindow      uintptr
	classUIScreen      uintptr
	classUIPasteboard  uintptr
	classUIKeyCommand  uintptr
	classUIFontMetrics uintptr
	classCADisplayLink uintptr
	classNSRunLoop     uintptr

	// UIKit notification and key constants.
	uiKeyboardWillShowNotification        uintptr
	uiKeyboardWillChangeFrameNotification uintptr
	uiKeyboardWillHideNotification        uintptr
	uiKeyboardFrameEndUserInfoKey         uintptr
	uiAppDidEnterBackgroundNotification   uintptr
	uiAppWillEnterForegroundNotification  uintptr
	uiSceneDidActivateNotification        uintptr
	uiSceneWillDeactivateNotification     uintptr
	uiWindowDidBecomeKeyNotification      uintptr
	uiWindowDidResignKeyNotification      uintptr
	uiKeyInputUpArrow                     uintptr
	uiKeyInputDownArrow                   uintptr
	uiKeyInputLeftArrow                   uintptr
	uiKeyInputRightArrow                  uintptr

	iosKeyCommands uintptr
)

var iosViews struct {
	sync.Mutex
	m map[uintptr]*window
}

func setIOSView(view uintptr, w *window) {
	iosViews.Lock()
	if iosViews.m == nil {
		iosViews.m = make(map[uintptr]*window)
	}
	iosViews.m[view] = w
	iosViews.Unlock()
}

func deleteIOSView(view uintptr) {
	if view == 0 {
		return
	}
	iosViews.Lock()
	delete(iosViews.m, view)
	iosViews.Unlock()
}

func windowOf(view uintptr) *window {
	if view == 0 {
		return nil
	}
	iosViews.Lock()
	w := iosViews.m[view]
	iosViews.Unlock()
	return w
}

var iosControllers struct {
	sync.Mutex
	keyboardHeight map[uintptr]float64
}

func controllerKeyboardHeight(ctrl uintptr) float64 {
	iosControllers.Lock()
	h := iosControllers.keyboardHeight[ctrl]
	iosControllers.Unlock()
	return h
}

func setControllerKeyboardHeight(ctrl uintptr, h float64) {
	iosControllers.Lock()
	if iosControllers.keyboardHeight == nil {
		iosControllers.keyboardHeight = make(map[uintptr]float64)
	}
	iosControllers.keyboardHeight[ctrl] = h
	iosControllers.Unlock()
}

func deleteControllerState(ctrl uintptr) {
	iosControllers.Lock()
	delete(iosControllers.keyboardHeight, ctrl)
	iosControllers.Unlock()
}

var appDelegateWindows struct {
	sync.Mutex
	m map[uintptr]uintptr
}

func getAppDelegateWindow(del uintptr) uintptr {
	appDelegateWindows.Lock()
	w := appDelegateWindows.m[del]
	appDelegateWindows.Unlock()
	return w
}

func setAppDelegateWindow(del, win uintptr) {
	if win != 0 {
		retainObject(win)
	}
	appDelegateWindows.Lock()
	if appDelegateWindows.m == nil {
		appDelegateWindows.m = make(map[uintptr]uintptr)
	}
	old := appDelegateWindows.m[del]
	if win == 0 {
		delete(appDelegateWindows.m, del)
	} else {
		appDelegateWindows.m[del] = win
	}
	appDelegateWindows.Unlock()
	releaseObject(old)
}

func loadNSStringConst(lib uintptr, name string) uintptr {
	addr, err := purego.Dlsym(lib, name)
	if err != nil || addr == 0 {
		return 0
	}
	return **(**uintptr)(unsafe.Pointer(&addr))
}

func loadPlatformRuntime() error {
	uikit, err := purego.Dlopen("/System/Library/Frameworks/UIKit.framework/UIKit", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	iosRT.applicationMain, err = purego.Dlsym(uikit, "UIApplicationMain")
	if err != nil {
		return err
	}

	purego.RegisterFunc(&msgF64F64, rt.msg)
	purego.RegisterFunc(&msgPointID, rt.msg)
	purego.RegisterFunc(&msgSetEdgeInsets, rt.msg)
	purego.RegisterFunc(&msgEdgeInsets, rt.msgStret)

	for _, c := range []struct {
		addr *uintptr
		name string
	}{
		{&classUIView, "UIView"},
		{&classUIWindow, "UIWindow"},
		{&classUIScreen, "UIScreen"},
		{&classUIPasteboard, "UIPasteboard"},
		{&classUIKeyCommand, "UIKeyCommand"},
		{&classCADisplayLink, "CADisplayLink"},
		{&classNSRunLoop, "NSRunLoop"},
	} {
		*c.addr = class(c.name)
	}
	classUIFontMetrics = uintptr(objc.GetClass("UIFontMetrics"))

	for _, c := range []struct {
		addr *uintptr
		name string
	}{
		{&uiKeyboardWillShowNotification, "UIKeyboardWillShowNotification"},
		{&uiKeyboardWillChangeFrameNotification, "UIKeyboardWillChangeFrameNotification"},
		{&uiKeyboardWillHideNotification, "UIKeyboardWillHideNotification"},
		{&uiKeyboardFrameEndUserInfoKey, "UIKeyboardFrameEndUserInfoKey"},
		{&uiAppDidEnterBackgroundNotification, "UIApplicationDidEnterBackgroundNotification"},
		{&uiAppWillEnterForegroundNotification, "UIApplicationWillEnterForegroundNotification"},
		{&uiSceneDidActivateNotification, "UISceneDidActivateNotification"},
		{&uiSceneWillDeactivateNotification, "UISceneWillDeactivateNotification"},
		{&uiWindowDidBecomeKeyNotification, "UIWindowDidBecomeKeyNotification"},
		{&uiWindowDidResignKeyNotification, "UIWindowDidResignKeyNotification"},
		{&uiKeyInputUpArrow, "UIKeyInputUpArrow"},
		{&uiKeyInputDownArrow, "UIKeyInputDownArrow"},
		{&uiKeyInputLeftArrow, "UIKeyInputLeftArrow"},
		{&uiKeyInputRightArrow, "UIKeyInputRightArrow"},
	} {
		*c.addr = loadNSStringConst(uikit, c.name)
	}

	registerIOSClasses()
	return nil
}

func registerIOSClasses() {
	var err error
	var appDelProtocols []*objc.Protocol
	if p := objc.GetProtocol("UIApplicationDelegate"); p != nil {
		appDelProtocols = append(appDelProtocols, p)
	}
	classGioAppDelegate, err = registerClass("_gioAppDelegate", "UIResponder", appDelProtocols, iosAppDelegateMethods())
	if err == nil {
		classGioViewController, err = registerClass("GioViewController", "UIViewController", nil, iosViewControllerMethods())
	}
	if err == nil {
		var viewProtocols []*objc.Protocol
		if p := objc.GetProtocol("UIKeyInput"); p != nil {
			viewProtocols = append(viewProtocols, p)
		}
		classGioView, err = registerClass("GioView", "UIView", viewProtocols, iosViewMethods())
	}
	if err != nil {
		panic(err)
	}
	addClassMethod(classGioView, selOnFrameCallback, func(_ objc.ID, _ objc.SEL, link uintptr) {
		onFrameCallback(link)
	}, "v@:@")
	addClassMethod(classGioView, selLayerClass, func(_ objc.ID, _ objc.SEL) uintptr {
		return gioLayerClass()
	}, "#:@:")
}

func iosAppDelegateMethods() []objc.MethodDef {
	return []objc.MethodDef{
		{Cmd: objc.SEL(selWindow), Fn: func(self objc.ID, _ objc.SEL) uintptr {
			return getAppDelegateWindow(uintptr(self))
		}},
		{Cmd: objc.SEL(selSetWindow), Fn: func(self objc.ID, _ objc.SEL, win uintptr) {
			setAppDelegateWindow(uintptr(self), win)
		}},
		{Cmd: objc.RegisterName("application:didFinishLaunchingWithOptions:"), Fn: func(self objc.ID, _ objc.SEL, _, _ uintptr) bool {
			bounds := msgRect(send(classUIScreen, selMainScreen), selBounds)
			win := msgInitRect(send(classUIWindow, selAlloc), selInitWithFrame, bounds)
			send(uintptr(self), selSetWindow, win)
			releaseObject(win)
			controller := send(send(classGioViewController, selAlloc), selInitWithNibName, 0, 0)
			send(win, selSetRootViewController, controller)
			releaseObject(controller)
			send(win, selMakeKeyAndVisible)
			return true
		}},
		{Cmd: objc.RegisterName("application:openURL:options:"), Fn: func(_ objc.ID, _ objc.SEL, _, url, _ uintptr) bool {
			if url != 0 {
				onOpenURI(nsstringToString(send(url, selAbsoluteString)))
			}
			return true
		}},
		{Cmd: objc.SEL(selDealloc), Fn: func(self objc.ID, _ objc.SEL) {
			setAppDelegateWindow(uintptr(self), 0)
			sendSuper(uintptr(self), classGioAppDelegate, selDealloc)
		}},
	}
}

func controllerDrawView(ctrl uintptr) uintptr {
	root := send(ctrl, selView)
	if root == 0 {
		return 0
	}
	subs := send(root, selSubviews)
	if subs == 0 || send(subs, selCount) == 0 {
		return 0
	}
	return send(subs, selObjectAtIndex, 0)
}

func iosViewControllerMethods() []objc.MethodDef {
	return []objc.MethodDef{
		{Cmd: objc.RegisterName("loadView"), Fn: func(self objc.ID, _ objc.SEL) {
			gioRunMain()

			zeroFrame := nsRect{}
			rootView := msgInitRect(send(classUIView, selAlloc), selInitWithFrame, zeroFrame)
			send(uintptr(self), selSetView, rootView)
			defer releaseObject(rootView)
			msgSetEdgeInsets(rootView, selSetLayoutMargins, uiEdgeInsets{})

			drawView := msgInitRect(send(classGioView, selAlloc), selInitWithFrame, zeroFrame)
			send(rootView, selAddSubview, drawView)
			defer releaseObject(drawView)
			send(drawView, selSetMultipleTouchEnabled, 1)
			send(drawView, selSetPreservesSuperviewLayoutMarg, 1)
			msgSetEdgeInsets(drawView, selSetLayoutMargins, uiEdgeInsets{})
			onCreate(drawView, uintptr(self))

			center := send(classNSNotificationCenter, selDefaultCenter)
			if uiKeyboardWillShowNotification != 0 {
				send(center, selAddObserver, uintptr(self), selKeyboardWillChange, uiKeyboardWillShowNotification, 0)
			}
			if uiKeyboardWillChangeFrameNotification != 0 {
				send(center, selAddObserver, uintptr(self), selKeyboardWillChange, uiKeyboardWillChangeFrameNotification, 0)
			}
			if uiKeyboardWillHideNotification != 0 {
				send(center, selAddObserver, uintptr(self), selKeyboardWillHide, uiKeyboardWillHideNotification, 0)
			}
			if uiAppDidEnterBackgroundNotification != 0 {
				send(center, selAddObserver, uintptr(self), selAppDidEnterBackground, uiAppDidEnterBackgroundNotification, 0)
			}
			if uiAppWillEnterForegroundNotification != 0 {
				send(center, selAddObserver, uintptr(self), selAppWillEnterForeground, uiAppWillEnterForegroundNotification, 0)
			}
		}},
		{Cmd: objc.SEL(selAppWillEnterForeground), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			if w := windowOf(controllerDrawView(uintptr(self))); w != nil {
				w.onStart()
			}
		}},
		{Cmd: objc.SEL(selAppDidEnterBackground), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			if w := windowOf(controllerDrawView(uintptr(self))); w != nil {
				w.onStop()
			}
		}},
		{Cmd: objc.SEL(selViewDidDisappear), Fn: func(self objc.ID, _ objc.SEL, animated bool) {
			sendSuper(uintptr(self), classGioViewController, selViewDidDisappear, boolArg(animated))
			if w := windowOf(controllerDrawView(uintptr(self))); w != nil {
				w.onDestroy()
			}
		}},
		{Cmd: objc.SEL(selViewDidLayoutSubviews), Fn: func(self objc.ID, _ objc.SEL) {
			sendSuper(uintptr(self), classGioViewController, selViewDidLayoutSubviews)
			view := controllerDrawView(uintptr(self))
			if view == 0 {
				return
			}
			frame := msgRect(send(uintptr(self), selView), selBounds)
			// Adjust view bounds to make room for the keyboard.
			frame.Size.Y -= controllerKeyboardHeight(uintptr(self))
			msgSetRect(view, selSetFrame, frame)
			if w := windowOf(view); w != nil {
				w.draw(true)
			}
		}},
		{Cmd: objc.SEL(selDidReceiveMemoryWarning), Fn: func(self objc.ID, _ objc.SEL) {
			onLowMemory()
			sendSuper(uintptr(self), classGioViewController, selDidReceiveMemoryWarning)
		}},
		{Cmd: objc.SEL(selKeyboardWillChange), Fn: func(self objc.ID, _ objc.SEL, note uintptr) {
			userInfo := send(note, selUserInfo)
			if userInfo == 0 || uiKeyboardFrameEndUserInfoKey == 0 {
				return
			}
			val := send(userInfo, selObjectForKey, uiKeyboardFrameEndUserInfoKey)
			if val == 0 {
				return
			}
			f := msgRect(val, selCGRectValue)
			setControllerKeyboardHeight(uintptr(self), f.Size.Y)
			send(send(uintptr(self), selView), selSetNeedsLayout)
		}},
		{Cmd: objc.SEL(selKeyboardWillHide), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			setControllerKeyboardHeight(uintptr(self), 0)
			send(send(uintptr(self), selView), selSetNeedsLayout)
		}},
		{Cmd: objc.SEL(selDealloc), Fn: func(self objc.ID, _ objc.SEL) {
			send(send(classNSNotificationCenter, selDefaultCenter), selRemoveObserver, uintptr(self))
			deleteControllerState(uintptr(self))
			sendSuper(uintptr(self), classGioViewController, selDealloc)
		}},
	}
}

func registerSceneNotifications(view, newWindow uintptr) {
	center := send(classNSNotificationCenter, selDefaultCenter)
	if oldWin := send(view, selWindow); oldWin != 0 {
		oldScene := send(oldWin, selWindowScene)
		send(center, selRemoveObserverNameObject, view, uiSceneDidActivateNotification, oldScene)
		send(center, selRemoveObserverNameObject, view, uiSceneWillDeactivateNotification, oldScene)
	}
	var newScene uintptr
	if newWindow != 0 {
		newScene = send(newWindow, selWindowScene)
	}
	send(center, selAddObserver, view, selOnSceneDidActivate, uiSceneDidActivateNotification, newScene)
	send(center, selAddObserver, view, selOnSceneWillDeactivate, uiSceneWillDeactivateNotification, newScene)
}

func registerWindowNotifications(view, newWindow uintptr) {
	center := send(classNSNotificationCenter, selDefaultCenter)
	if oldWin := send(view, selWindow); oldWin != 0 {
		send(center, selRemoveObserverNameObject, view, uiWindowDidBecomeKeyNotification, oldWin)
		send(center, selRemoveObserverNameObject, view, uiWindowDidResignKeyNotification, oldWin)
	}
	send(center, selAddObserver, view, selOnWindowDidBecomeKey, uiWindowDidBecomeKeyNotification, newWindow)
	send(center, selAddObserver, view, selOnWindowDidResignKey, uiWindowDidResignKeyNotification, newWindow)
}

func handleTouches(last bool, view, touches, event uintptr) {
	w := windowOf(view)
	if w == nil {
		return
	}
	scale := msgF64(view, selContentScaleFactor)
	allTouches := send(touches, selAllObjects)
	n := send(allTouches, selCount)
	for i := uintptr(0); i < n; i++ {
		touch := send(allTouches, selObjectAtIndex, i)
		phase := int(send(touch, selPhase))
		coalesced := send(event, selCoalescedTouchesForTouch, touch)
		m := send(coalesced, selCount)
		for j := uintptr(0); j < m; j++ {
			ct := send(coalesced, selObjectAtIndex, j)
			loc := msgPointID(ct, selLocationInView, view)
			lastTouch := last && i+1 == n && j+1 == m
			w.onTouch(lastTouch, touch, phase, loc.X*scale, loc.Y*scale, msgF64(ct, selTimestamp))
		}
	}
}

func iosViewMethods() []objc.MethodDef {
	return []objc.MethodDef{
		// ui/window/scene_ios.go inspects view.handle before replaying initial focus.
		{Cmd: objc.RegisterName("handle"), Fn: func(self objc.ID, _ objc.SEL) uintptr {
			if windowOf(uintptr(self)) != nil {
				return uintptr(self)
			}
			return 0
		}},
		{Cmd: objc.RegisterName("willMoveToWindow:"), Fn: func(self objc.ID, _ objc.SEL, newWindow uintptr) {
			if newWindow != 0 {
				if scr := send(newWindow, selScreen); scr != 0 {
					msgSetF64(uintptr(self), selSetContentScaleFactor, msgF64(scr, selNativeScale))
				}
			}
			if uiSceneDidActivateNotification != 0 && send(classUIWindow, selInstancesRespondToSel, selWindowScene)&0xff != 0 {
				registerSceneNotifications(uintptr(self), newWindow)
			} else {
				registerWindowNotifications(uintptr(self), newWindow)
			}
		}},
		{Cmd: objc.SEL(selOnSceneDidActivate), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onFocus(true)
			}
		}},
		{Cmd: objc.SEL(selOnSceneWillDeactivate), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onFocus(false)
			}
		}},
		{Cmd: objc.SEL(selOnWindowDidBecomeKey), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			if send(uintptr(self), selIsFirstResponder)&0xff != 0 {
				if w := windowOf(uintptr(self)); w != nil {
					w.onFocus(true)
				}
			}
		}},
		{Cmd: objc.SEL(selOnWindowDidResignKey), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			if send(uintptr(self), selIsFirstResponder)&0xff != 0 {
				if w := windowOf(uintptr(self)); w != nil {
					w.onFocus(false)
				}
			}
		}},
		{Cmd: objc.RegisterName("touchesBegan:withEvent:"), Fn: func(self objc.ID, _ objc.SEL, touches, event uintptr) {
			handleTouches(false, uintptr(self), touches, event)
		}},
		{Cmd: objc.RegisterName("touchesMoved:withEvent:"), Fn: func(self objc.ID, _ objc.SEL, touches, event uintptr) {
			handleTouches(false, uintptr(self), touches, event)
		}},
		{Cmd: objc.RegisterName("touchesEnded:withEvent:"), Fn: func(self objc.ID, _ objc.SEL, touches, event uintptr) {
			handleTouches(true, uintptr(self), touches, event)
		}},
		{Cmd: objc.RegisterName("touchesCancelled:withEvent:"), Fn: func(self objc.ID, _ objc.SEL, touches, event uintptr) {
			handleTouches(true, uintptr(self), touches, event)
		}},
		{Cmd: objc.RegisterName("insertText:"), Fn: func(self objc.ID, _ objc.SEL, text uintptr) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onText(nsstringToString(text))
			}
		}},
		{Cmd: objc.RegisterName("canBecomeFirstResponder"), Fn: func(objc.ID, objc.SEL) bool {
			return true
		}},
		{Cmd: objc.RegisterName("hasText"), Fn: func(objc.ID, objc.SEL) bool {
			return true
		}},
		{Cmd: objc.RegisterName("deleteBackward"), Fn: func(self objc.ID, _ objc.SEL) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onKeyCommand(key.NameDeleteBackward)
			}
		}},
		{Cmd: objc.SEL(selOnUpArrow), Fn: func(self objc.ID, _ objc.SEL) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onKeyCommand(key.NameUpArrow)
			}
		}},
		{Cmd: objc.SEL(selOnDownArrow), Fn: func(self objc.ID, _ objc.SEL) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onKeyCommand(key.NameDownArrow)
			}
		}},
		{Cmd: objc.SEL(selOnLeftArrow), Fn: func(self objc.ID, _ objc.SEL) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onKeyCommand(key.NameLeftArrow)
			}
		}},
		{Cmd: objc.SEL(selOnRightArrow), Fn: func(self objc.ID, _ objc.SEL) {
			if w := windowOf(uintptr(self)); w != nil {
				w.onKeyCommand(key.NameRightArrow)
			}
		}},
		{Cmd: objc.RegisterName("keyCommands"), Fn: func(objc.ID, objc.SEL) uintptr {
			if iosKeyCommands == 0 && uiKeyInputUpArrow != 0 {
				cmds := [4]uintptr{
					send(classUIKeyCommand, selKeyCommandWithInput, uiKeyInputUpArrow, 0, selOnUpArrow),
					send(classUIKeyCommand, selKeyCommandWithInput, uiKeyInputDownArrow, 0, selOnDownArrow),
					send(classUIKeyCommand, selKeyCommandWithInput, uiKeyInputLeftArrow, 0, selOnLeftArrow),
					send(classUIKeyCommand, selKeyCommandWithInput, uiKeyInputRightArrow, 0, selOnRightArrow),
				}
				iosKeyCommands = send(send(classNSArray, selAlloc), selInitWithObjectsCount, uintptr(unsafe.Pointer(&cmds[0])), uintptr(len(cmds)))
				runtime.KeepAlive(cmds)
			}
			return iosKeyCommands
		}},
		{Cmd: objc.SEL(selDealloc), Fn: func(self objc.ID, _ objc.SEL) {
			send(send(classNSNotificationCenter, selDefaultCenter), selRemoveObserver, uintptr(self))
			deleteIOSView(uintptr(self))
			sendSuper(uintptr(self), classGioView, selDealloc)
		}},
	}
}

func viewDrawParams(view uintptr) iosDrawParams {
	var p iosDrawParams
	withPool(func() {
		layer := send(view, selLayer)
		scale := msgF64(layer, selContentsScale)
		// Use 163 as the standard ppi on iOS.
		p.dpi = 163 * scale
		p.sdpi = p.dpi
		insets := msgEdgeInsets(view, selLayoutMargins)
		if classUIFontMetrics != 0 {
			if metrics := send(classUIFontMetrics, selDefaultMetrics); metrics != 0 {
				p.sdpi = msgF64F64(metrics, selScaledValueForValue, p.sdpi)
			}
			insets = msgEdgeInsets(view, selSafeAreaInsets)
		}
		bounds := msgRect(view, selBounds)
		p.width = bounds.Size.X * scale
		p.height = bounds.Size.Y * scale
		p.top = insets.Top * scale
		p.right = insets.Right * scale
		p.bottom = insets.Bottom * scale
		p.left = insets.Left * scale
	})
	return p
}

func showTextInput(view uintptr) {
	withPool(func() { send(view, selBecomeFirstResponder) })
}

func hideTextInput(view uintptr) {
	withPool(func() { send(view, selResignFirstResponder) })
}

func readClipboard() string {
	var content string
	withPool(func() {
		pb := send(classUIPasteboard, selGeneralPasteboard)
		content = nsstringToString(send(pb, selString))
	})
	return content
}

func writeClipboard(s string) {
	withPool(func() {
		ns := stringToNSString(s)
		defer releaseObject(ns)
		pb := send(classUIPasteboard, selGeneralPasteboard)
		send(pb, selSetString, ns)
	})
}

func iosApplicationMain(args []string) {
	withPool(func() {
		cArgs := make([][]byte, len(args))
		argv := make([]*byte, len(args))
		for i, arg := range args {
			cArgs[i] = append([]byte(arg), 0)
			argv[i] = &cArgs[i][0]
		}
		var argvPtr uintptr
		if len(argv) > 0 {
			argvPtr = uintptr(unsafe.Pointer(&argv[0]))
		}
		delegateName := stringToNSString("_gioAppDelegate")
		defer releaseObject(delegateName)
		purego.SyscallN(iosRT.applicationMain, uintptr(len(argv)), argvPtr, 0, delegateName)
		runtime.KeepAlive(cArgs)
		runtime.KeepAlive(argv)
	})
}

func createDisplayLink() uintptr {
	var dl uintptr
	withPool(func() {
		dl = retainObject(send(classCADisplayLink, selDisplayLinkWithTarget, classGioView, selOnFrameCallback))
		if dl == 0 {
			return
		}
		send(dl, selSetPaused, 1)
		runLoop := send(classNSRunLoop, selMainRunLoop)
		mode := send(runLoop, selCurrentMode)
		send(dl, selAddToRunLoopForMode, runLoop, mode)
	})
	return dl
}

func startDisplayLink(dl uintptr) {
	withPool(func() { send(dl, selSetPaused, 0) })
}

func stopDisplayLink(dl uintptr) {
	withPool(func() { send(dl, selSetPaused, 1) })
}

func releaseDisplayLink(dl uintptr) {
	withPool(func() {
		send(dl, selInvalidate)
		releaseObject(dl)
	})
}

func setDisplayLinkDisplay(dl uintptr, did uint64) {
	// Nothing to do on iOS.
}

func hideCursor()       {}
func showCursor()       {}
func setCursor(id uint) {}
