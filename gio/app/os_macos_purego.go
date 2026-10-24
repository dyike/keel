// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && (!cgo || purego)

package app

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The AppKit binding of os_macos.go without cgo: os_macos.m's classes,
// registered at run time with the same names (Keel looks GioView up by
// name), and the helpers of os_macos_cgo.go as messages. Keel patch.
//
// Each class is registered once. A view finds its window through a map
// keyed by the view pointer; the map entry goes when the view deallocates.

// Typed messages, registered once: float and struct arguments and results
// need the C calling convention that RegisterFunc implements.
var (
	msgF64          func(id, sel uintptr) float64
	msgSetF64       func(id, sel uintptr, v float64)
	msgPoint        func(id, sel uintptr) nsPoint
	msgPointPoint   func(id, sel uintptr, p nsPoint) nsPoint
	msgPointPointID func(id, sel uintptr, p nsPoint, view uintptr) nsPoint
	msgSetSize      func(id, sel uintptr, s nsPoint)
	msgRect         func(id, sel uintptr) nsRect
	msgSetRect      func(id, sel uintptr, r nsRect)
	msgSetRectBool  func(id, sel uintptr, r nsRect, display bool)
	msgRectRect     func(id, sel uintptr, r nsRect) nsRect
	msgRectRectID   func(id, sel uintptr, r nsRect, view uintptr) nsRect
	msgInitRect     func(id, sel uintptr, r nsRect) uintptr
	msgInitWindow   func(id, sel uintptr, r nsRect, style, backing uint, deferred bool) uintptr
	msgSuperSetSize func(sup *objcSuper, sel uintptr, s nsPoint)
)

var (
	selWindow                = sel("window")
	selBounds                = sel("bounds")
	selFrame                 = sel("frame")
	selSetFrame              = sel("setFrame:")
	selSetFrameDisplay       = sel("setFrame:display:")
	selLayer                 = sel("layer")
	selSetDelegate           = sel("setDelegate:")
	selDelegate              = sel("delegate")
	selBackingScaleFactor    = sel("backingScaleFactor")
	selMainScreen            = sel("mainScreen")
	selScreen                = sel("screen")
	selSetNeedsDisplay       = sel("setNeedsDisplay:")
	selCascade               = sel("cascadeTopLeftFromPoint:")
	selOcclusionState        = sel("occlusionState")
	selMakeKeyAndOrderFront  = sel("makeKeyAndOrderFront:")
	selMakeFirstResponder    = sel("makeFirstResponder:")
	selFirstResponder        = sel("firstResponder")
	selToggleFullScreen      = sel("toggleFullScreen:")
	selStyleMask             = sel("styleMask")
	selSetStyleMask          = sel("setStyleMask:")
	selSetTitleVisibility    = sel("setTitleVisibility:")
	selSetTitlebarTransp     = sel("setTitlebarAppearsTransparent:")
	selStandardWindowButton  = sel("standardWindowButton:")
	selSetHidden             = sel("setHidden:")
	selPerformWindowDrag     = sel("performWindowDragWithEvent:")
	selPerformClose          = sel("performClose:")
	selSetContentSize        = sel("setContentSize:")
	selSetContentMinSize     = sel("setContentMinSize:")
	selSetContentMaxSize     = sel("setContentMaxSize:")
	selSetMaxFullScreenSize  = sel("setMaxFullScreenContentSize:")
	selMiniaturize           = sel("miniaturize:")
	selDeminiaturize         = sel("deminiaturize:")
	selSetTitle              = sel("setTitle:")
	selSetLevel              = sel("setLevel:")
	selIsZoomed              = sel("isZoomed")
	selIsMiniaturized        = sel("isMiniaturized")
	selZoom                  = sel("zoom:")
	selActivate              = sel("activateIgnoringOtherApps:")
	selCurrentInputContext   = sel("currentInputContext")
	selClient                = sel("client")
	selDiscardMarkedText     = sel("discardMarkedText")
	selInvalidateCharCoords  = sel("invalidateCharacterCoordinates")
	selInterpretKeyEvents    = sel("interpretKeyEvents:")
	selArrayWithObject       = sel("arrayWithObject:")
	selGeneralPasteboard     = sel("generalPasteboard")
	selDeclareTypes          = sel("declareTypes:owner:")
	selSetStringForType      = sel("setString:forType:")
	selStringForType         = sel("stringForType:")
	selInitWithFrame         = sel("initWithFrame:")
	selSetWantsLayer         = sel("setWantsLayer:")
	selSetRedrawPolicy       = sel("setLayerContentsRedrawPolicy:")
	selDefaultCenter         = sel("defaultCenter")
	selAddObserver           = sel("addObserver:selector:name:object:")
	selInitWithContentRect   = sel("initWithContentRect:styleMask:backing:defer:")
	selSetAcceptsMouseMoved  = sel("setAcceptsMouseMovedEvents:")
	selSetContentView        = sel("setContentView:")
	selSetContentsScale      = sel("setContentsScale:")
	selLocationInWindow      = sel("locationInWindow")
	selConvertPointFromView  = sel("convertPoint:fromView:")
	selHasPreciseScrolling   = sel("hasPreciseScrollingDeltas")
	selButtonNumber          = sel("buttonNumber")
	selTimestamp             = sel("timestamp")
	selModifierFlags         = sel("modifierFlags")
	selScrollingDeltaX       = sel("scrollingDeltaX")
	selScrollingDeltaY       = sel("scrollingDeltaY")
	selCharsIgnoringMods     = sel("charactersIgnoringModifiers")
	selDoCommandBySelector   = sel("doCommandBySelector:")
	selInitWithString        = sel("initWithString:attributes:")
	selAutorelease           = sel("autorelease")
	selConvertRectToView     = sel("convertRect:toView:")
	selConvertRectToScreen   = sel("convertRectToScreen:")
	selBecomeFirstResponder  = sel("becomeFirstResponder")
	selResignFirstResponder  = sel("resignFirstResponder")
	selIsKeyWindow           = sel("isKeyWindow")
	selDealloc               = sel("dealloc")
	selObject                = sel("object")
	selDeviceDescription     = sel("deviceDescription")
	selObjectForKey          = sel("objectForKey:")
	selUnsignedIntValue      = sel("unsignedIntValue")
	selBoolValue             = sel("boolValue")
	selCount                 = sel("count")
	selObjectAtIndex         = sel("objectAtIndex:")
	selAbsoluteString        = sel("absoluteString")
	selSharedApplication     = sel("sharedApplication")
	selMainBundle            = sel("mainBundle")
	selObjectForInfoKey      = sel("objectForInfoDictionaryKey:")
	selSetActivationPolicy   = sel("setActivationPolicy:")
	selNew                   = sel("new")
	selInitWithTitle         = sel("initWithTitle:action:keyEquivalent:")
	selAddItem               = sel("addItem:")
	selSetSubmenu            = sel("setSubmenu:")
	selSetMainMenu           = sel("setMainMenu:")
	selRun                   = sel("run")
	selHideApp               = sel("hide:")
	selTerminate             = sel("terminate:")
	selStringWithUTF8String  = sel("stringWithUTF8String:")
	selGioInputView          = sel("gioInputView")
	selGioHostWindow         = sel("gioHostWindow")
	selApplicationWillUnhide = sel("applicationWillUnhide:")
	selApplicationDidHide    = sel("applicationDidHide:")
	selLaunchFinished        = sel("launchFinished:")
)

