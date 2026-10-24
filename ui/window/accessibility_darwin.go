//go:build darwin && !ios

package window

// Platform accessibility on macOS. VoiceOver and other assistive apps ask
// Gio's content view for its accessibility children; Keel answers with one
// NSAccessibilityElement per element of the window (elementsOf, the model
// agents read), built from the semantic tree of the latest frame. AppKit
// asks on the main thread, where Gio keeps that tree, so no frame lock is
// taken. Nothing runs until a client asks; from then on each frame queues
// one refresh on the main thread, which updates only the objects that
// changed and posts AppKit's change notifications. An idle window does no
// work. A press (VoiceOver's VO-Space) clicks the element's center.

import (
	"image"
	"sync"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/ebitengine/purego/objc"
)

// Classes are extended once; the element class is registered once.
var axClasses struct {
	sync.Mutex
	views   map[objc.Class]bool
	element appkit.ID
}

// Main thread only.
var (
	axViews   = map[appkit.ID]*axView{}   // by content view (retained)
	axObjects = map[appkit.ID]*axObject{} // by element object
)

type axView struct {
	view     appkit.ID
	w        *Window
	nodes    []input.SemanticNode
	els      []Element
	objects  map[input.SemanticID]*axObject
	order    []*axObject // in element order: containers before what they hold
	children appkit.ID   // retained NSArray of the top-level objects
	gen      uint64
}

type axObject struct {
	obj    appkit.ID // owned
	v      *axView
	el     Element // as last applied
	set    bool
	parent appkit.ID
	frame  appkit.Rect // in view points
	shown  appkit.Rect // the part of it that can be seen, for hit testing
	gen    uint64
}

// installAccessibility gives the class of a Gio content view its
// accessibility methods. Called from the window's event goroutine while
// AppKit's main thread waits for the view event.
func installAccessibility(view uintptr) {
	cls := appkit.ID(view).Class()
	axClasses.Lock()
	defer axClasses.Unlock()
	if axClasses.element == 0 {
		axClasses.views = map[objc.Class]bool{}
		axClasses.element = appkit.RegisterClass("KeelAccessibilityElement", "NSAccessibilityElement", nil, []objc.MethodDef{
			appkit.Method("accessibilityPerformPress", func(self appkit.ID, _ appkit.SEL) bool { return axPress(self) }),
			// Scrolling has no method of its own in AppKit's element API
			// before macOS 26: it is a named action.
			// AppKit offers a press for every element of a class that can
			// press; text, groups and images cannot.
			appkit.Method("isAccessibilitySelectorAllowed:", func(self appkit.ID, _ appkit.SEL, sel appkit.SEL) bool {
				if o := axObjects[self]; o != nil && sel == appkit.Sel("accessibilityPerformPress") {
					return o.pressable()
				}
				return true
			}),
			appkit.Method("accessibilityActionNames", func(self appkit.ID, _ appkit.SEL) appkit.ID { return axActionNames(self) }),
			appkit.Method("accessibilityPerformAction:", func(self appkit.ID, _ appkit.SEL, name appkit.ID) { axPerformAction(self, name) }),
			// AppKit converts a frame in parent space to the screen once, when
			// it is set; this one follows the window when it moves.
			appkit.Method("accessibilityFrame", func(self appkit.ID, _ appkit.SEL) appkit.Rect { return axFrame(self) }),
		})
		axImps.children = objc.NewIMP(func(self appkit.ID, _ appkit.SEL) appkit.ID { return axChildren(self) })
		axImps.role = objc.NewIMP(func(self appkit.ID, _ appkit.SEL) appkit.ID { return appkit.String("AXGroup") })
		axImps.hitTest = objc.NewIMP(func(self appkit.ID, _ appkit.SEL, p appkit.Point) appkit.ID { return axHitTest(self, p) })
	}
	if axClasses.views[cls] {
		return
	}
	axClasses.views[cls] = true
	cls.AddMethod(appkit.Sel("accessibilityChildren"), axImps.children, "@@:")
	cls.AddMethod(appkit.Sel("accessibilityRole"), axImps.role, "@@:")
	cls.AddMethod(appkit.Sel("accessibilityHitTest:"), axImps.hitTest, "@@:{CGPoint=dd}")
}

