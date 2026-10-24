// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios

package app

import (
	"errors"
	"image"
	"io"
	"runtime"
	"strings"
	"structs"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dyike/keel/gio/internal/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/io/transfer"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"

	_ "github.com/dyike/keel/gio/internal/cocoainit"
)

// The AppKit calls live in os_macos_cgo.go (Gio's Objective-C, with
// os_macos.m) and os_macos_purego.go (the same classes registered at run
// time). Keel patch: this file holds the logic both share, over plain
// handles; the native callbacks call the window methods named on*.

// AppKit values; os_macos_cgo.go checks them against the SDK.
const (
	nsWindowStyleMaskTitled              = 1 << 0
	nsWindowStyleMaskClosable            = 1 << 1
	nsWindowStyleMaskMiniaturizable      = 1 << 2
	nsWindowStyleMaskResizable           = 1 << 3
	nsWindowStyleMaskFullScreen          = 1 << 14
	nsWindowStyleMaskFullSizeContentView = 1 << 15

	nsWindowTitleVisible = 0
	nsWindowTitleHidden  = 1

	nsFloatingWindowLevel = 3

	nsWindowCloseButton       = 0
	nsWindowMiniaturizeButton = 1
	nsWindowZoomButton        = 2

	nsShiftKeyMask     = 1 << 17
	nsControlKeyMask   = 1 << 18
	nsAlternateKeyMask = 1 << 19
	nsCommandKeyMask   = 1 << 20

	nsUpArrowFunctionKey    = 0xF700
	nsDownArrowFunctionKey  = 0xF701
	nsLeftArrowFunctionKey  = 0xF702
	nsRightArrowFunctionKey = 0xF703
	nsF1FunctionKey         = 0xF704
	nsF2FunctionKey         = 0xF705
	nsF3FunctionKey         = 0xF706
	nsF4FunctionKey         = 0xF707
	nsF5FunctionKey         = 0xF708
	nsF6FunctionKey         = 0xF709
	nsF7FunctionKey         = 0xF70A
	nsF8FunctionKey         = 0xF70B
	nsF9FunctionKey         = 0xF70C
	nsF10FunctionKey        = 0xF70D
	nsF11FunctionKey        = 0xF70E
	nsF12FunctionKey        = 0xF70F
	nsDeleteFunctionKey     = 0xF728
	nsHomeFunctionKey       = 0xF729
	nsEndFunctionKey        = 0xF72B
	nsPageUpFunctionKey     = 0xF72C
	nsPageDownFunctionKey   = 0xF72D

	nsNotFound = 1<<63 - 1

	mouseMove   = 1
	mouseUp     = 2
	mouseDown   = 3
	mouseScroll = 4
)

type nsPoint struct {
	_    structs.HostLayout
	X, Y float64
}

type nsRect struct {
	_            structs.HostLayout
	Origin, Size nsPoint
}

type nsRange struct {
	_                structs.HostLayout
	Location, Length uint
}

func init() {
	// Darwin requires that UI operations happen on the main thread only.
	runtime.LockOSThread()
	// Register launch finished listener.
	appInit()
}

// AppKitViewEvent notifies the client of changes to the window AppKit handles.
// The handles are retained until another AppKitViewEvent is sent.
type AppKitViewEvent struct {
	// View is a CFTypeRef for the NSView for the window.
	View uintptr
	// Layer is a CFTypeRef of the CALayer of View.
	Layer uintptr
}

type window struct {
	view        uintptr
	w           *callbacks
	anim        bool
	displayLink *displayLink
	// occluded is set while nothing of the window shows on screen, and drawn
	// once its first frame is (Keel patch, see gio_onOcclusion).
	occluded, drawn bool
	// redraw is a single entry channel for making sure only one
	// display link redraw request is in flight.
	redraw      chan struct{}
	cursor      pointer.Cursor
	pointerBtns pointer.Buttons
	loop        *eventLoop
	lastMods    uint

	scale  float32
	config Config

	keysDown map[key.Name]struct{}
	// cmdKeys is for storing the current key event while
	// waiting for a doCommandBySelector.
	cmdKeys cmdKeys
}

