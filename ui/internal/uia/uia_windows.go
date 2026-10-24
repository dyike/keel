//go:build windows

// Package uia exposes Keel windows to Windows UI Automation (Narrator and
// other assistive technology) without cgo.
//
// A window answers WM_GETOBJECT with its root provider; the root and one
// provider per element implement IRawElementProviderSimple and
// IRawElementProviderFragment (the root also FragmentRoot), and, by role,
// the Invoke, Toggle, ScrollItem and Value patterns. ui/window hands each
// frame's elements to Update on the window's thread; UI Automation calls
// arrive on its own threads and read that copy under a lock. Actions are
// handed back to ui/window, which queues them as input for the window.
package uia

import (
	"image"
	"runtime"
	"slices"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Node is one element of a window.
type Node struct {
	Key      string // stable while the element exists; a decimal number
	Role     string // Keel's element role, see docs/automation.md
	Name     string
	Value    string
	Checked  *bool
	Selected *bool
	Disabled bool
	Focused  bool            // holds keyboard focus
	Bounds   image.Rectangle // client pixels
	Viewport image.Rectangle // the part of the window that can show it
	Parent   int             // index of the enclosing node, or -1
}

// Actions are what a client can ask of a window's elements. They are
// called on UI Automation's threads, never with the lock held.
type Actions interface {
	Press(key string) bool
	ScrollTo(key string) bool
}

var (
	uiacore              = windows.NewLazySystemDLL("UIAutomationCore.dll")
	procReturnProvider   = uiacore.NewProc("UiaReturnRawElementProvider")
	procHostFromHwnd     = uiacore.NewProc("UiaHostProviderFromHwnd")
	procRaiseStructure   = uiacore.NewProc("UiaRaiseStructureChangedEvent")
	procRaiseProperty    = uiacore.NewProc("UiaRaiseAutomationPropertyChangedEvent")
	procRaiseEvent       = uiacore.NewProc("UiaRaiseAutomationEvent")
	procClientsListening = uiacore.NewProc("UiaClientsAreListening")
	procDisconnect       = uiacore.NewProc("UiaDisconnectProvider")
	user32               = windows.NewLazySystemDLL("user32.dll")
	procClientToScreen   = user32.NewProc("ClientToScreen")
	procScreenToClient   = user32.NewProc("ScreenToClient")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
)

// Window serves one native window's elements.
type Window struct {
	hwnd     uintptr
	actions  Actions
	root     *object
	elements map[string]*object // live element objects by key
	nodes    []Node
	index    map[string]int
	kids     [][]int // children of each node; kids[len(nodes)] are the top level
	asked    bool
	closed   bool
}

// NewWindow serves hwnd's elements. Call it on the window's thread.
func NewWindow(hwnd uintptr, actions Actions) *Window {
	mu.Lock()
	defer mu.Unlock()
	w := &Window{hwnd: hwnd, actions: actions, elements: map[string]*object{}, index: map[string]int{}, kids: make([][]int, 1)}
	w.root = newObject(w, "")
	w.root.refs = 1 // ours, until Close
	return w
}

// GetObject answers WM_GETOBJECT; ok is false for other objects, which
// the window procedure passes on. Call it on the window's thread.
func (w *Window) GetObject(wparam, lparam uintptr) (result uintptr, ok bool) {
	if int32(lparam) != rootObjectID {
		return 0, false
	}
	mu.Lock()
	if w.closed {
		mu.Unlock()
		return 0, false
	}
	w.asked = true
	root := w.root.as(ifSimple)
	mu.Unlock()
	r, _, _ := procReturnProvider.Call(w.hwnd, wparam, lparam, root)
	mu.Lock()
	w.root.release() // UiaReturnRawElementProvider took its own
	mu.Unlock()
	return r, true
}

// Asked reports whether a client has asked for the window's elements.
func (w *Window) Asked() bool {
	mu.Lock()
	defer mu.Unlock()
	return w.asked
}

// Update replaces the elements and tells listening clients what changed.
// It never waits for UI Automation, which may be calling in.
func (w *Window) Update(nodes []Node) {
	mu.Lock()
	if w.closed {
		mu.Unlock()
		return
	}
	old, oldIndex := w.nodes, w.index
	w.nodes = append(w.nodes[:0:0], nodes...)
	w.index = make(map[string]int, len(nodes))
	for i, n := range nodes {
		w.index[n.Key] = i
	}
	w.kids = make([][]int, len(nodes)+1)
	for i, n := range nodes {
		p := n.Parent
		if p < 0 {
			p = len(nodes)
		}
		w.kids[p] = append(w.kids[p], i)
	}
	structure := len(old) != len(nodes)
	focusKey := func(ns []Node) string {
		for _, n := range ns {
			if n.Focused {
				return n.Key
			}
		}
		return ""
	}
	newFocus := focusKey(nodes)
	focusMoved := newFocus != focusKey(old)
	type change struct {
		key      string
		prop     int32
		old, new [2]uint64
	}
	var changes []change
	for i, n := range nodes {
		j, ok := oldIndex[n.Key]
		if !ok || old[j].Parent < 0 != (n.Parent < 0) || (n.Parent >= 0 && old[j].Parent >= 0 && old[old[j].Parent].Key != nodes[n.Parent].Key) {
			structure = true
			continue
		}
		o := old[j]
		if o.Checked != nil && n.Checked != nil && (*o.Checked != *n.Checked || o.Value != n.Value) {
			changes = append(changes, change{n.Key, propToggleState, [2]uint64{vtI4, uint64(toggleOf(o))}, [2]uint64{vtI4, uint64(toggleOf(n))}})
		}
		_ = i
	}
	for _, o := range old {
		if _, ok := w.index[o.Key]; !ok {
			structure = true
		}
	}
	mu.Unlock()
	if !structure && len(changes) == 0 && !(focusMoved && newFocus != "") {
		return
	}
	go func() {
		if r, _, _ := procClientsListening.Call(); r == 0 {
			return
		}
		if structure {
			mu.Lock()
			root := w.root.as(ifSimple)
			mu.Unlock()
			rid := [2]int32{appendRuntimeID, 0}
			procRaiseStructure.Call(root, structureChildrenInvalidated, uintptr(unsafe.Pointer(&rid[0])), 2)
			mu.Lock()
			w.root.release()
			mu.Unlock()
		}
		if focusMoved && newFocus != "" {
			mu.Lock()
			var p uintptr
			o := w.element(newFocus)
			if o != nil {
				p = o.as(ifSimple)
			}
			mu.Unlock()
			if p != 0 {
				procRaiseEvent.Call(p, eventFocusChanged)
				mu.Lock()
				o.release()
				mu.Unlock()
			}
		}
		for _, c := range changes {
			mu.Lock()
			o := w.element(c.key)
			var p uintptr
			if o != nil {
				p = o.as(ifSimple)
			}
			mu.Unlock()
			if p == 0 {
				continue
			}
			raiseProperty(p, c.prop, c.old, c.new)
			mu.Lock()
			o.release()
			mu.Unlock()
		}
	}()
}

// raiseProperty calls UiaRaiseAutomationPropertyChangedEvent, whose two
// VARIANTs travel by value: by reference on amd64, in two registers each
// on arm64, as four stack words each on 386.
func raiseProperty(provider uintptr, prop int32, old, new [2]uint64) {
	switch runtime.GOARCH {
	case "amd64":
		procRaiseProperty.Call(provider, uintptr(prop), uintptr(unsafe.Pointer(&old)), uintptr(unsafe.Pointer(&new)))
		runtime.KeepAlive(&old)
		runtime.KeepAlive(&new)
	case "arm64":
		procRaiseProperty.Call(provider, uintptr(prop), uintptr(old[0]), uintptr(old[1]), uintptr(new[0]), uintptr(new[1]))
	case "386":
		procRaiseProperty.Call(provider, uintptr(prop),
			uintptr(uint32(old[0])), uintptr(uint32(old[0]>>32)), uintptr(uint32(old[1])), uintptr(uint32(old[1]>>32)),
			uintptr(uint32(new[0])), uintptr(uint32(new[0]>>32)), uintptr(uint32(new[1])), uintptr(uint32(new[1]>>32)))
	}
}

// Disconnect tells UI Automation the window is going away. Call it on the
// window's thread, while handling WM_NCDESTROY or before.
func (w *Window) Disconnect() {
	mu.Lock()
	if w.closed {
		mu.Unlock()
		return
	}
	w.closed = true
	root := w.root.as(ifSimple)
	mu.Unlock()
	procReturnProvider.Call(w.hwnd, 0, 0, 0)
	procDisconnect.Call(root)
	mu.Lock()
	w.root.release()
	w.root.release() // ours
	mu.Unlock()
}

// element returns the live object of key, creating one; the caller holds
// mu.
func (w *Window) element(key string) *object {
	if _, ok := w.index[key]; !ok {
		return nil
	}
	if o := w.elements[key]; o != nil {
		return o
	}
	o := newObject(w, key)
	w.elements[key] = o
	return o
}

// node returns o's node, or false when o is the root or gone.
func (o *object) node() (Node, int, bool) {
	if o.key == "" || o.win.closed {
		return Node{}, -1, false
	}
	i, ok := o.win.index[o.key]
	if !ok {
		return Node{}, -1, false
	}
	return o.win.nodes[i], i, true
}

func (o *object) gone() bool {
	if o.win.closed {
		return true
	}
	if o.key == "" {
		return false
	}
	_, ok := o.win.index[o.key]
	return !ok
}

func (o *object) implements(i int) bool {
	switch i {
	case ifSimple, ifFragment:
		return true
	case ifRoot:
		return o.key == ""
	}
	n, _, ok := o.node()
	if !ok {
		return false
	}
	switch i {
	case ifInvoke:
		return pressable(n.Role) && n.Checked == nil
	case ifToggle:
		return n.Checked != nil
	case ifScrollItem:
		return n.Viewport != (image.Rectangle{}) && n.Viewport != o.win.windowRect()
	case ifValue:
		return hasValue(n.Role)
	}
	return false
}

// windowRect is the area the top-level elements share.
func (w *Window) windowRect() image.Rectangle {
	var r image.Rectangle
	for _, n := range w.nodes {
		if n.Parent < 0 {
			r = r.Union(n.Viewport)
		}
	}
	return r
}

func toggleOf(n Node) int32 {
	switch {
	case n.Value == "mixed":
		return toggleIndeterminate
	case n.Checked != nil && *n.Checked:
		return toggleOn
	}
	return toggleOff
}

// Provider methods. Each runs holding mu (see call).

func simpleProviderOptions(o *object, out uintptr) uintptr {
	if out != 0 {
		*(*int32)(unsafe.Pointer(out)) = providerServerSide
	}
	return sOK
}

func simplePatternProvider(o *object, pattern, out uintptr) uintptr {
	put(out, 0)
	iface := -1
	switch pattern {
	case patternInvoke:
		iface = ifInvoke
	case patternToggle:
		iface = ifToggle
	case patternScrollItem:
		iface = ifScrollItem
	case patternValue:
		iface = ifValue
	}
	if iface >= 0 && o.implements(iface) {
		put(out, o.as(iface))
	}
	return sOK
}

func simplePropertyValue(o *object, prop, out uintptr) uintptr {
	variant(out, vtEmpty, 0)
	if o.key == "" {
		// The window's own provider (the host) answers for the root.
		if prop == propFrameworkID {
			variant(out, vtBSTR, uint64(bstr("Keel")))
		}
		return sOK
	}
	n, _, ok := o.node()
	if !ok {
		return uiaElementNotAvailable
	}
	switch prop {
	case propControlType:
		variant(out, vtI4, uint64(controlType(n.Role)))
	case propLocalizedControlType:
		if name := localizedType(n.Role); name != "" {
			variant(out, vtBSTR, uint64(bstr(name)))
		}
	case propName:
		if n.Name != "" {
			variant(out, vtBSTR, uint64(bstr(n.Name)))
		}
	case propAutomationID:
		variant(out, vtBSTR, uint64(bstr(n.Key)))
	case propFrameworkID:
		variant(out, vtBSTR, uint64(bstr("Keel")))
	case propIsEnabled:
		variantBool(out, !n.Disabled)
	case propIsKeyboardFocusable:
		variantBool(out, pressable(n.Role) && !n.Disabled)
	case propHasKeyboardFocus:
		variantBool(out, n.Focused)
	case propIsOffscreen:
		variantBool(out, n.Bounds.Intersect(n.Viewport).Empty())
	case propIsControlElement, propIsContentElement:
		variantBool(out, true)
	case propItemStatus:
		if n.Value != "" && !hasValue(n.Role) && n.Checked == nil {
			variant(out, vtBSTR, uint64(bstr(n.Value)))
		}
	case propValueValue:
		if hasValue(n.Role) {
			variant(out, vtBSTR, uint64(bstr(n.Value)))
		}
	case propValueIsReadOnly:
		if hasValue(n.Role) {
			variantBool(out, true)
		}
	case propToggleState:
		if n.Checked != nil {
			variant(out, vtI4, uint64(toggleOf(n)))
		}
	}
	return sOK
}

func simpleHostProvider(o *object, out uintptr) uintptr {
	put(out, 0)
	if o.key == "" && !o.win.closed {
		procHostFromHwnd.Call(o.win.hwnd, out)
	}
	return sOK
}

func fragmentNavigate(o *object, direction, out uintptr) uintptr {
	put(out, 0)
	if o.gone() {
		return uiaElementNotAvailable
	}
	w := o.win
	var siblings []int
	at := -1
	parentIsRoot := true
	if n, i, ok := o.node(); ok {
		if n.Parent >= 0 {
			siblings, parentIsRoot = w.kids[n.Parent], false
		} else {
			siblings = w.kids[len(w.nodes)]
		}
		at = slices.Index(siblings, i)
	}
	give := func(i int) {
		if e := w.element(w.nodes[i].Key); e != nil {
			put(out, e.as(ifFragment))
		}
	}
	switch direction {
	case navParent:
		if o.key == "" {
			return sOK // the host is the root's parent
		}
		n, _, _ := o.node()
		if parentIsRoot {
			put(out, w.root.as(ifFragment))
		} else {
			give(n.Parent)
		}
	case navNextSibling:
		if o.key != "" && at+1 < len(siblings) {
			give(siblings[at+1])
		}
	case navPreviousSibling:
		if o.key != "" && at > 0 {
			give(siblings[at-1])
		}
	case navFirstChild, navLastChild:
		kids := w.kids[len(w.nodes)]
		if _, i, ok := o.node(); ok {
			kids = w.kids[i]
		}
		if len(kids) > 0 {
			if direction == navFirstChild {
				give(kids[0])
			} else {
				give(kids[len(kids)-1])
			}
		}
	}
	return sOK
}

func fragmentRuntimeID(o *object, out uintptr) uintptr {
	put(out, 0)
	if o.key == "" {
		return sOK // the host's
	}
	if o.gone() {
		return uiaElementNotAvailable
	}
	id, _ := strconv.ParseInt(o.key, 10, 32)
	put(out, intArray(appendRuntimeID, int32(id)))
	return sOK
}

// clientOrigin is the window's client area origin on the screen.
func (w *Window) clientOrigin() image.Point {
	var p [2]int32
	procClientToScreen.Call(w.hwnd, uintptr(unsafe.Pointer(&p)))
	return image.Pt(int(p[0]), int(p[1]))
}

func fragmentBoundingRectangle(o *object, out uintptr) uintptr {
	if out == 0 {
		return eInvalidArg
	}
	rect := (*[4]float64)(unsafe.Pointer(out))
	*rect = [4]float64{}
	if o.key == "" {
		return sOK // the host's
	}
	n, _, ok := o.node()
	if !ok {
		return uiaElementNotAvailable
	}
	r := n.Bounds.Add(o.win.clientOrigin())
	*rect = [4]float64{float64(r.Min.X), float64(r.Min.Y), float64(r.Dx()), float64(r.Dy())}
	return sOK
}

func fragmentEmbeddedRoots(o *object, out uintptr) uintptr {
	put(out, 0)
	return sOK
}

func fragmentSetFocus(o *object) uintptr { return sOK }

func fragmentRoot(o *object, out uintptr) uintptr {
	put(out, 0)
	if o.win.closed {
		return uiaElementNotAvailable
	}
	put(out, o.win.root.as(ifRoot))
	return sOK
}

// rootElementFromPoint answers for the mouse position: the point's
// coordinates are doubles, which reach no callback (see bind).
func rootElementFromPoint(o *object, args [5]uintptr) uintptr {
	out := args[2] // amd64: this, x, y in XMM registers (their slots unused), out
	switch runtime.GOARCH {
	case "arm64":
		out = args[0] // x and y in vector registers; out is the next integer register
	case "386":
		out = args[4] // x and y take two stack words each
	}
	put(out, 0)
	if o.win.closed {
		return uiaElementNotAvailable
	}
	var p [2]int32
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	procScreenToClient.Call(o.win.hwnd, uintptr(unsafe.Pointer(&p)))
	pt := image.Pt(int(p[0]), int(p[1]))
	for i := len(o.win.nodes) - 1; i >= 0; i-- {
		n := o.win.nodes[i]
		if pt.In(n.Bounds.Intersect(n.Viewport)) {
			if e := o.win.element(n.Key); e != nil {
				put(out, e.as(ifFragment))
			}
			return sOK
		}
	}
	return sOK
}

func rootFocus(o *object, out uintptr) uintptr {
	put(out, 0)
	if o.win.closed {
		return uiaElementNotAvailable
	}
	for _, n := range o.win.nodes {
		if n.Focused {
			if e := o.win.element(n.Key); e != nil {
				put(out, e.as(ifFragment))
			}
			break
		}
	}
	return sOK
}

// act runs a window action without the lock, which the callback holds.
func (o *object) act(do func(a Actions, key string) bool) uintptr {
	if o.gone() || o.key == "" {
		return uiaElementNotAvailable
	}
	a, key := o.win.actions, o.key
	mu.Unlock()
	ok := a != nil && do(a, key)
	mu.Lock()
	if !ok {
		return eAccessDenied
	}
	return sOK
}

func invokeInvoke(o *object) uintptr {
	return o.act(func(a Actions, key string) bool { return a.Press(key) })
}

func toggleToggle(o *object) uintptr {
	return o.act(func(a Actions, key string) bool { return a.Press(key) })
}

func toggleState(o *object, out uintptr) uintptr {
	n, _, ok := o.node()
	if !ok {
		return uiaElementNotAvailable
	}
	if out != 0 {
		*(*int32)(unsafe.Pointer(out)) = toggleOf(n)
	}
	return sOK
}

func scrollItemScrollIntoView(o *object) uintptr {
	return o.act(func(a Actions, key string) bool { return a.ScrollTo(key) })
}

func valueSetValue(o *object, value uintptr) uintptr { return eAccessDenied }

func valueValue(o *object, out uintptr) uintptr {
	put(out, 0)
	n, _, ok := o.node()
	if !ok {
		return uiaElementNotAvailable
	}
	put(out, bstr(n.Value))
	return sOK
}

func valueIsReadOnly(o *object, out uintptr) uintptr {
	if out != 0 {
		*(*int32)(unsafe.Pointer(out)) = 1
	}
	return sOK
}