var axImps struct{ children, role, hitTest objc.IMP }

// accessibilityFrame queues a refresh after a frame of a window an
// accessibility client has asked about, one at a time.
func accessibilityFrame(w *Window) {
	if !w.axActive.Load() || !w.axQueued.CompareAndSwap(false, true) {
		return
	}
	appkit.MainAsync(func() {
		w.axQueued.Store(false)
		for _, v := range axViews {
			if v.w == w {
				v.refresh(true)
			}
		}
	})
}

// forgetAccessibility drops a content view's elements before Gio's view
// goes away.
func forgetAccessibility(view appkit.ID) {
	appkit.MainAsync(func() {
		v := axViews[view]
		if v == nil {
			return
		}
		delete(axViews, view)
		for _, o := range v.objects {
			delete(axObjects, o.obj)
			appkit.Release(o.obj)
		}
		appkit.Release(v.children)
		appkit.Release(v.view)
	})
}

func axChildren(view appkit.ID) appkit.ID {
	v := axViews[view]
	if v == nil {
		value, ok := nativeWindows.Load(uintptr(view))
		if !ok {
			return 0
		}
		w := value.(*Window)
		v = &axView{view: appkit.Retain(view), w: w, objects: map[input.SemanticID]*axObject{}}
		axViews[view] = v
		w.axActive.Store(true)
		// Describe what is scrolled out of view from the next frame on. Not
		// here: invalidating from AppKit's callback re-enters Gio.
		go loop.SetAccessibility(true)
		v.refresh(false)
	}
	return v.children
}

// refresh rebuilds the elements from the latest frame and updates the
// objects that changed; notify posts the change notifications.
func (v *axView) refresh(notify bool) {
	if _, ok := nativeWindows.Load(uintptr(v.view)); !ok {
		return // detached; forgetAccessibility follows
	}
	scale := 1.0
	if win := appkit.Send(v.view, "window"); win != 0 {
		if s := appkit.MsgFloat(win, appkit.Sel("backingScaleFactor")); s > 0 {
			scale = s
		}
	}
	bounds := appkit.MsgRect(v.view, appkit.Sel("bounds"))
	size := image.Pt(int(bounds.Size.Width*scale+0.5), int(bounds.Size.Height*scale+0.5))
	v.nodes = v.w.win.AppendSemantics(v.nodes[:0])
	v.els = elementsOf(v.nodes, image.Rectangle{Max: size}, true)
	v.gen++
	layout := false
	order := v.order[:0]
	for _, e := range v.els {
		o := v.objects[e.id]
		if o == nil {
			o = &axObject{obj: appkit.Send(appkit.Send(axClasses.element, "alloc"), "init"), v: v}
			v.objects[e.id] = o
			axObjects[o.obj] = o
			layout = true
		}
		o.gen = v.gen
		parent := v.view
		if e.parent >= 0 {
			parent = order[e.parent].obj
		}
		if o.parent != parent {
			o.parent = parent
			appkit.Send(o.obj, "setAccessibilityParent:", uintptr(parent))
			layout = true
		}
		// View points, origin at the bottom left (Gio's view is not flipped).
		points := func(r image.Rectangle) appkit.Rect {
			return appkit.Rect{
				Origin: appkit.Point{X: float64(r.Min.X) / scale, Y: bounds.Size.Height - float64(r.Max.Y)/scale},
				Size:   appkit.Size{Width: float64(r.Dx()) / scale, Height: float64(r.Dy()) / scale},
			}
		}
		frame := points(e.box())
		o.shown = points(e.box().Intersect(e.viewport))
		if o.frame != frame {
			o.frame = frame
			layout = true
		}
		if !o.set || !sameElement(o.el, e) {
			valueChanged := o.set && (o.el.Value != e.Value || o.el.Name != e.Name || !sameBool(o.el.Checked, e.Checked) || !sameBool(o.el.Selected, e.Selected))
			o.apply(e)
			if notify && valueChanged {
				postAX(o.obj, "NSAccessibilityValueChangedNotification", "AXValueChanged")
			}
		}
		o.el = e
		order = append(order, o)
	}
	v.order = order
	for id, o := range v.objects {
		if o.gen != v.gen {
			delete(v.objects, id)
			delete(axObjects, o.obj)
			postAX(o.obj, "NSAccessibilityUIElementDestroyedNotification", "AXUIElementDestroyed")
			appkit.Release(o.obj)
			layout = true
		}
	}
	if !layout && v.children != 0 {
		return
	}
	// The tree changed: give every container, and the view, its children.
	arrays := make([]appkit.ID, len(order)+1)
	array := func(i int) appkit.ID {
		if arrays[i+1] == 0 {
			arrays[i+1] = appkit.Send(appkit.Class("NSMutableArray"), "array")
		}
		return arrays[i+1]
	}
	array(-1)
	for i, e := range v.els {
		if containerRoles[e.Role] {
			array(i)
		}
		appkit.Send(array(e.parent), "addObject:", uintptr(order[i].obj))
	}
	for i, e := range v.els {
		if containerRoles[e.Role] {
			appkit.Send(order[i].obj, "setAccessibilityChildren:", uintptr(arrays[i+1]))
			if e.Role == "table" {
				appkit.Send(order[i].obj, "setAccessibilityRows:", uintptr(arrays[i+1]))
			}
		}
	}
	appkit.Release(v.children)
	v.children = appkit.Retain(arrays[0])
	if notify {
		postAX(v.view, "NSAccessibilityLayoutChangedNotification", "AXLayoutChanged")
	}
}