type cmdKeys struct {
	eventStr  string
	eventMods key.Modifiers
}

// launched is closed when applicationDidFinishLaunching is called.
var launched = make(chan struct{})

// nextTopLeft is the offset to use for the next window's call to
// cascadeTopLeftFromPoint.
var nextTopLeft nsPoint

func (w *window) contextView() uintptr {
	return w.view
}

func (w *window) ReadClipboard() {
	content := readClipboard()
	w.ProcessEvent(transfer.DataEvent{
		Type: "application/text",
		Open: func() io.ReadCloser {
			return io.NopCloser(strings.NewReader(content))
		},
	})
}

func (w *window) WriteClipboard(mime string, s []byte) {
	writeClipboard(string(s))
}

func (w *window) updateWindowMode() {
	w.scale = float32(viewBackingScale(w.view))
	wf, hf := float32(viewWidth(w.view)), float32(viewHeight(w.view))
	w.config.Size = image.Point{
		X: int(wf*w.scale + .5),
		Y: int(hf*w.scale + .5),
	}
	w.config.Mode = Windowed
	window := windowForView(w.view)
	if window == 0 {
		return
	}
	style := windowStyleMask(windowForView(w.view))
	switch {
	case style&nsWindowStyleMaskFullScreen != 0:
		w.config.Mode = Fullscreen
	case isWindowZoomed(window):
		w.config.Mode = Maximized
	}
	w.config.Decorated = style&nsWindowStyleMaskFullSizeContentView == 0
}

func (w *window) Configure(options []Option) {
	screenScale := float32(screenBackingScale())
	cfg := configFor(screenScale)
	cnf := w.config
	cnf.apply(cfg, options)
	window := windowForView(w.view)

	mask := windowStyleMask(window)
	fullscreen := mask&nsWindowStyleMaskFullScreen != 0
	switch cnf.Mode {
	case Fullscreen:
		if isWindowMiniaturized(window) {
			unhideWindow(window)
		}
		if !fullscreen {
			toggleFullScreen(window)
		}
	case Minimized:
		hideWindow(window)
	case Maximized:
		if isWindowMiniaturized(window) {
			unhideWindow(window)
		}
		if fullscreen {
			toggleFullScreen(window)
		}
		w.setTitle(cnf.Title)
		if !isWindowZoomed(window) {
			zoomWindow(window)
		}
	case Windowed:
		if isWindowMiniaturized(window) {
			unhideWindow(window)
		}
		if fullscreen {
			toggleFullScreen(window)
		}
		w.setTitle(cnf.Title)
		w.config.Size = cnf.Size
		cnf.Size = cnf.Size.Div(int(screenScale))
		setSize(window, float64(cnf.Size.X), float64(cnf.Size.Y))
		w.config.MinSize = cnf.MinSize
		cnf.MinSize = cnf.MinSize.Div(int(screenScale))
		setMinSize(window, float64(cnf.MinSize.X), float64(cnf.MinSize.Y))
		w.config.MaxSize = cnf.MaxSize
		cnf.MaxSize = cnf.MaxSize.Div(int(screenScale))
		if cnf.MaxSize != (image.Point{}) {
			setMaxSize(window, float64(cnf.MaxSize.X), float64(cnf.MaxSize.Y))
		}
		if isWindowZoomed(window) {
			zoomWindow(window)
		}
	}
	style := uint(nsWindowStyleMaskTitled | nsWindowStyleMaskResizable | nsWindowStyleMaskMiniaturizable | nsWindowStyleMaskClosable)
	style = nsWindowStyleMaskFullSizeContentView
	mask &^= style
	barTrans := false
	titleVis := nsWindowTitleVisible
	if !cnf.Decorated {
		mask |= style
		barTrans = true
		titleVis = nsWindowTitleHidden
	}
	if cnf.TopMost {
		setWindowLevel(window, nsFloatingWindowLevel)
	}
	setWindowTitlebarAppearsTransparent(window, barTrans)
	setWindowTitleVisibility(window, titleVis)
	setWindowStyleMask(window, mask)
	setWindowStandardButtonHidden(window, nsWindowCloseButton, barTrans)
	setWindowStandardButtonHidden(window, nsWindowMiniaturizeButton, barTrans)
	setWindowStandardButtonHidden(window, nsWindowZoomButton, barTrans)
	// When toggling the titlebar, the layer doesn't update its frame
	// until the next resize. Force it.
	resetLayerFrame(w.view)
}

