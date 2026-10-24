//go:build !js

// Package atspi exposes Keel windows to Linux assistive technology (Orca
// and other AT-SPI clients) over D-Bus.
//
// The application registers on the accessibility bus and serves one object
// per window and per element, at /org/a11y/atspi/accessible/<window>_<key>.
// ui/window hands each window's elements to Update after a frame and does
// the actions (press, scroll) on its own goroutine; this package never
// touches Gio. Calls arrive on godbus goroutines and read the last update
// under a mutex.
package atspi

import (
	"fmt"
	"image"
	"slices"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

// Node is one element of a window.
type Node struct {
	Key      string // stable while the element exists; a path segment
	Role     string // Keel's element role, see docs/automation.md
	Name     string
	Value    string
	Checked  *bool
	Selected *bool
	Disabled bool
	Focused  bool            // holds keyboard focus
	Bounds   image.Rectangle // window pixels
	Viewport image.Rectangle // the part of the window that can show it
	Parent   int             // index of the enclosing node, or -1
}

// Actions are what a client can ask of a window's elements. They are
// called on godbus goroutines and must hand the work to the window.
type Actions interface {
	Press(key string) bool
	Focus(key string) bool
	ScrollTo(key string) bool
	ScrollPage(key string, dx, dy int) bool
}

const (
	busName       = "org.a11y.atspi"
	rootPath      = dbus.ObjectPath("/org/a11y/atspi/accessible/root")
	pathPrefix    = "/org/a11y/atspi/accessible/"
	ifaceAccess   = "org.a11y.atspi.Accessible"
	ifaceApp      = "org.a11y.atspi.Application"
	ifaceComp     = "org.a11y.atspi.Component"
	ifaceAction   = "org.a11y.atspi.Action"
	ifaceObject   = "org.a11y.atspi.Event.Object"
	ifaceWindowEv = "org.a11y.atspi.Event.Window"
	ifaceProps    = "org.freedesktop.DBus.Properties"
)

// ref is an object reference: bus name and path.
type ref struct {
	Name string
	Path dbus.ObjectPath
}

// Bridge serves the accessible objects of an application.
type Bridge struct {
	conn    *dbus.Conn
	name    string
	appName string
	actions func(window int) Actions
	active  func()

	mu       sync.Mutex
	parent   ref // the registry's desktop
	windows  []*window
	wasAsked bool
}

type window struct {
	id      int
	title   string
	focused bool
	nodes   []Node
	index   map[string]int // key -> node
	kids    [][]int        // children of each node; kids[len(nodes)] are the top level
}

// Connect dials the accessibility bus that the session bus advertises.
func Connect() (*dbus.Conn, error) {
	session, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}
	var addr string
	if err := session.Object("org.a11y.Bus", "/org/a11y/bus").Call("org.a11y.Bus.GetAddress", 0).Store(&addr); err != nil {
		return nil, err
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return nil, err
	}
	if err := conn.Auth(nil); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.Hello(); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// New serves an application's objects on conn. actions returns a window's
// actions; active runs once, at the first request from a client.
func New(conn *dbus.Conn, appName string, actions func(window int) Actions, active func()) (*Bridge, error) {
	b := &Bridge{conn: conn, name: conn.Names()[0], appName: appName, actions: actions, active: active}
	for _, iface := range []string{ifaceAccess, ifaceApp, ifaceComp, ifaceAction, ifaceProps} {
		var v any
		switch iface {
		case ifaceAccess:
			v = accessibleIface{b}
		case ifaceApp:
			v = applicationIface{b}
		case ifaceComp:
			v = componentIface{b}
		case ifaceAction:
			v = actionIface{b}
		case ifaceProps:
			v = propsIface{b}
		}
		if err := conn.ExportSubtree(v, "/org/a11y/atspi/accessible", iface); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// Register announces the application to the registry, which lists it on
// the desktop clients start from.
func (b *Bridge) Register() error {
	var parent ref
	err := b.conn.Object("org.a11y.atspi.Registry", rootPath).Call("org.a11y.atspi.Socket.Embed", 0, ref{b.name, rootPath}).Store(&parent)
	if err != nil {
		return err
	}
	b.mu.Lock()
	b.parent = parent
	b.mu.Unlock()
	return nil
}

// Asked reports whether a client has made a request.
func (b *Bridge) Asked() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.wasAsked
}

// asked records a client request; the caller holds b.mu.
func (b *Bridge) asked() {
	if !b.wasAsked {
		b.wasAsked = true
		if b.active != nil {
			go b.active()
		}
	}
}

func (b *Bridge) find(id int) (*window, int) {
	for i, w := range b.windows {
		if w.id == id {
			return w, i
		}
	}
	return nil, -1
}

// AddWindow lists a new window under the application.
func (b *Bridge) AddWindow(id int, title string) {
	b.mu.Lock()
	w := &window{id: id, title: title, index: map[string]int{}, kids: make([][]int, 1)}
	b.windows = append(b.windows, w)
	n := len(b.windows) - 1
	b.mu.Unlock()
	b.emit(rootPath, "ChildrenChanged", "add", n, 0, dbus.MakeVariant(ref{b.name, b.windowPath(id)}))
}

// RemoveWindow drops a closed window.
func (b *Bridge) RemoveWindow(id int) {
	b.mu.Lock()
	w, i := b.find(id)
	if w == nil {
		b.mu.Unlock()
		return
	}
	b.windows = slices.Delete(b.windows, i, i+1)
	b.mu.Unlock()
	b.emit(b.windowPath(id), "StateChanged", "defunct", 1, 0, dbus.MakeVariant(int32(0)))
	b.emit(rootPath, "ChildrenChanged", "remove", i, 0, dbus.MakeVariant(ref{b.name, b.windowPath(id)}))
}

// SetFocused reports whether the window has keyboard focus.
func (b *Bridge) SetFocused(id int, focused bool) {
	b.mu.Lock()
	w, _ := b.find(id)
	if w == nil || w.focused == focused {
		b.mu.Unlock()
		return
	}
	w.focused = focused
	title := w.title
	b.mu.Unlock()
	path := b.windowPath(id)
	b.emit(path, "StateChanged", "active", boolInt(focused), 0, dbus.MakeVariant(int32(0)))
	member := "Deactivate"
	if focused {
		member = "Activate"
	}
	b.emitIface(path, ifaceWindowEv, member, "", 0, 0, dbus.MakeVariant(title))
}

// Update replaces a window's elements and tells clients what changed.
func (b *Bridge) Update(id int, title string, nodes []Node) {
	b.mu.Lock()
	w, _ := b.find(id)
	if w == nil {
		b.mu.Unlock()
		return
	}
	old := *w
	w.title = title
	w.nodes = append([]Node(nil), nodes...)
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
	type signal struct {
		path             dbus.ObjectPath
		member, detail   string
		detail1, detail2 int
		value            dbus.Variant
	}
	var signals []signal
	// Removed elements first, then additions, then changes.
	for _, n := range old.nodes {
		if _, ok := w.index[n.Key]; !ok {
			path := b.nodePath(id, n.Key)
			parentPath, at := b.parentOf(&old, n, id)
			signals = append(signals, signal{path, "StateChanged", "defunct", 1, 0, dbus.MakeVariant(int32(0))},
				signal{parentPath, "ChildrenChanged", "remove", at, 0, dbus.MakeVariant(ref{b.name, path})})
		}
	}
	for i, n := range w.nodes {
		path := b.nodePath(id, n.Key)
		j, existed := old.index[n.Key]
		if !existed || old.parentKey(j) != w.parentKey(i) {
			parentPath, at := b.parentOf(w, n, id)
			signals = append(signals, signal{parentPath, "ChildrenChanged", "add", at, 0, dbus.MakeVariant(ref{b.name, path})})
			continue
		}
		o := old.nodes[j]
		if o.Name != n.Name {
			signals = append(signals, signal{path, "PropertyChange", "accessible-name", 0, 0, dbus.MakeVariant(n.Name)})
		}
		if o.Value != n.Value && n.Value != "" {
			signals = append(signals, signal{path, "PropertyChange", "accessible-description", 0, 0, dbus.MakeVariant(n.Value)})
		}
		if !sameBool(o.Checked, n.Checked) && n.Checked != nil {
			signals = append(signals, signal{path, "StateChanged", "checked", boolInt(*n.Checked), 0, dbus.MakeVariant(int32(0))})
		}
		if !sameBool(o.Selected, n.Selected) && n.Selected != nil {
			signals = append(signals, signal{path, "StateChanged", "selected", boolInt(*n.Selected), 0, dbus.MakeVariant(int32(0))})
		}
		if o.Disabled != n.Disabled {
			signals = append(signals, signal{path, "StateChanged", "enabled", boolInt(!n.Disabled), 0, dbus.MakeVariant(int32(0))},
				signal{path, "StateChanged", "sensitive", boolInt(!n.Disabled), 0, dbus.MakeVariant(int32(0))})
		}
		if o.Focused != n.Focused {
			signals = append(signals, signal{path, "StateChanged", "focused", boolInt(n.Focused), 0, dbus.MakeVariant(int32(0))})
		}
		if o.showing() != n.showing() {
			signals = append(signals, signal{path, "StateChanged", "showing", boolInt(n.showing()), 0, dbus.MakeVariant(int32(0))})
		}
	}
	if old.title != title {
		signals = append(signals, signal{b.windowPath(id), "PropertyChange", "accessible-name", 0, 0, dbus.MakeVariant(title)})
	}
	b.mu.Unlock()
	for _, s := range signals {
		b.emit(s.path, s.member, s.detail, s.detail1, s.detail2, s.value)
	}
}

func (w *window) parentKey(i int) string {
	if p := w.nodes[i].Parent; p >= 0 {
		return w.nodes[p].Key
	}
	return ""
}

// parentOf returns the path of n's parent in w and n's index among its
// children.
func (b *Bridge) parentOf(w *window, n Node, id int) (dbus.ObjectPath, int) {
	path := b.windowPath(id)
	siblings := w.kids[len(w.nodes)]
	if n.Parent >= 0 {
		path = b.nodePath(id, w.nodes[n.Parent].Key)
		siblings = w.kids[n.Parent]
	}
	at := slices.IndexFunc(siblings, func(i int) bool { return w.nodes[i].Key == n.Key })
	return path, at
}

func (n Node) showing() bool { return !n.Bounds.Intersect(n.Viewport).Empty() }

func sameBool(a, b *bool) bool { return (a == nil) == (b == nil) && (a == nil || *a == *b) }

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (b *Bridge) emit(path dbus.ObjectPath, member, detail string, detail1, detail2 int, value dbus.Variant) {
	b.emitIface(path, ifaceObject, member, detail, detail1, detail2, value)
}

func (b *Bridge) emitIface(path dbus.ObjectPath, iface, member, detail string, detail1, detail2 int, value dbus.Variant) {
	_ = b.conn.Emit(path, iface+"."+member, detail, int32(detail1), int32(detail2), value, map[string]dbus.Variant{})
}

func (b *Bridge) windowPath(id int) dbus.ObjectPath {
	return dbus.ObjectPath(fmt.Sprintf("%sw%d", pathPrefix, id))
}

func (b *Bridge) nodePath(id int, key string) dbus.ObjectPath {
	return dbus.ObjectPath(fmt.Sprintf("%sw%d_%s", pathPrefix, id, key))
}

// object is what a path names: the application, a window or a node.
type object struct {
	app  bool
	w    *window
	node int // -1 for the window itself
}

// lookup resolves a path; the caller holds b.mu.
func (b *Bridge) lookup(path dbus.ObjectPath) (object, bool) {
	if path == rootPath {
		return object{app: true, node: -1}, true
	}
	rest, ok := strings.CutPrefix(string(path), pathPrefix+"w")
	if !ok {
		return object{}, false
	}
	idPart, key, hasKey := strings.Cut(rest, "_")
	var id int
	if _, err := fmt.Sscanf(idPart, "%d", &id); err != nil {
		return object{}, false
	}
	w, _ := b.find(id)
	if w == nil {
		return object{}, false
	}
	if !hasKey {
		return object{w: w, node: -1}, true
	}
	i, ok := w.index[key]
	if !ok {
		return object{}, false
	}
	return object{w: w, node: i}, true
}

func (b *Bridge) refOf(o object) ref {
	switch {
	case o.app:
		return ref{b.name, rootPath}
	case o.node < 0:
		return ref{b.name, b.windowPath(o.w.id)}
	}
	return ref{b.name, b.nodePath(o.w.id, o.w.nodes[o.node].Key)}
}

// children lists an object's children.
func (b *Bridge) children(o object) []object {
	var out []object
	switch {
	case o.app:
		for _, w := range b.windows {
			out = append(out, object{w: w, node: -1})
		}
	case o.node < 0:
		for _, i := range o.w.kids[len(o.w.nodes)] {
			out = append(out, object{w: o.w, node: i})
		}
	default:
		for _, i := range o.w.kids[o.node] {
			out = append(out, object{w: o.w, node: i})
		}
	}
	return out
}

func (b *Bridge) parentObject(o object) (object, bool) {
	switch {
	case o.app:
		return object{}, false
	case o.node < 0:
		return object{app: true, node: -1}, true
	}
	if p := o.w.nodes[o.node].Parent; p >= 0 {
		return object{w: o.w, node: p}, true
	}
	return object{w: o.w, node: -1}, true
}

func (b *Bridge) parentRef(o object) ref {
	if p, ok := b.parentObject(o); ok {
		return b.refOf(p)
	}
	if b.parent.Name != "" {
		return b.parent
	}
	return ref{"", "/org/a11y/atspi/null"}
}

func (b *Bridge) label(o object) string {
	switch {
	case o.app:
		return b.appName
	case o.node < 0:
		return o.w.title
	}
	return o.w.nodes[o.node].Name
}

func (b *Bridge) description(o object) string {
	if o.app || o.node < 0 {
		return ""
	}
	n := o.w.nodes[o.node]
	if n.Value != n.Name && n.Checked == nil {
		return n.Value // a select's choice, a progress, a badge count
	}
	return ""
}

func (b *Bridge) role(o object) uint32 {
	switch {
	case o.app:
		return roleApplication
	case o.node < 0:
		return roleFrame
	}
	return roleOf(o.w.nodes[o.node].Role)
}

func (b *Bridge) states(o object) stateSet {
	var s stateSet
	switch {
	case o.app:
		return s
	case o.node < 0:
		s.add(stateEnabled, stateSensitive, stateVisible, stateShowing, stateResizable)
		if o.w.focused {
			s.add(stateActive)
		}
		return s
	}
	n := o.w.nodes[o.node]
	s.add(stateVisible)
	if n.showing() {
		s.add(stateShowing)
	}
	if !n.Disabled {
		s.add(stateEnabled, stateSensitive)
	}
	if pressable(n.Role) {
		s.add(stateFocusable)
	}
	if n.Focused {
		s.add(stateFocused)
	}
	if n.Checked != nil {
		s.add(stateCheckable)
		if *n.Checked {
			s.add(stateChecked)
		}
		if n.Value == "mixed" {
			s.add(stateIndeterminate)
		}
	}
	if n.Selected != nil {
		s.add(stateSelectable)
		if *n.Selected {
			s.add(stateSelected)
		}
	}
	switch n.Role {
	case "textbox":
		s.add(stateEditable, stateSingleLine)
	case "disclosure", "treeitem":
		s.add(stateExpandable)
		if n.Value == "expanded" {
			s.add(stateExpanded)
		}
	}
	return s
}

// interfaces lists what an object implements.
func (b *Bridge) interfaces(o object) []string {
	out := []string{ifaceAccess}
	switch {
	case o.app:
		return append(out, ifaceApp)
	case o.node < 0:
		return append(out, ifaceComp)
	}
	out = append(out, ifaceComp)
	if pressable(o.w.nodes[o.node].Role) {
		out = append(out, ifaceAction)
	}
	return out
}

// extents returns an object's bounds in window pixels.
func (b *Bridge) extents(o object) image.Rectangle {
	switch {
	case o.app:
		return image.Rectangle{}
	case o.node < 0:
		var r image.Rectangle
		for _, n := range o.w.nodes {
			if n.Parent < 0 {
				r = r.Union(n.Viewport)
			}
		}
		return r
	}
	return o.w.nodes[o.node].Bounds
}

// pressable reports whether a press on an element can do something.
func pressable(role string) bool {
	switch roleOf(role) {
	case roleLabel, rolePanel, roleScrollPane, roleImage, roleList, roleTable, roleTree, roleToolBar,
		rolePageTabList, roleMenu, roleToolTip, roleProgressBar, roleDialog, roleAlert, roleStatusBar,
		roleForm, roleHeader, roleSeparator:
		return false
	}
	return true
}