func sameBool(a, b *bool) bool { return (a == nil) == (b == nil) && (a == nil || *a == *b) }

// sameElement reports whether two elements look the same to a client,
// ignoring where they are.
func sameElement(a, b Element) bool {
	return a.Role == b.Role && a.Name == b.Name && a.Value == b.Value && a.Disabled == b.Disabled &&
		sameBool(a.Checked, b.Checked) && sameBool(a.Selected, b.Selected)
}

func (o *axObject) apply(e Element) {
	o.set = true
	role, subrole := axRole(e.Role)
	appkit.Send(o.obj, "setAccessibilityRole:", uintptr(appkit.String(role)))
	var sub appkit.ID
	if subrole != "" {
		sub = appkit.String(subrole)
	}
	appkit.Send(o.obj, "setAccessibilitySubrole:", uintptr(sub))
	label, value := e.Name, e.Value
	if role == "AXStaticText" {
		label, value = "", e.Name // VoiceOver reads static text's value
		if e.Value != "" && e.Value != e.Name {
			label = e.Value // badge counts, statuses
		}
	}
	var str appkit.ID
	if label != "" {
		str = appkit.String(label)
	}
	appkit.Send(o.obj, "setAccessibilityLabel:", uintptr(str))
	var val appkit.ID
	switch {
	case e.Checked != nil:
		n := 0
		if *e.Checked {
			n = 1
		}
		if e.Value == "mixed" {
			n = 2
		}
		val = appkit.Send(appkit.Class("NSNumber"), "numberWithInt:", uintptr(n))
	case e.Role == "tab" && e.Selected != nil:
		val = appkit.Send(appkit.Class("NSNumber"), "numberWithInt:", appkit.Bool(*e.Selected))
	case e.Role == "disclosure":
		val = appkit.Send(appkit.Class("NSNumber"), "numberWithInt:", appkit.Bool(e.Value == "expanded"))
	case value != "":
		val = appkit.String(value)
	}
	appkit.Send(o.obj, "setAccessibilityValue:", uintptr(val))
	appkit.Send(o.obj, "setAccessibilityEnabled:", appkit.Bool(!e.Disabled))
	if e.Selected != nil {
		appkit.Send(o.obj, "setAccessibilitySelected:", appkit.Bool(*e.Selected))
	}
}