func (w *window) setTitle(title string) {
	w.config.Title = title
	setTitle(windowForView(w.view), title)
}

func (w *window) Perform(acts system.Action) {
	window := windowForView(w.view)
	walkActions(acts, func(a system.Action) {
		switch a {
		case system.ActionCenter:
			r := screenFrame(window) // the screen size of the window
			screenScale := float32(screenBackingScale())
			sz := w.config.Size.Div(int(screenScale))
			x := (int(r.Size.X) - sz.X) / 2
			y := (int(r.Size.Y) - sz.Y) / 2
			setScreenFrame(window, float64(x), float64(y), float64(sz.X), float64(sz.Y))
		case system.ActionRaise:
			raiseWindow(window)
		}
	})
	if acts&system.ActionClose != 0 {
		closeWindow(window)
	}
}

func (w *window) SetCursor(cursor pointer.Cursor) {
	w.cursor = windowSetCursor(w.cursor, cursor)
}

func (w *window) EditorStateChanged(old, new editorState) {
	if shouldCancelComposition(old, new) {
		discardMarkedText(w.view)
		w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
	}
	if old.Selection.Caret != new.Selection.Caret || old.Selection.Transform != new.Selection.Transform {
		invalidateCharacterCoordinates(w.view)
	}
}

func (w *window) ShowTextInput(show bool) {}

func (w *window) SetInputHint(_ key.InputHint) {}

func (w *window) SetAnimating(anim bool) {
	w.anim = anim
	window := windowForView(w.view)
	if w.anim && !w.occluded && window != 0 && !isWindowMiniaturized(window) {
		w.displayLink.Start()
	} else {
		w.displayLink.Stop()
	}
}

func (w *window) runOnMain(f func()) {
	runOnMain(func() {
		// Make sure the view is still valid. The window might've been closed
		// during the switch to the main thread.
		if w.view != 0 {
			f()
		}
	})
}

func (w *window) onKeys(event uintptr, str string, mods uint, keyDown bool) {
	if w.keysDown == nil {
		w.keysDown = make(map[key.Name]struct{})
	}
	kmods := convertMods(mods)
	ks := key.Release
	if keyDown {
		ks = key.Press
		w.cmdKeys.eventStr = str
		w.cmdKeys.eventMods = kmods
		interpretKeyEvents(w.view, event)
	}
	for _, k := range str {
		if n, ok := convertKey(k); ok {
			ke := key.Event{
				Name:      n,
				Modifiers: kmods,
				State:     ks,
			}
			if keyDown {
				w.keysDown[ke.Name] = struct{}{}
				if _, isCmd := convertCommandKey(k); isCmd || kmods.Contain(key.ModCommand) {
					// doCommandBySelector already processed the event.
					return
				}
			} else {
				if _, pressed := w.keysDown[n]; !pressed {
					continue
				}
				delete(w.keysDown, n)
			}
			w.ProcessEvent(ke)
		}
	}
}

func (w *window) onCommandBySelector() bool {
	ev := w.cmdKeys
	w.cmdKeys = cmdKeys{}
	handled := false
	for _, k := range ev.eventStr {
		n, ok := convertCommandKey(k)
		if !ok && ev.eventMods.Contain(key.ModCommand) {
			n, ok = convertKey(k)
		}
		if !ok {
			continue
		}
		ke := key.Event{
			Name:      n,
			Modifiers: ev.eventMods,
			State:     key.Press,
		}
		handled = w.processEvent(ke) || handled
	}
	return handled
}