const (
	nsWindowOcclusionStateVisible             = 1 << 1
	nsBackingStoreBuffered                    = 2
	nsViewLayerContentsRedrawDuringViewResize = 2
	nsActivationPolicyRegular                 = 0
	nsActivationPolicyAccessory               = 1
)

var (
	classGioView, classGioWindowDelegate, classGioAppDelegate, classAppListener uintptr

	classNSWindow, classNSApplication, classNSPasteboard, classNSArray uintptr
	classNSTextInputContext, classNSNotificationCenter, classNSScreen  uintptr
	classNSBundle, classNSMenu, classNSMenuItem                        uintptr

	nsApp           uintptr
	globalWindowDel uintptr
	appListener     uintptr

	// AppKit string constants.
	nsPasteboardTypeString, nsAppDidFinishLaunching uintptr
	nsAppWillUnhide, nsAppDidHide                   uintptr
)

// views maps a GioView to its window and creation options; AppKit calls
// the view's methods on the main thread only, but deallocation order is
// AppKit's.
var views struct {
	sync.Mutex
	m map[uintptr]*viewState
}

type viewState struct {
	w                *window
	presentWithTrans bool
}

func viewFor(view uintptr) *viewState {
	views.Lock()
	defer views.Unlock()
	return views.m[view]
}

// windowOf returns the window of a view, or nil before it has one.
func windowOf(view objc.ID) *window {
	if s := viewFor(uintptr(view)); s != nil {
		return s.w
	}
	return nil
}