// axRole maps an element role (see docs/automation.md) to AppKit's role
// and subrole.
func axRole(role string) (string, string) {
	switch role {
	case "button", "columnheader", "gridcell":
		return "AXButton", ""
	case "link":
		return "AXLink", ""
	case "textbox":
		return "AXTextField", ""
	case "checkbox", "menuitemcheckbox":
		return "AXCheckBox", ""
	case "switch":
		return "AXCheckBox", "AXSwitch"
	case "toggle":
		return "AXCheckBox", "AXToggle"
	case "radio":
		return "AXRadioButton", ""
	case "tab":
		return "AXRadioButton", "AXTabButton"
	case "tablist":
		return "AXTabGroup", ""
	case "select", "combobox":
		return "AXPopUpButton", ""
	case "slider":
		return "AXSlider", ""
	case "progressbar":
		return "AXProgressIndicator", ""
	case "image", "img", "avatar":
		return "AXImage", ""
	case "menu":
		return "AXMenu", ""
	case "menuitem":
		return "AXMenuItem", ""
	case "list", "listbox":
		return "AXList", ""
	case "table":
		return "AXTable", ""
	case "row", "treeitem":
		return "AXRow", ""
	case "tree":
		return "AXOutline", ""
	case "disclosure":
		return "AXDisclosureTriangle", ""
	case "toolbar":
		return "AXToolbar", ""
	case "separator":
		return "AXSplitter", ""
	case "tooltip":
		return "AXHelpTag", ""
	case "scrollarea":
		return "AXScrollArea", ""
	}
	if containerRoles[role] {
		return "AXGroup", ""
	}
	return "AXStaticText", ""
}

func postAX(obj appkit.ID, constant, fallback string) {
	name := appkit.Constant(constant)
	if name == 0 {
		name = appkit.String(fallback)
	}
	appkit.Call(appkit.Sym("NSAccessibilityPostNotification"), uintptr(obj), uintptr(name))
}

// axFrame converts an element's frame to the screen.
func axFrame(obj appkit.ID) appkit.Rect {
	o := axObjects[obj]
	if o == nil {
		return appkit.Rect{}
	}
	win := appkit.Send(o.v.view, "window")
	if win == 0 {
		return appkit.Rect{}
	}
	return axScreen(o.v.view, o.frame)
}

// axScreen converts a rectangle of view points to the screen.
func axScreen(view appkit.ID, r appkit.Rect) appkit.Rect {
	win := appkit.Send(view, "window")
	if win == 0 || r.Size.Width <= 0 || r.Size.Height <= 0 {
		return appkit.Rect{}
	}
	r = appkit.MsgRectToView(view, appkit.Sel("convertRect:toView:"), r, 0)
	return appkit.MsgRectRect(win, appkit.Sel("convertRectToScreen:"), r)
}

const (
	axPressAction       = "AXPress"
	axScrollToVisible   = "AXScrollToVisible"
	axScrollUpByPage    = "AXScrollUpByPage"
	axScrollDownByPage  = "AXScrollDownByPage"
	axScrollLeftByPage  = "AXScrollLeftByPage"
	axScrollRightByPage = "AXScrollRightByPage"
)

// pressable reports whether a press can do something.
func (o *axObject) pressable() bool {
	switch role, _ := axRole(o.el.Role); role {
	case "AXStaticText", "AXGroup", "AXScrollArea", "AXImage", "AXList", "AXTable", "AXOutline", "AXToolbar", "AXTabGroup", "AXMenu", "AXHelpTag", "AXProgressIndicator":
		return false
	}
	return !o.el.Disabled
}