func (w *window) onFlagsChanged(curMods uint) {
	mods := []uint{nsControlKeyMask, nsAlternateKeyMask, nsShiftKeyMask, nsCommandKeyMask}
	keys := []key.Name{key.NameCtrl, key.NameAlt, key.NameShift, key.NameCommand}

	for i, mod := range mods {
		wasPressed := w.lastMods&mod != 0
		isPressed := curMods&mod != 0

		if wasPressed != isPressed {
			st := key.Release
			if isPressed {
				st = key.Press
			}
			w.ProcessEvent(key.Event{
				Name:  keys[i],
				State: st,
			})
		}
	}

	w.lastMods = curMods
}

func (w *window) onText(str string) {
	w.w.EditorInsert(str)
}

func (w *window) onMouse(evt uintptr, cdir int, cbtn int, x, y, dx, dy float64, ti float64, mods uint) {
	t := time.Duration(float64(ti)*float64(time.Second) + .5)
	xf, yf := float32(x)*w.scale, float32(y)*w.scale
	dxf, dyf := float32(dx)*w.scale, float32(dy)*w.scale
	pos := f32.Point{X: xf, Y: yf}
	var btn pointer.Buttons
	switch cbtn {
	case 0:
		btn = pointer.ButtonPrimary
	case 1:
		btn = pointer.ButtonSecondary
	case 2:
		btn = pointer.ButtonTertiary
	}
	var typ pointer.Kind
	switch cdir {
	case mouseMove:
		typ = pointer.Move
	case mouseUp:
		typ = pointer.Release
		w.pointerBtns &^= btn
	case mouseDown:
		typ = pointer.Press
		w.pointerBtns |= btn
		act, ok := w.w.ActionAt(pos)
		if ok && w.config.Mode != Fullscreen {
			switch act {
			case system.ActionMove:
				performWindowDragWithEvent(windowForView(w.view), evt)
				return
			}
		}
	case mouseScroll:
		typ = pointer.Scroll
	default:
		panic("invalid direction")
	}
	w.ProcessEvent(pointer.Event{
		Kind:      typ,
		Source:    pointer.Mouse,
		Time:      t,
		Buttons:   w.pointerBtns,
		Position:  pos,
		Scroll:    f32.Point{X: dxf, Y: dyf},
		Modifiers: convertMods(mods),
	})
}

// onOcclusion runs when the window starts or stops showing on screen
// (Keel patch).
func (w *window) onOcclusion(visible bool) {
	w.occluded = !visible
	if w.occluded {
		w.displayLink.Stop()
		return
	}
	// Catch up: animations resume and the window shows its current state.
	w.SetAnimating(w.anim)
	setNeedsDisplay(w.view)
}

func (w *window) onFocus(focus bool) {
	w.SetCursor(w.cursor)
	w.config.Focused = focus
	w.ProcessEvent(ConfigEvent{Config: w.config})
}

func (w *window) onChangeScreen(did uint64) {
	w.displayLink.SetDisplayID(did)
	setNeedsDisplay(w.view)
}

func (w *window) hasMarkedText() bool {
	state := w.w.EditorState()
	return state.compose.Start != -1
}

func (w *window) markedRange() nsRange {
	state := w.w.EditorState()
	rng := state.compose
	start, end := rng.Start, rng.End
	if start == -1 {
		return nsRange{Location: nsNotFound}
	}
	u16start := state.UTF16Index(start)
	return nsRange{
		Location: uint(u16start),
		Length:   uint(state.UTF16Index(end) - u16start),
	}
}

func (w *window) selectedRange() nsRange {
	state := w.w.EditorState()
	rng := state.Selection
	start, end := rng.Start, rng.End
	if start > end {
		start, end = end, start
	}
	u16start := state.UTF16Index(start)
	return nsRange{
		Location: uint(u16start),
		Length:   uint(state.UTF16Index(end) - u16start),
	}
}

func (w *window) unmarkText() {
	w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
}