func appInit() {
	for _, f := range []any{&msgF64, &msgSetF64, &msgPoint, &msgPointPoint, &msgPointPointID, &msgSetSize, &msgSetRect, &msgSetRectBool, &msgInitRect, &msgInitWindow} {
		purego.RegisterFunc(f, rt.msg)
	}
	for _, f := range []any{&msgRect, &msgRectRect, &msgRectRectID} {
		purego.RegisterFunc(f, rt.msgStret)
	}
	purego.RegisterFunc(&msgSuperSetSize, rt.msgSuper)

	for _, c := range []struct {
		addr *uintptr
		name string
	}{
		{&classNSWindow, "NSWindow"},
		{&classNSApplication, "NSApplication"},
		{&classNSPasteboard, "NSPasteboard"},
		{&classNSArray, "NSArray"},
		{&classNSTextInputContext, "NSTextInputContext"},
		{&classNSNotificationCenter, "NSNotificationCenter"},
		{&classNSScreen, "NSScreen"},
		{&classNSBundle, "NSBundle"},
		{&classNSMenu, "NSMenu"},
		{&classNSMenuItem, "NSMenuItem"},
	} {
		*c.addr = class(c.name)
	}
	for _, c := range []struct {
		addr *uintptr
		name string
	}{
		{&nsPasteboardTypeString, "NSPasteboardTypeString"},
		{&nsAppDidFinishLaunching, "NSApplicationDidFinishLaunchingNotification"},
		{&nsAppWillUnhide, "NSApplicationWillUnhideNotification"},
		{&nsAppDidHide, "NSApplicationDidHideNotification"},
	} {
		addr, err := purego.Dlsym(purego.RTLD_DEFAULT, c.name)
		if err != nil {
			panic(err)
		}
		*c.addr = **(**uintptr)(unsafe.Pointer(&addr)) // the NSString the symbol holds
	}
	registerClasses()
	withPool(func() {
		appListener = send(send(classAppListener, selAlloc), selInit)
		center := send(classNSNotificationCenter, selDefaultCenter)
		send(center, selAddObserver, appListener, selLaunchFinished, nsAppDidFinishLaunching, 0)
	})
}

func appMain() {
	withPool(func() {
		nsApp = send(classNSApplication, selSharedApplication)
		del := send(send(classGioAppDelegate, selAlloc), selInit)
		send(nsApp, selSetDelegate, del)

		mainMenu := send(classNSMenuItem, selNew)
		menu := send(classNSMenu, selNew)
		for _, item := range []struct {
			title, key string
			action     uintptr
		}{
			{"Hide", "h", selHideApp},
			{"Quit", "q", selTerminate},
		} {
			title, key := stringToNSString(item.title), stringToNSString(item.key)
			mi := send(send(classNSMenuItem, selAlloc), selInitWithTitle, title, item.action, key)
			releaseObject(title)
			releaseObject(key)
			send(menu, selAddItem, mi)
			releaseObject(mi)
		}
		send(mainMenu, selSetSubmenu, menu)
		releaseObject(menu)
		menuBar := send(classNSMenu, selNew)
		send(menuBar, selAddItem, mainMenu)
		releaseObject(mainMenu)
		send(nsApp, selSetMainMenu, menuBar)
		releaseObject(menuBar)

		// Delegates are weakly referenced from their peers; keep the one
		// window delegate forever, as os_macos.m does.
		globalWindowDel = send(send(classGioWindowDelegate, selAlloc), selInit)

		send(nsApp, selRun)
	})
}

func registerClasses() {
	var err error
	classAppListener, err = registerClass("AppListener", "NSObject", nil, []objc.MethodDef{
		{Cmd: objc.SEL(selLaunchFinished), Fn: func(self objc.ID, _ objc.SEL, _ uintptr) {
			releaseObject(appListener)
			appListener = 0
			onFinishLaunching()
		}},
	})
	if err == nil {
		classGioAppDelegate, err = registerClass("GioAppDelegate", "NSObject", nil, appDelegateMethods())
	}
	if err == nil {
		classGioWindowDelegate, err = registerClass("GioWindowDelegate", "NSObject", nil, windowDelegateMethods())
	}
	if err == nil {
		var protocols []*objc.Protocol
		// NSView's inputContext exists only for NSTextInputClient views.
		for _, name := range []string{"NSTextInputClient", "CALayerDelegate"} {
			if p := objc.GetProtocol(name); p != nil {
				protocols = append(protocols, p)
			}
		}
		classGioView, err = registerClass("GioView", "NSView", protocols, viewMethods())
	}
	if err != nil {
		panic(err)
	}
}

func registerClass(name, super string, protocols []*objc.Protocol, methods []objc.MethodDef) (uintptr, error) {
	c, err := objc.RegisterClass(name, objc.Class(class(super)), protocols, nil, methods)
	return uintptr(c), err
}