// axActionNames lists what an element does: every element can be scrolled
// into view, controls can be pressed, scroll areas scroll by page.
func axActionNames(obj appkit.ID) appkit.ID {
	names := appkit.Send(appkit.Class("NSMutableArray"), "array")
	add := func(s string) { appkit.Send(names, "addObject:", uintptr(appkit.String(s))) }
	o := axObjects[obj]
	if o == nil {
		return names
	}
	if o.pressable() {
		add(axPressAction)
	}
	if o.el.Role == "scrollarea" {
		add(axScrollUpByPage)
		add(axScrollDownByPage)
		add(axScrollLeftByPage)
		add(axScrollRightByPage)
	}
	add(axScrollToVisible)
	return names
}

func axPerformAction(obj appkit.ID, name appkit.ID) {
	switch appkit.GoString(name) {
	case axPressAction:
		axPress(obj)
	case axScrollToVisible:
		axScrollTo(obj)
	case axScrollUpByPage:
		axScrollPage(obj, 0, -1)
	case axScrollDownByPage:
		axScrollPage(obj, 0, 1)
	case axScrollLeftByPage:
		axScrollPage(obj, -1, 0)
	case axScrollRightByPage:
		axScrollPage(obj, 1, 0)
	}
}

// axLatest refreshes obj's window and reports obj's element in the
// latest frame.
func axLatest(obj appkit.ID) (*axObject, bool) {
	o := axObjects[obj]
	if o == nil {
		return nil, false
	}
	if _, ok := nativeWindows.Load(uintptr(o.v.view)); !ok {
		return nil, false
	}
	o.v.refresh(true)
	return o, axObjects[obj] == o
}

// axScrollTo scrolls the area that holds the element until the element
// shows, its top first when it is taller than the area: a wheel scroll at
// the area's center, as a person would.
func axScrollTo(obj appkit.ID) {
	o, ok := axLatest(obj)
	if !ok {
		return
	}
	b, vp := o.el.box(), o.el.viewport
	delta := func(lo, hi, vlo, vhi int) int {
		switch {
		case lo < vlo:
			return lo - vlo
		case hi > vhi:
			return min(hi-vhi, lo-vlo)
		}
		return 0
	}
	axScroll(o.v.w, vp, delta(b.Min.X, b.Max.X, vp.Min.X, vp.Max.X), delta(b.Min.Y, b.Max.Y, vp.Min.Y, vp.Max.Y))
}

// axScrollPage scrolls a scroll area by most of its size.
func axScrollPage(obj appkit.ID, dx, dy int) {
	o, ok := axLatest(obj)
	if !ok {
		return
	}
	area := o.el.box().Intersect(o.el.viewport)
	axScroll(o.v.w, area, dx*area.Dx()*9/10, dy*area.Dy()*9/10)
}

func axScroll(w *Window, area image.Rectangle, dx, dy int) {
	if area.Empty() || dx == 0 && dy == 0 {
		return
	}
	c := area.Min.Add(area.Max).Div(2)
	w.win.ProcessInput(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(float32(c.X), float32(c.Y)), Scroll: f32.Pt(float32(dx), float32(dy))})
}

// axPress clicks the center of the element's visible part, as VoiceOver's
// press does, where the latest frame drew it: the last refresh may be a
// frame behind.
func axPress(obj appkit.ID) bool {
	o, ok := axLatest(obj)
	if !ok || o.el.Disabled {
		return false // gone or disabled in the latest frame
	}
	w := o.v.w
	shown := o.el.box().Intersect(o.el.viewport)
	if shown.Empty() {
		return false // scrolled out of view: a client scrolls to it first
	}
	c := shown.Min.Add(shown.Max).Div(2)
	p := f32.Pt(float32(c.X), float32(c.Y))
	w.win.ProcessInput(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p})
	w.win.ProcessInput(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p})
	w.win.ProcessInput(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
	return true
}

// axHitTest returns the element under a screen point: the last listed
// element containing it is the innermost, or the one drawn on top.
func axHitTest(view appkit.ID, p appkit.Point) appkit.ID {
	if v := axViews[view]; v != nil {
		for i := len(v.order) - 1; i >= 0; i-- {
			if axScreen(v.view, v.order[i].shown).Contains(p) {
				return v.order[i].obj
			}
		}
	}
	return view
}