func (w *window) setMarkedText(str string, selRange, replaceRange nsRange) {
	state := w.w.EditorState()
	rng := state.compose
	if rng.Start == -1 {
		rng = state.Selection.Range
	}
	if replaceRange.Location != nsNotFound {
		// replaceRange is relative to marked (or selected) text.
		offset := state.UTF16Index(rng.Start)
		start := state.RunesIndex(int(replaceRange.Location) + offset)
		end := state.RunesIndex(int(replaceRange.Location+replaceRange.Length) + offset)
		rng = key.Range{
			Start: start,
			End:   end,
		}
	}
	w.w.EditorReplace(rng, str)
	comp := key.Range{
		Start: rng.Start,
		End:   rng.Start + utf8.RuneCountInString(str),
	}
	w.w.SetComposingRegion(comp)

	sel := key.Range{Start: comp.End, End: comp.End}
	if selRange.Location != nsNotFound {
		// selRange is relative to inserted text.
		offset := state.UTF16Index(rng.Start)
		start := state.RunesIndex(int(selRange.Location) + offset)
		end := state.RunesIndex(int(selRange.Location+selRange.Length) + offset)
		sel = key.Range{
			Start: start,
			End:   end,
		}
	}
	w.w.SetEditorSelection(sel)
}

// substringForProposedRange returns the editor snippet and its range.
func (w *window) substringForProposedRange(crng nsRange) (string, nsRange) {
	state := w.w.EditorState()
	start, end := state.Snippet.Start, state.Snippet.End
	if start > end {
		start, end = end, start
	}
	rng := key.Range{
		Start: state.RunesIndex(int(crng.Location)),
		End:   state.RunesIndex(int(crng.Location + crng.Length)),
	}
	if rng.Start < start || end < rng.End {
		w.w.SetEditorSnippet(rng)
	}
	u16start := state.UTF16Index(start)
	actual := nsRange{Location: uint(u16start), Length: uint(state.UTF16Index(end) - u16start)}
	return state.Snippet.Text, actual
}

func (w *window) insertText(str string, crng nsRange) {
	// macOS IME in some cases calls insertText for command keys such as backspace
	// instead of doCommandBySelector.
	for _, r := range str {
		if _, ok := convertCommandKey(r); ok {
			w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
			return
		}
	}
	state := w.w.EditorState()
	rng := state.compose
	if rng.Start == -1 {
		rng = state.Selection.Range
	}
	if crng.Location != nsNotFound {
		rng = key.Range{
			Start: state.RunesIndex(int(crng.Location)),
			End:   state.RunesIndex(int(crng.Location + crng.Length)),
		}
	}
	w.w.EditorReplace(rng, str)
	w.w.SetComposingRegion(key.Range{Start: -1, End: -1})
	start := rng.Start
	if rng.End < start {
		start = rng.End
	}
	pos := start + utf8.RuneCountInString(str)
	w.w.SetEditorSelection(key.Range{Start: pos, End: pos})
}

// firstRectForCharacterRange returns the caret in view coordinates, and
// the range it covers.
func (w *window) firstRectForCharacterRange(crng nsRange) (nsRect, nsRange) {
	state := w.w.EditorState()
	sel := state.Selection
	u16start := state.UTF16Index(sel.Start)
	actual := nsRange{Location: uint(u16start)}
	// Transform to NSView local coordinates (lower left origin, undo backing scale).
	scale := 1. / float32(viewBackingScale(w.view))
	height := float32(viewHeight(w.view))
	local := f32.AffineId().Scale(f32.Pt(0, 0), f32.Pt(scale, -scale)).Offset(f32.Pt(0, height))
	t := local.Mul(sel.Transform)
	bounds := f32.Rectangle{
		Min: t.Transform(sel.Pos.Sub(f32.Pt(0, sel.Ascent))),
		Max: t.Transform(sel.Pos.Add(f32.Pt(0, sel.Descent))),
	}.Canon()
	sz := bounds.Size()
	return nsRect{
		Origin: nsPoint{X: float64(bounds.Min.X), Y: float64(bounds.Min.Y)},
		Size:   nsPoint{X: float64(sz.X), Y: float64(sz.Y)},
	}, actual
}