func appDelegateMethods() []objc.MethodDef {
	return []objc.MethodDef{
		{Cmd: objc.RegisterName("applicationDidFinishLaunching:"), Fn: func(objc.ID, objc.SEL, uintptr) {
			// Keel patch: preserve accessory applications' bundle launch policy.
			key := stringToNSString("LSUIElement")
			defer releaseObject(key)
			v := send(send(classNSBundle, selMainBundle), selObjectForInfoKey, key)
			policy := uintptr(nsActivationPolicyRegular)
			if v != 0 && send(v, selBoolValue)&0xff != 0 {
				policy = nsActivationPolicyAccessory
			}
			send(nsApp, selSetActivationPolicy, policy)
			send(nsApp, selActivate, 1)
		}},
		{Cmd: objc.RegisterName("application:openURLs:"), Fn: func(_ objc.ID, _ objc.SEL, _, urls uintptr) {
			n := send(urls, selCount)
			for i := uintptr(0); i < n; i++ {
				url := send(urls, selObjectAtIndex, i)
				onOpenURI(nsstringToString(send(url, selAbsoluteString)))
			}
		}},
	}
}

// notificationWindow returns the window and Gio window of a notification
// about an NSWindow.
func notificationWindow(n uintptr) (uintptr, uintptr, *window) {
	win := send(n, selObject)
	view := associated(win, selGioInputView)
	if view == 0 {
		return win, 0, nil
	}
	return win, view, windowOf(objc.ID(view))
}

func windowDelegateMethods() []objc.MethodDef {
	draw := func(_ objc.ID, _ objc.SEL, n uintptr) {
		if _, _, w := notificationWindow(n); w != nil {
			w.draw()
		}
	}
	return []objc.MethodDef{
		{Cmd: objc.RegisterName("windowWillClose:"), Fn: func(_ objc.ID, _ objc.SEL, n uintptr) {
			if _, view, _ := notificationWindow(n); view != 0 {
				// The inverse reference is non-owning; never leave a closed host behind.
				setAssociated(view, selGioHostWindow, 0, objcAssociationAssign)
			}
		}},
		{Cmd: objc.RegisterName("windowShouldClose:"), Fn: func(_ objc.ID, _ objc.SEL, win uintptr) bool {
			view := associated(win, selGioInputView)
			if view == 0 {
				return true
			}
			w := windowOf(objc.ID(view))
			return w == nil || w.onCloseRequest()
		}},
		{Cmd: objc.RegisterName("windowWillMiniaturize:"), Fn: draw},
		{Cmd: objc.RegisterName("windowDidDeminiaturize:"), Fn: draw},
		{Cmd: objc.RegisterName("windowWillEnterFullScreen:"), Fn: draw},
		{Cmd: objc.RegisterName("windowWillExitFullScreen:"), Fn: draw},
		// Keel patch: windows nobody can see (covered, on another space, the
		// screen locked) stop drawing; see onOcclusion.
		{Cmd: objc.RegisterName("windowDidChangeOcclusionState:"), Fn: func(_ objc.ID, _ objc.SEL, n uintptr) {
			if win, _, w := notificationWindow(n); w != nil {
				w.onOcclusion(send(win, selOcclusionState)&nsWindowOcclusionStateVisible != 0)
			}
		}},
		{Cmd: objc.RegisterName("windowDidChangeScreen:"), Fn: func(_ objc.ID, _ objc.SEL, n uintptr) {
			win, _, w := notificationWindow(n)
			if w == nil {
				return
			}
			key := stringToNSString("NSScreenNumber")
			defer releaseObject(key)
			num := send(send(send(win, selScreen), selDeviceDescription), selObjectForKey, key)
			w.onChangeScreen(uint64(uint32(send(num, selUnsignedIntValue))))
		}},
		{Cmd: objc.RegisterName("windowDidBecomeKey:"), Fn: func(_ objc.ID, _ objc.SEL, n uintptr) {
			if win, view, w := notificationWindow(n); w != nil && send(win, selFirstResponder) == view {
				w.onFocus(true)
			}
		}},
		{Cmd: objc.RegisterName("windowDidResignKey:"), Fn: func(_ objc.ID, _ objc.SEL, n uintptr) {
			if _, _, w := notificationWindow(n); w != nil {
				w.onFocus(false)
			}
		}},
	}
}

