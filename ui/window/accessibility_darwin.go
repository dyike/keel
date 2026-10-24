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
	v.els = elementsOf(v.nodes, image.Rectangle{Max: size})
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
		frame := appkit.Rect{
			Origin: appkit.Point{X: float64(e.X) / scale, Y: bounds.Size.Height - float64(e.Y+e.Height)/scale},
			Size:   appkit.Size{Width: float64(e.Width) / scale, Height: float64(e.Height) / scale},
		}
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
	r := appkit.MsgRectToView(o.v.view, appkit.Sel("convertRect:toView:"), o.frame, 0)
	return appkit.MsgRectRect(win, appkit.Sel("convertRectToScreen:"), r)
}

// axPress clicks the center of the element, as VoiceOver's press does,
// where the latest frame drew it: the last refresh may be a frame behind.
func axPress(obj appkit.ID) bool {
	o := axObjects[obj]
	if o == nil {
		return false
	}
	if _, ok := nativeWindows.Load(uintptr(o.v.view)); !ok {
		return false
	}
	o.v.refresh(true)
	if axObjects[obj] != o || o.el.Disabled {
		return false // gone or disabled in the latest frame
	}
	w := o.v.w
	e := o.el
	p := f32.Pt(float32(e.X)+float32(e.Width)/2, float32(e.Y)+float32(e.Height)/2)
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
			if appkit.MsgRect(v.order[i].obj, appkit.Sel("accessibilityFrame")).Contains(p) {
				return v.order[i].obj
			}
		}
	}
	return view
}