func (w *window) draw() {
	select {
	case <-w.redraw:
	default:
	}
	// Keel patch: a window nobody can see draws nothing after its first
	// frame (which window code may wait for), as AppKit apps do; it
	// redraws when it shows again. Gio drew every animation frame of
	// covered windows and of every window while the screen was locked.
	if w.occluded && w.drawn {
		return
	}
	w.drawn = true
	cnf := w.config
	w.updateWindowMode()
	if w.config != cnf {
		w.ProcessEvent(ConfigEvent{Config: w.config})
	}
	if w.anim {
		w.SetAnimating(w.anim)
	}
	sz := w.config.Size
	if sz.X == 0 || sz.Y == 0 {
		return
	}
	cfg := configFor(w.scale)
	w.ProcessEvent(frameEvent{
		FrameEvent: FrameEvent{
			Now:    time.Now(),
			Size:   sz,
			Metric: cfg,
		},
		Sync: true,
	})
}

func (w *window) ProcessEvent(e event.Event) {
	w.processEvent(e)
}

func (w *window) processEvent(e event.Event) bool {
	handled := w.w.ProcessEvent(e)
	w.loop.FlushEvents()
	return handled
}

func (w *window) Event() event.Event {
	return w.loop.Event()
}

func (w *window) Invalidate() {
	w.loop.Invalidate()
}

func (w *window) Run(f func()) {
	w.loop.Run(f)
}

func (w *window) Frame(frame *op.Ops) {
	w.loop.Frame(frame)
}

func configFor(scale float32) unit.Metric {
	return unit.Metric{
		PxPerDp: scale,
		PxPerSp: scale,
	}
}

func (w *window) onAttached(attached bool) {
	if attached {
		layer := layerForView(w.view)
		// Keel patch: AppKit reports occlusion changes only, and a window
		// created while the screen is locked stays occluded from the start.
		w.occluded = !isWindowVisible(w.view)
		w.ProcessEvent(AppKitViewEvent{View: uintptr(w.view), Layer: uintptr(layer)})
	} else {
		w.ProcessEvent(AppKitViewEvent{})
		w.SetAnimating(w.anim)
	}
}

// onDestroy runs as the view deallocates; the binding then forgets the
// view's window.
func (w *window) onDestroy() {
	w.ProcessEvent(DestroyEvent{})
	w.displayLink.Close()
	w.displayLink = nil
	w.view = 0
}

func (w *window) onCloseRequest() bool {
	e := &ClosingEvent{}
	w.ProcessEvent(e)
	return !e.abort
}

func onFinishLaunching() {
	close(launched)
}

func onOpenURI(uri string) {
	evt, err := newURLEvent(uri)
	if err != nil {
		return
	}

	processGlobalEvent(evt)
}

func newWindow(win *callbacks, options []Option) {
	<-launched
	res := make(chan struct{})
	runOnMain(func() {
		w := &window{
			redraw: make(chan struct{}, 1),
			w:      win,
		}
		w.loop = newEventLoop(w.w, w.wakeup)
		win.SetDriver(w)
		res <- struct{}{}
		var cnf Config
		cnf.apply(unit.Metric{}, options)
		if err := w.init(cnf.CustomRenderer); err != nil {
			w.ProcessEvent(DestroyEvent{Err: err})
			return
		}
		window := createWindow(w.view, float64(cnf.Size.X), float64(cnf.Size.Y))
		// Release our reference now that the NSWindow has it.
		releaseObject(w.view)
		w.Configure(options)
		if nextTopLeft.X == 0 && nextTopLeft.Y == 0 {
			// cascadeTopLeftFromPoint treats (0, 0) as a no-op,
			// and just returns the offset we need for the first window.
			nextTopLeft = cascadeTopLeftFromPoint(window, nextTopLeft)
		}
		nextTopLeft = cascadeTopLeftFromPoint(window, nextTopLeft)
		makeFirstResponder(window, w.view)
		// makeKeyAndOrderFront assumes ownership of our window reference.
		makeKeyAndOrderFront(window)
	})
	<-res
}