// handleMouse reports a mouse event at its position in the view, from the
// upper left corner.
func handleMouse(self objc.ID, event uintptr, typ int, dx, dy float64) {
	w := windowOf(self)
	if w == nil {
		return
	}
	view := uintptr(self)
	p := msgPointPointID(view, selConvertPointFromView, msgPoint(event, selLocationInWindow), 0)
	if send(event, selHasPreciseScrolling)&0xff == 0 {
		// dx and dy are in rows and columns.
		dx *= 10
		dy *= 10
	}
	height := msgRect(view, selBounds).Size.Y
	w.onMouse(event, typ, int(send(event, selButtonNumber)), p.X, height-p.Y, dx, dy, msgF64(event, selTimestamp), uint(send(event, selModifierFlags)))
}

func viewMethods() []objc.MethodDef {
	mouse := func(typ int) func(objc.ID, objc.SEL, uintptr) {
		return func(self objc.ID, _ objc.SEL, event uintptr) { handleMouse(self, event, typ, 0, 0) }
	}
	draw := func(self objc.ID, _ objc.SEL, _ uintptr) {
		if w := windowOf(self); w != nil {
			w.draw()
		}
	}
	keys := func(down bool) func(objc.ID, objc.SEL, uintptr) {
		return func(self objc.ID, _ objc.SEL, event uintptr) {
			if w := windowOf(self); w != nil {
				str := nsstringToString(send(event, selCharsIgnoringMods))
				w.onKeys(event, str, uint(send(event, selModifierFlags)), down)
			}
		}
	}
	return []objc.MethodDef{
		{Cmd: objc.RegisterName("setFrameSize:"), Fn: func(self objc.ID, cmd objc.SEL, size nsPoint) {
			msgSuperSetSize(&objcSuper{uintptr(self), classGioView}, uintptr(cmd), size)
			send(uintptr(self), selSetNeedsDisplay, 1)
		}},
		// drawRect is called when OpenGL is used, displayLayer otherwise.
		{Cmd: objc.RegisterName("drawRect:"), Fn: func(self objc.ID, _ objc.SEL, _ nsRect) {
			if w := windowOf(self); w != nil {
				w.draw()
			}
		}},
		{Cmd: objc.RegisterName("displayLayer:"), Fn: func(self objc.ID, _ objc.SEL, layer uintptr) {
			msgSetF64(layer, selSetContentsScale, msgF64(send(uintptr(self), selWindow), selBackingScaleFactor))
			if w := windowOf(self); w != nil {
				w.draw()
			}
		}},
		{Cmd: objc.RegisterName("makeBackingLayer"), Fn: func(self objc.ID, _ objc.SEL) uintptr {
			s := viewFor(uintptr(self))
			layer := layerFactory(s != nil && s.presentWithTrans)
			send(layer, selSetDelegate, uintptr(self))
			return layer
		}},
		{Cmd: objc.RegisterName("viewDidMoveToWindow"), Fn: func(self objc.ID, _ objc.SEL) {
			if w := windowOf(self); w != nil {
				w.onAttached(send(uintptr(self), selWindow) != 0)
			}
		}},
		{Cmd: objc.RegisterName("mouseDown:"), Fn: mouse(mouseDown)},
		{Cmd: objc.RegisterName("mouseUp:"), Fn: mouse(mouseUp)},
		{Cmd: objc.RegisterName("rightMouseDown:"), Fn: mouse(mouseDown)},
		{Cmd: objc.RegisterName("rightMouseUp:"), Fn: mouse(mouseUp)},
		{Cmd: objc.RegisterName("otherMouseDown:"), Fn: mouse(mouseDown)},
		{Cmd: objc.RegisterName("otherMouseUp:"), Fn: mouse(mouseUp)},
		{Cmd: objc.RegisterName("mouseMoved:"), Fn: mouse(mouseMove)},
		{Cmd: objc.RegisterName("mouseDragged:"), Fn: mouse(mouseMove)},
		{Cmd: objc.RegisterName("rightMouseDragged:"), Fn: mouse(mouseMove)},
		{Cmd: objc.RegisterName("otherMouseDragged:"), Fn: mouse(mouseMove)},
		{Cmd: objc.RegisterName("scrollWheel:"), Fn: func(self objc.ID, _ objc.SEL, event uintptr) {
			dx := -msgF64(event, selScrollingDeltaX)
			dy := -msgF64(event, selScrollingDeltaY)
			handleMouse(self, event, mouseScroll, dx, dy)
		}},
		{Cmd: objc.RegisterName("keyDown:"), Fn: keys(true)},
		{Cmd: objc.RegisterName("keyUp:"), Fn: keys(false)},
		{Cmd: objc.RegisterName("flagsChanged:"), Fn: func(self objc.ID, _ objc.SEL, event uintptr) {
			interpretKeyEvents(uintptr(self), event)
			if w := windowOf(self); w != nil {
				w.onFlagsChanged(uint(send(event, selModifierFlags)))
			}
		}},
		{Cmd: objc.RegisterName("insertText:"), Fn: func(self objc.ID, _ objc.SEL, str uintptr) {
			if w := windowOf(self); w != nil {
				w.onText(nsstringToString(plainString(str)))
			}
		}},
		{Cmd: objc.RegisterName("doCommandBySelector:"), Fn: func(self objc.ID, cmd objc.SEL, action uintptr) {
			if w := windowOf(self); w == nil || !w.onCommandBySelector() {
				sendSuper(uintptr(self), classGioView, uintptr(cmd), action)
			}
		}},
		{Cmd: objc.RegisterName("hasMarkedText"), Fn: func(self objc.ID, _ objc.SEL) bool {
			w := windowOf(self)
			return w != nil && w.hasMarkedText()
		}},
		{Cmd: objc.RegisterName("markedRange"), Fn: func(self objc.ID, _ objc.SEL) nsRange {
			if w := windowOf(self); w != nil {
				return w.markedRange()
			}
			return nsRange{Location: nsNotFound}
		}},
		{Cmd: objc.RegisterName("selectedRange"), Fn: func(self objc.ID, _ objc.SEL) nsRange {
			if w := windowOf(self); w != nil {
				return w.selectedRange()
			}
			return nsRange{Location: nsNotFound}
		}},
		{Cmd: objc.RegisterName("unmarkText"), Fn: func(self objc.ID, _ objc.SEL) {
			if w := windowOf(self); w != nil {
				w.unmarkText()
			}
		}},
		{Cmd: objc.RegisterName("setMarkedText:selectedRange:replacementRange:"), Fn: func(self objc.ID, _ objc.SEL, str uintptr, selRange, replaceRange nsRange) {
			if w := windowOf(self); w != nil {
				// string is either an NSAttributedString or an NSString.
				w.setMarkedText(nsstringToString(plainString(str)), selRange, replaceRange)
			}
		}},
		{Cmd: objc.RegisterName("validAttributesForMarkedText"), Fn: func(objc.ID, objc.SEL) uintptr {
			return 0
		}},
		{Cmd: objc.RegisterName("attributedSubstringForProposedRange:actualRange:"), Fn: func(self objc.ID, _ objc.SEL, rng nsRange, actual *nsRange) uintptr {
			w := windowOf(self)
			if w == nil {
				return 0
			}
			text, got := w.substringForProposedRange(rng)
			if actual != nil {
				*actual = got
			}
			str := stringToNSString(text)
			defer releaseObject(str)
			attr := send(send(classNSAttributedStr, selAlloc), selInitWithString, str, 0)
			return send(attr, selAutorelease)
		}},
		{Cmd: objc.RegisterName("insertText:replacementRange:"), Fn: func(self objc.ID, _ objc.SEL, str uintptr, rng nsRange) {
			if w := windowOf(self); w != nil {
				// string is either an NSAttributedString or an NSString.
				w.insertText(nsstringToString(plainString(str)), rng)
			}
		}},
		{Cmd: objc.RegisterName("characterIndexForPoint:"), Fn: func(objc.ID, objc.SEL, nsPoint) uint {
			return nsNotFound
		}},
		{Cmd: objc.RegisterName("firstRectForCharacterRange:actualRange:"), Fn: func(self objc.ID, _ objc.SEL, rng nsRange, actual *nsRange) nsRect {
			w := windowOf(self)
			if w == nil {
				return nsRect{}
			}
			r, got := w.firstRectForCharacterRange(rng)
			if actual != nil {
				*actual = got
			}
			view := uintptr(self)
			r = msgRectRectID(view, selConvertRectToView, r, 0)
			return msgRectRect(send(view, selWindow), selConvertRectToScreen, r)
		}},
		{Cmd: objc.SEL(selApplicationWillUnhide), Fn: draw},
		{Cmd: objc.SEL(selApplicationDidHide), Fn: draw},
		{Cmd: objc.SEL(selDealloc), Fn: func(self objc.ID, cmd objc.SEL) {
			views.Lock()
			s := views.m[uintptr(self)]
			delete(views.m, uintptr(self))
			views.Unlock()
			// Keel patch: a view whose window failed to initialize never got one.
			if s != nil && s.w != nil {
				s.w.onDestroy()
			}
			sendSuper(uintptr(self), classGioView, uintptr(cmd))
		}},
		{Cmd: objc.SEL(selBecomeFirstResponder), Fn: func(self objc.ID, cmd objc.SEL) bool {
			accepted := sendSuper(uintptr(self), classGioView, uintptr(cmd))&0xff != 0
			// A responder can be assigned before its accessory app is activated.
			// Report keyboard focus only after the native window actually becomes key.
			if accepted && send(send(uintptr(self), selWindow), selIsKeyWindow)&0xff != 0 {
				if w := windowOf(self); w != nil {
					w.onFocus(true)
				}
			}
			return accepted
		}},
		{Cmd: objc.RegisterName("acceptsFirstResponder"), Fn: func(objc.ID, objc.SEL) bool {
			return true
		}},
		// Reparenting GioView into a glass root temporarily changes responders
		// while the window remains key. The window delegate owns focus loss.
		{Cmd: objc.SEL(selResignFirstResponder), Fn: func(self objc.ID, cmd objc.SEL) bool {
			return sendSuper(uintptr(self), classGioView, uintptr(cmd))&0xff != 0
		}},
	}
}

