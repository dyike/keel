// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && cgo && !purego

package app

// Gio's AppKit calls (with os_macos.m) behind the functions os_macos.go
// uses. os_macos_purego.go registers the same classes without cgo. Keel
// patch: the Objective-C is Gio's; the Go wrappers are new.

/*
#cgo CFLAGS: -Werror -Wno-deprecated-declarations -fobjc-arc -x objective-c
#cgo LDFLAGS: -framework AppKit -framework QuartzCore

#include <AppKit/AppKit.h>
#include <objc/runtime.h>

#define MOUSE_MOVE 1
#define MOUSE_UP 2
#define MOUSE_DOWN 3
#define MOUSE_SCROLL 4

__attribute__ ((visibility ("hidden"))) void gio_main(void);
__attribute__ ((visibility ("hidden"))) void gio_init(void);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createView(int presentWithTrans);
__attribute__ ((visibility ("hidden"))) CFTypeRef gio_createWindow(CFTypeRef viewRef, CGFloat width, CGFloat height);
__attribute__ ((visibility ("hidden"))) void gio_viewSetHandle(CFTypeRef viewRef, uintptr_t handle);

static void writeClipboard(CFTypeRef str) {
	@autoreleasepool {
		NSString *s = (__bridge NSString *)str;
		NSPasteboard *p = NSPasteboard.generalPasteboard;
		[p declareTypes:@[NSPasteboardTypeString] owner:nil];
		[p setString:s forType:NSPasteboardTypeString];
	}
}

static CFTypeRef readClipboard(void) {
	@autoreleasepool {
		NSPasteboard *p = NSPasteboard.generalPasteboard;
		NSString *content = [p stringForType:NSPasteboardTypeString];
		return (__bridge_retained CFTypeRef)content;
	}
}

static CGFloat viewHeight(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		return [view bounds].size.height;
	}
}

static CGFloat viewWidth(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		return [view bounds].size.width;
	}
}

static CGFloat getScreenBackingScale(void) {
	@autoreleasepool {
		return [NSScreen.mainScreen backingScaleFactor];
	}
}

static CGFloat getViewBackingScale(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		return [view.window backingScaleFactor];
	}
}

static void setNeedsDisplay(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		[view setNeedsDisplay:YES];
	}
}

static NSPoint cascadeTopLeftFromPoint(CFTypeRef windowRef, NSPoint topLeft) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		return [window cascadeTopLeftFromPoint:topLeft];
	}
}

// isWindowVisible reports the occlusion state of the view's window when
// that window reports changes to us; any other view counts as visible.
static int isWindowVisible(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		NSWindow *window = view.window;
		if (window == nil || objc_getAssociatedObject(window, @selector(gioInputView)) != view || ![window.delegate isKindOfClass:NSClassFromString(@"GioWindowDelegate")]) {
			return 1;
		}
		return (window.occlusionState & NSWindowOcclusionStateVisible) != 0;
	}
}

static void makeKeyAndOrderFront(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		[window makeKeyAndOrderFront:nil];
	}
}

static void makeFirstResponder(CFTypeRef windowRef, CFTypeRef viewRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		NSView *view = (__bridge NSView *)viewRef;
		[window makeFirstResponder:view];
	}
}

static void toggleFullScreen(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		[window toggleFullScreen:nil];
	}
}

static NSWindowStyleMask getWindowStyleMask(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		return [window styleMask];
	}
}

static void setWindowStyleMask(CFTypeRef windowRef, NSWindowStyleMask mask) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		window.styleMask = mask;
	}
}

static void setWindowTitleVisibility(CFTypeRef windowRef, NSWindowTitleVisibility state) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		window.titleVisibility = state;
	}
}

static void setWindowTitlebarAppearsTransparent(CFTypeRef windowRef, int transparent) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		window.titlebarAppearsTransparent = (BOOL)transparent;
	}
}

static void setWindowStandardButtonHidden(CFTypeRef windowRef, NSWindowButton btn, int hide) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		[window standardWindowButton:btn].hidden = (BOOL)hide;
	}
}

static void performWindowDragWithEvent(CFTypeRef windowRef, CFTypeRef evt) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		[window performWindowDragWithEvent:(__bridge NSEvent*)evt];
	}
}

static void closeWindow(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		[window performClose:nil];
	}
}

static void setSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		NSSize size = NSMakeSize(width, height);
		[window setContentSize:size];
	}
}

static void setMinSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		window.contentMinSize = NSMakeSize(width, height);
	}
}

static void setMaxSize(CFTypeRef windowRef, CGFloat width, CGFloat height) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		window.contentMaxSize = NSMakeSize(width, height);
		window.maxFullScreenContentSize = NSMakeSize(width, height);
	}
}

static void setScreenFrame(CFTypeRef windowRef, CGFloat x, CGFloat y, CGFloat w, CGFloat h) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		NSRect r = NSMakeRect(x, y, w, h);
		[window setFrame:r display:YES];
	}
}

static void resetLayerFrame(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView* view = (__bridge NSView *)viewRef;
		NSRect r = view.frame;
		view.layer.frame = r;
	}
}

static void hideWindow(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		[window miniaturize:window];
	}
}

static void unhideWindow(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		[window deminiaturize:window];
	}
}

static NSRect getScreenFrame(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow* window = (__bridge NSWindow *)windowRef;
		return [[window screen] frame];
	}
}

static void setTitle(CFTypeRef windowRef, CFTypeRef titleRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		window.title = (__bridge NSString *)titleRef;
	}
}

static void setWindowLevel(CFTypeRef windowRef, NSWindowLevel level) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		window.level = level;
	}
}

static int isWindowZoomed(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		return window.zoomed ? 1 : 0;
	}
}

static int isWindowMiniaturized(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		return window.miniaturized ? 1 : 0;
	}
}

static void zoomWindow(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		[window zoom:nil];
	}
}

static CFTypeRef layerForView(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		return (__bridge CFTypeRef)view.layer;
	}
}

static CFTypeRef windowForView(CFTypeRef viewRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		// Material installation briefly detaches the input view. Driver actions
		// flushed by that callback must still reach their original window.
		NSWindow *window = view.window;
		if (window == nil) {
			window = objc_getAssociatedObject(view, @selector(gioHostWindow));
		}
		return (__bridge CFTypeRef)window;
	}
}

static void raiseWindow(CFTypeRef windowRef) {
	@autoreleasepool {
		// NSRunningApplication's active state can lag accessory-app transitions.
		// Raise is an explicit activation request, even when that snapshot is stale.
		[NSApp activateIgnoringOtherApps:YES];
		NSWindow* window = (__bridge NSWindow *)windowRef;
		[window makeKeyAndOrderFront:nil];
	}
}

static CFTypeRef createInputContext(CFTypeRef clientRef) {
	@autoreleasepool {
		id<NSTextInputClient> client = (__bridge id<NSTextInputClient>)clientRef;
		NSTextInputContext *ctx = [[NSTextInputContext alloc] initWithClient:client];
		return CFBridgingRetain(ctx);
	}
}

static void discardMarkedText(CFTypeRef viewRef) {
	@autoreleasepool {
		id<NSTextInputClient> view = (__bridge id<NSTextInputClient>)viewRef;
		NSTextInputContext *ctx = [NSTextInputContext currentInputContext];
		if (view == [ctx client]) {
			[ctx discardMarkedText];
		}
	}
}

static void invalidateCharacterCoordinates(CFTypeRef viewRef) {
	@autoreleasepool {
		id<NSTextInputClient> view = (__bridge id<NSTextInputClient>)viewRef;
		NSTextInputContext *ctx = [NSTextInputContext currentInputContext];
		if (view == [ctx client]) {
			[ctx invalidateCharacterCoordinates];
		}
	}
}

static void interpretKeyEvents(CFTypeRef viewRef, CFTypeRef eventRef) {
	@autoreleasepool {
		NSView *view = (__bridge NSView *)viewRef;
		NSEvent *event = (__bridge NSEvent *)eventRef;
		[view interpretKeyEvents:[NSArray arrayWithObject:event]];
	}
}

static int isMiniaturized(CFTypeRef windowRef) {
	@autoreleasepool {
		NSWindow *window = (__bridge NSWindow *)windowRef;
		return window.miniaturized ? 1 : 0;
	}
}
*/
import "C"