func (w *window) init(customRenderer bool) error {
	presentWithTrans := 1
	if customRenderer {
		presentWithTrans = 0
	}
	view := createView(presentWithTrans != 0)
	if view == 0 {
		return errors.New("newOSWindow: failed to create view")
	}
	scale := float32(viewBackingScale(view))
	w.scale = scale
	dl, err := newDisplayLink(func() {
		select {
		case w.redraw <- struct{}{}:
		default:
			return
		}
		w.runOnMain(func() {
			setNeedsDisplay(w.view)
		})
	})
	w.displayLink = dl
	if err != nil {
		releaseObject(view)
		return err
	}
	viewSetWindow(view, w)
	w.view = view
	return nil
}

func osMain() {
	if !isMainThread() {
		panic("app.Main must run on the main goroutine")
	}
	appMain()
}

func convertCommandKey(k rune) (key.Name, bool) {
	var n key.Name
	switch k {
	case '\x1b': // ASCII escape.
		n = key.NameEscape
	case nsLeftArrowFunctionKey:
		n = key.NameLeftArrow
	case nsRightArrowFunctionKey:
		n = key.NameRightArrow
	case nsUpArrowFunctionKey:
		n = key.NameUpArrow
	case nsDownArrowFunctionKey:
		n = key.NameDownArrow
	case '\r':
		n = key.NameReturn
	case '\x03':
		n = key.NameEnter
	case nsHomeFunctionKey:
		n = key.NameHome
	case nsEndFunctionKey:
		n = key.NameEnd
	case '\x7f', '\b':
		n = key.NameDeleteBackward
	case nsDeleteFunctionKey:
		n = key.NameDeleteForward
	case '\t', 0x19:
		n = key.NameTab
	case nsPageUpFunctionKey:
		n = key.NamePageUp
	case nsPageDownFunctionKey:
		n = key.NamePageDown
	default:
		return "", false
	}
	return n, true
}

func convertKey(k rune) (key.Name, bool) {
	if n, ok := convertCommandKey(k); ok {
		return n, true
	}
	var n key.Name
	switch k {
	case nsF1FunctionKey:
		n = key.NameF1
	case nsF2FunctionKey:
		n = key.NameF2
	case nsF3FunctionKey:
		n = key.NameF3
	case nsF4FunctionKey:
		n = key.NameF4
	case nsF5FunctionKey:
		n = key.NameF5
	case nsF6FunctionKey:
		n = key.NameF6
	case nsF7FunctionKey:
		n = key.NameF7
	case nsF8FunctionKey:
		n = key.NameF8
	case nsF9FunctionKey:
		n = key.NameF9
	case nsF10FunctionKey:
		n = key.NameF10
	case nsF11FunctionKey:
		n = key.NameF11
	case nsF12FunctionKey:
		n = key.NameF12
	case 0x20:
		n = key.NameSpace
	default:
		k = unicode.ToUpper(k)
		if !unicode.IsPrint(k) {
			return "", false
		}
		n = key.Name(k)
	}
	return n, true
}

func convertMods(mods uint) key.Modifiers {
	var kmods key.Modifiers
	if mods&nsAlternateKeyMask != 0 {
		kmods |= key.ModAlt
	}
	if mods&nsControlKeyMask != 0 {
		kmods |= key.ModCtrl
	}
	if mods&nsCommandKeyMask != 0 {
		kmods |= key.ModCommand
	}
	if mods&nsShiftKeyMask != 0 {
		kmods |= key.ModShift
	}
	return kmods
}

func (AppKitViewEvent) implementsViewEvent() {}
func (AppKitViewEvent) ImplementsEvent()     {}
func (a AppKitViewEvent) Valid() bool {
	return a != (AppKitViewEvent{})
}