func createView(presentWithTrans bool) uintptr {
	var view uintptr
	withPool(func() {
		view = send(classGioView, selAlloc)
		// The backing layer is made while the view becomes layer-backed, and
		// reads the option; register it first.
		views.Lock()
		if views.m == nil {
			views.m = make(map[uintptr]*viewState)
		}
		views.m[view] = &viewState{presentWithTrans: presentWithTrans}
		views.Unlock()
		if msgInitRect(view, selInitWithFrame, nsRect{}) != view {
			// NSView's initializer returns its receiver; anything else would
			// leave the map pointing at a different object.
			panic("app: -[GioView initWithFrame:] returned another object")
		}
		send(view, selSetWantsLayer, 1)
		send(view, selSetRedrawPolicy, nsViewLayerContentsRedrawDuringViewResize)
		center := send(classNSNotificationCenter, selDefaultCenter)
		send(center, selAddObserver, view, selApplicationWillUnhide, nsAppWillUnhide, 0)
		send(center, selAddObserver, view, selApplicationDidHide, nsAppDidHide, 0)
	})
	return view
}

func viewSetWindow(view uintptr, w *window) {
	views.Lock()
	defer views.Unlock()
	views.m[view].w = w
}

func createWindow(view uintptr, width, height float64) uintptr {
	var window uintptr
	withPool(func() {
		style := uint(nsWindowStyleMaskTitled | nsWindowStyleMaskResizable | nsWindowStyleMaskMiniaturizable | nsWindowStyleMaskClosable)
		r := nsRect{Size: nsPoint{X: width, Y: height}}
		window = msgInitWindow(send(classNSWindow, selAlloc), selInitWithContentRect, r, style, nsBackingStoreBuffered, false)
		send(window, selSetAcceptsMouseMoved, 1)
		// Keel patch: the delegate first, so the view sees it when it attaches.
		setAssociated(window, selGioInputView, view, objcAssociationRetainNonatomic)
		setAssociated(view, selGioHostWindow, window, objcAssociationAssign)
		send(window, selSetDelegate, globalWindowDel)
		send(window, selSetContentView, view)
	})
	return window
}