import "runtime/cgo"

func ref(r uintptr) C.CFTypeRef { return C.CFTypeRef(r) }

func cBool(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

// The values os_macos.go spells out must match the SDK.
func init() {
	for _, c := range [...][2]uint64{
		{C.NSWindowStyleMaskTitled, nsWindowStyleMaskTitled},
		{C.NSWindowStyleMaskClosable, nsWindowStyleMaskClosable},
		{C.NSWindowStyleMaskMiniaturizable, nsWindowStyleMaskMiniaturizable},
		{C.NSWindowStyleMaskResizable, nsWindowStyleMaskResizable},
		{C.NSWindowStyleMaskFullScreen, nsWindowStyleMaskFullScreen},
		{C.NSWindowStyleMaskFullSizeContentView, nsWindowStyleMaskFullSizeContentView},
		{C.NSWindowTitleVisible, nsWindowTitleVisible},
		{C.NSWindowTitleHidden, nsWindowTitleHidden},
		{C.NSFloatingWindowLevel, nsFloatingWindowLevel},
		{C.NSWindowCloseButton, nsWindowCloseButton},
		{C.NSWindowMiniaturizeButton, nsWindowMiniaturizeButton},
		{C.NSWindowZoomButton, nsWindowZoomButton},
		{C.NSShiftKeyMask, nsShiftKeyMask},
		{C.NSControlKeyMask, nsControlKeyMask},
		{C.NSAlternateKeyMask, nsAlternateKeyMask},
		{C.NSCommandKeyMask, nsCommandKeyMask},
		{C.NSUpArrowFunctionKey, nsUpArrowFunctionKey},
		{C.NSDownArrowFunctionKey, nsDownArrowFunctionKey},
		{C.NSLeftArrowFunctionKey, nsLeftArrowFunctionKey},
		{C.NSRightArrowFunctionKey, nsRightArrowFunctionKey},
		{C.NSF1FunctionKey, nsF1FunctionKey},
		{C.NSF2FunctionKey, nsF2FunctionKey},
		{C.NSF3FunctionKey, nsF3FunctionKey},
		{C.NSF4FunctionKey, nsF4FunctionKey},
		{C.NSF5FunctionKey, nsF5FunctionKey},
		{C.NSF6FunctionKey, nsF6FunctionKey},
		{C.NSF7FunctionKey, nsF7FunctionKey},
		{C.NSF8FunctionKey, nsF8FunctionKey},
		{C.NSF9FunctionKey, nsF9FunctionKey},
		{C.NSF10FunctionKey, nsF10FunctionKey},
		{C.NSF11FunctionKey, nsF11FunctionKey},
		{C.NSF12FunctionKey, nsF12FunctionKey},
		{C.NSDeleteFunctionKey, nsDeleteFunctionKey},
		{C.NSHomeFunctionKey, nsHomeFunctionKey},
		{C.NSEndFunctionKey, nsEndFunctionKey},
		{C.NSPageUpFunctionKey, nsPageUpFunctionKey},
		{C.NSPageDownFunctionKey, nsPageDownFunctionKey},
		{C.NSNotFound, nsNotFound},
		{C.MOUSE_MOVE, mouseMove},
		{C.MOUSE_UP, mouseUp},
		{C.MOUSE_DOWN, mouseDown},
		{C.MOUSE_SCROLL, mouseScroll},
	} {
		if c[0] != c[1] {
			panic("app: AppKit value differs from the SDK")
		}
	}
}

func appInit() { C.gio_init() }
func appMain() { C.gio_main() }

func createView(presentWithTrans bool) uintptr {
	return uintptr(C.gio_createView(cBool(presentWithTrans)))
}

func createWindow(view uintptr, width, height float64) uintptr {
	return uintptr(C.gio_createWindow(ref(view), C.CGFloat(width), C.CGFloat(height)))
}

func viewSetWindow(view uintptr, w *window) {
	C.gio_viewSetHandle(ref(view), C.uintptr_t(cgo.NewHandle(w)))
}

func releaseObject(obj uintptr) { C.CFRelease(ref(obj)) }

func readClipboard() string {
	cstr := C.readClipboard()
	if cstr != 0 {
		defer C.CFRelease(cstr)
	}
	return nsstringToString(cstr)
}

func writeClipboard(s string) {
	cstr := stringToNSString(s)
	defer C.CFRelease(cstr)
	C.writeClipboard(cstr)
}

func viewHeight(v uintptr) float64       { return float64(C.viewHeight(ref(v))) }
func viewWidth(v uintptr) float64        { return float64(C.viewWidth(ref(v))) }
func screenBackingScale() float64        { return float64(C.getScreenBackingScale()) }
func viewBackingScale(v uintptr) float64 { return float64(C.getViewBackingScale(ref(v))) }
func setNeedsDisplay(v uintptr)          { C.setNeedsDisplay(ref(v)) }
func isWindowVisible(v uintptr) bool     { return C.isWindowVisible(ref(v)) != 0 }
func makeKeyAndOrderFront(win uintptr)   { C.makeKeyAndOrderFront(ref(win)) }
func makeFirstResponder(win, v uintptr)  { C.makeFirstResponder(ref(win), ref(v)) }
func toggleFullScreen(win uintptr)       { C.toggleFullScreen(ref(win)) }
func windowStyleMask(win uintptr) uint   { return uint(C.getWindowStyleMask(ref(win))) }
func setWindowStyleMask(win uintptr, m uint) {
	C.setWindowStyleMask(ref(win), C.NSWindowStyleMask(m))
}
func setWindowTitleVisibility(win uintptr, v int) {
	C.setWindowTitleVisibility(ref(win), C.NSWindowTitleVisibility(v))
}
func setWindowTitlebarAppearsTransparent(win uintptr, t bool) {
	C.setWindowTitlebarAppearsTransparent(ref(win), cBool(t))
}
func setWindowStandardButtonHidden(win uintptr, btn int, hide bool) {
	C.setWindowStandardButtonHidden(ref(win), C.NSWindowButton(btn), cBool(hide))
}
func performWindowDragWithEvent(win, evt uintptr) {
	C.performWindowDragWithEvent(ref(win), ref(evt))
}
func closeWindow(win uintptr) { C.closeWindow(ref(win)) }
func setSize(win uintptr, w, h float64) {
	C.setSize(ref(win), C.CGFloat(w), C.CGFloat(h))
}
func setMinSize(win uintptr, w, h float64) {
	C.setMinSize(ref(win), C.CGFloat(w), C.CGFloat(h))
}
func setMaxSize(win uintptr, w, h float64) {
	C.setMaxSize(ref(win), C.CGFloat(w), C.CGFloat(h))
}
func setScreenFrame(win uintptr, x, y, w, h float64) {
	C.setScreenFrame(ref(win), C.CGFloat(x), C.CGFloat(y), C.CGFloat(w), C.CGFloat(h))
}
func resetLayerFrame(v uintptr) { C.resetLayerFrame(ref(v)) }
func hideWindow(win uintptr)    { C.hideWindow(ref(win)) }
func unhideWindow(win uintptr)  { C.unhideWindow(ref(win)) }
func screenFrame(win uintptr) nsRect {
	r := C.getScreenFrame(ref(win))
	return nsRect{
		Origin: nsPoint{X: float64(r.origin.x), Y: float64(r.origin.y)},
		Size:   nsPoint{X: float64(r.size.width), Y: float64(r.size.height)},
	}
}
func setTitle(win uintptr, title string) {
	t := stringToNSString(title)
	defer C.CFRelease(t)
	C.setTitle(ref(win), t)
}
func setWindowLevel(win uintptr, level int) { C.setWindowLevel(ref(win), C.NSWindowLevel(level)) }
func isWindowZoomed(win uintptr) bool       { return C.isWindowZoomed(ref(win)) != 0 }
func isWindowMiniaturized(win uintptr) bool { return C.isWindowMiniaturized(ref(win)) != 0 }
func zoomWindow(win uintptr)                { C.zoomWindow(ref(win)) }
func layerForView(v uintptr) uintptr        { return uintptr(C.layerForView(ref(v))) }
func windowForView(v uintptr) uintptr       { return uintptr(C.windowForView(ref(v))) }
func raiseWindow(win uintptr)               { C.raiseWindow(ref(win)) }
func discardMarkedText(v uintptr)           { C.discardMarkedText(ref(v)) }
func invalidateCharacterCoordinates(v uintptr) {
	C.invalidateCharacterCoordinates(ref(v))
}
func interpretKeyEvents(v, evt uintptr) { C.interpretKeyEvents(ref(v), ref(evt)) }

func cascadeTopLeftFromPoint(win uintptr, p nsPoint) nsPoint {
	r := C.cascadeTopLeftFromPoint(ref(win), C.NSPoint{x: C.CGFloat(p.X), y: C.CGFloat(p.Y)})
	return nsPoint{X: float64(r.x), Y: float64(r.y)}
}

func goRange(r C.NSRange) nsRange { return nsRange{Location: uint(r.location), Length: uint(r.length)} }
func cRange(r nsRange) C.NSRange {
	return C.NSRange{location: C.NSUInteger(r.Location), length: C.NSUInteger(r.Length)}
}

func windowFor(h C.uintptr_t) *window {
	return cgo.Handle(h).Value().(*window)
}

//export gio_onKeys
func gio_onKeys(h C.uintptr_t, event C.CFTypeRef, cstr C.CFTypeRef, ti C.double, mods C.NSUInteger, keyDown C.bool) {
	windowFor(h).onKeys(uintptr(event), nsstringToString(cstr), uint(mods), bool(keyDown))
}

//export gio_onCommandBySelector
func gio_onCommandBySelector(h C.uintptr_t) C.bool {
	return C.bool(windowFor(h).onCommandBySelector())
}

//export gio_onFlagsChanged
func gio_onFlagsChanged(h C.uintptr_t, curMods C.NSUInteger) {
	windowFor(h).onFlagsChanged(uint(curMods))
}

//export gio_onText
func gio_onText(h C.uintptr_t, cstr C.CFTypeRef) {
	windowFor(h).onText(nsstringToString(cstr))
}

//export gio_onMouse
func gio_onMouse(h C.uintptr_t, evt C.CFTypeRef, cdir C.int, cbtn C.NSInteger, x, y, dx, dy C.CGFloat, ti C.double, mods C.NSUInteger) {
	windowFor(h).onMouse(uintptr(evt), int(cdir), int(cbtn), float64(x), float64(y), float64(dx), float64(dy), float64(ti), uint(mods))
}

//export gio_onDraw
func gio_onDraw(h C.uintptr_t) {
	windowFor(h).draw()
}

//export gio_onOcclusion
func gio_onOcclusion(h C.uintptr_t, visible C.int) {
	windowFor(h).onOcclusion(visible != 0)
}

//export gio_onFocus
func gio_onFocus(h C.uintptr_t, focus C.int) {
	windowFor(h).onFocus(focus == 1)
}

//export gio_onChangeScreen
func gio_onChangeScreen(h C.uintptr_t, did uint64) {
	windowFor(h).onChangeScreen(did)
}

//export gio_hasMarkedText
func gio_hasMarkedText(h C.uintptr_t) C.int {
	return cBool(windowFor(h).hasMarkedText())
}

//export gio_markedRange
func gio_markedRange(h C.uintptr_t) C.NSRange {
	return cRange(windowFor(h).markedRange())
}

//export gio_selectedRange
func gio_selectedRange(h C.uintptr_t) C.NSRange {
	return cRange(windowFor(h).selectedRange())
}

//export gio_unmarkText
func gio_unmarkText(h C.uintptr_t) {
	windowFor(h).unmarkText()
}

//export gio_setMarkedText
func gio_setMarkedText(h C.uintptr_t, cstr C.CFTypeRef, selRange C.NSRange, replaceRange C.NSRange) {
	windowFor(h).setMarkedText(nsstringToString(cstr), goRange(selRange), goRange(replaceRange))
}

//export gio_substringForProposedRange
func gio_substringForProposedRange(h C.uintptr_t, crng C.NSRange, actual C.NSRangePointer) C.CFTypeRef {
	text, rng := windowFor(h).substringForProposedRange(goRange(crng))
	*actual = cRange(rng)
	return stringToNSString(text)
}

//export gio_insertText
func gio_insertText(h C.uintptr_t, cstr C.CFTypeRef, crng C.NSRange) {
	windowFor(h).insertText(nsstringToString(cstr), goRange(crng))
}

//export gio_characterIndexForPoint
func gio_characterIndexForPoint(h C.uintptr_t, p C.NSPoint) C.NSUInteger {
	return C.NSNotFound
}

//export gio_firstRectForCharacterRange
func gio_firstRectForCharacterRange(h C.uintptr_t, crng C.NSRange, actual C.NSRangePointer) C.NSRect {
	r, rng := windowFor(h).firstRectForCharacterRange(goRange(crng))
	*actual = cRange(rng)
	return C.NSMakeRect(C.CGFloat(r.Origin.X), C.CGFloat(r.Origin.Y), C.CGFloat(r.Size.X), C.CGFloat(r.Size.Y))
}

//export gio_onAttached
func gio_onAttached(h C.uintptr_t, attached C.int) {
	windowFor(h).onAttached(attached != 0)
}

//export gio_onDestroy
func gio_onDestroy(h C.uintptr_t) {
	windowFor(h).onDestroy()
	cgo.Handle(h).Delete()
}

//export gio_onCloseRequest
func gio_onCloseRequest(h C.uintptr_t) C.int {
	return cBool(windowFor(h).onCloseRequest())
}

//export gio_onFinishLaunching
func gio_onFinishLaunching() {
	onFinishLaunching()
}

//export gio_onOpenURI
func gio_onOpenURI(uri C.CFTypeRef) {
	onOpenURI(nsstringToString(uri))
}