func readClipboard() string {
	var s string
	withPool(func() {
		p := send(classNSPasteboard, selGeneralPasteboard)
		s = nsstringToString(send(p, selStringForType, nsPasteboardTypeString))
	})
	return s
}

func writeClipboard(s string) {
	withPool(func() {
		str := stringToNSString(s)
		defer releaseObject(str)
		p := send(classNSPasteboard, selGeneralPasteboard)
		send(p, selDeclareTypes, send(classNSArray, selArrayWithObject, nsPasteboardTypeString), 0)
		send(p, selSetStringForType, str, nsPasteboardTypeString)
	})
}

func viewHeight(v uintptr) float64 { return msgRect(v, selBounds).Size.Y }
func viewWidth(v uintptr) float64  { return msgRect(v, selBounds).Size.X }

func screenBackingScale() float64 {
	return msgF64(send(classNSScreen, selMainScreen), selBackingScaleFactor)
}

func viewBackingScale(v uintptr) float64 {
	return msgF64(send(v, selWindow), selBackingScaleFactor)
}

func setNeedsDisplay(v uintptr) { send(v, selSetNeedsDisplay, 1) }

// isWindowVisible reports the occlusion state of the view's window when
// that window reports changes to us; any other view counts as visible.
func isWindowVisible(v uintptr) bool {
	win := send(v, selWindow)
	if win == 0 || associated(win, selGioInputView) != v || send(send(win, selDelegate), selIsKindOfClass, classGioWindowDelegate)&0xff == 0 {
		return true
	}
	return send(win, selOcclusionState)&nsWindowOcclusionStateVisible != 0
}

func makeKeyAndOrderFront(win uintptr)  { send(win, selMakeKeyAndOrderFront, 0) }
func makeFirstResponder(win, v uintptr) { send(win, selMakeFirstResponder, v) }
func toggleFullScreen(win uintptr)      { send(win, selToggleFullScreen, 0) }
func windowStyleMask(win uintptr) uint  { return uint(send(win, selStyleMask)) }

func setWindowStyleMask(win uintptr, m uint) { send(win, selSetStyleMask, uintptr(m)) }

func setWindowTitleVisibility(win uintptr, v int) {
	send(win, selSetTitleVisibility, uintptr(v))
}

func setWindowTitlebarAppearsTransparent(win uintptr, t bool) {
	send(win, selSetTitlebarTransp, boolArg(t))
}

func setWindowStandardButtonHidden(win uintptr, btn int, hide bool) {
	send(send(win, selStandardWindowButton, uintptr(btn)), selSetHidden, boolArg(hide))
}

func performWindowDragWithEvent(win, evt uintptr) { send(win, selPerformWindowDrag, evt) }
func closeWindow(win uintptr)                     { send(win, selPerformClose, 0) }

func setSize(win uintptr, w, h float64) {
	msgSetSize(win, selSetContentSize, nsPoint{X: w, Y: h})
}

func setMinSize(win uintptr, w, h float64) {
	msgSetSize(win, selSetContentMinSize, nsPoint{X: w, Y: h})
}

func setMaxSize(win uintptr, w, h float64) {
	msgSetSize(win, selSetContentMaxSize, nsPoint{X: w, Y: h})
	msgSetSize(win, selSetMaxFullScreenSize, nsPoint{X: w, Y: h})
}

func setScreenFrame(win uintptr, x, y, w, h float64) {
	msgSetRectBool(win, selSetFrameDisplay, nsRect{Origin: nsPoint{X: x, Y: y}, Size: nsPoint{X: w, Y: h}}, true)
}

func resetLayerFrame(v uintptr) { msgSetRect(send(v, selLayer), selSetFrame, msgRect(v, selFrame)) }
func hideWindow(win uintptr)    { send(win, selMiniaturize, win) }
func unhideWindow(win uintptr)  { send(win, selDeminiaturize, win) }

func screenFrame(win uintptr) nsRect { return msgRect(send(win, selScreen), selFrame) }

func setTitle(win uintptr, title string) {
	t := stringToNSString(title)
	defer releaseObject(t)
	send(win, selSetTitle, t)
}

func setWindowLevel(win uintptr, level int) { send(win, selSetLevel, uintptr(level)) }

func isWindowZoomed(win uintptr) bool       { return send(win, selIsZoomed)&0xff != 0 }
func isWindowMiniaturized(win uintptr) bool { return send(win, selIsMiniaturized)&0xff != 0 }
func zoomWindow(win uintptr)                { send(win, selZoom, 0) }
func layerForView(v uintptr) uintptr        { return send(v, selLayer) }

func windowForView(v uintptr) uintptr {
	// Material installation briefly detaches the input view. Driver actions
	// flushed by that callback must still reach their original window.
	if win := send(v, selWindow); win != 0 {
		return win
	}
	return associated(v, selGioHostWindow)
}

func raiseWindow(win uintptr) {
	// NSRunningApplication's active state can lag accessory-app transitions.
	// Raise is an explicit activation request, even when that snapshot is stale.
	send(send(classNSApplication, selSharedApplication), selActivate, 1)
	send(win, selMakeKeyAndOrderFront, 0)
}

// currentInputClient returns the current input context if v is its client.
func currentInputContext(v uintptr) uintptr {
	ctx := send(classNSTextInputContext, selCurrentInputContext)
	if ctx != 0 && send(ctx, selClient) == v {
		return ctx
	}
	return 0
}

func discardMarkedText(v uintptr) {
	if ctx := currentInputContext(v); ctx != 0 {
		send(ctx, selDiscardMarkedText)
	}
}

func invalidateCharacterCoordinates(v uintptr) {
	if ctx := currentInputContext(v); ctx != 0 {
		send(ctx, selInvalidateCharCoords)
	}
}

func interpretKeyEvents(v, evt uintptr) {
	send(v, selInterpretKeyEvents, send(classNSArray, selArrayWithObject, evt))
}

func cascadeTopLeftFromPoint(win uintptr, p nsPoint) nsPoint {
	return msgPointPoint(win, selCascade, p)
}
