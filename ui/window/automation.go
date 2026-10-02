package window

// Automation mode lets a test agent drive an app.
//
// Start the app with KEEL_AUTOMATION set: to a unix socket path, or to 1 to
// listen at SocketDir()/<program>-<pid>.sock, where keel-mcp's attach finds it.
// Every window then gets a virtual twin that renders in memory through its own
// Gio input router, so the same components, callbacks, shortcuts and
// core.Update run as for a person. Requests (see automation_server.go) act on
// the twin and render it on demand. Coordinates are in dp and screenshots are
// 1×, so a pixel in a screenshot is a dp in every request.
//
// Visible (default): real windows stay on screen for the user, and each twin
// is a shadow that follows its window's size. Both lay out the same
// components, so what an agent changes through the shadow shows up in the real
// window, and the other way round. Focus is per router: a field the agent
// focuses shows no caret on screen.
//
// Headless (KEEL_HEADLESS=1): no real windows at all; Main only serves requests.

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	gioevent "gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/theme"
)

var auto struct {
	addr string

	mu      sync.Mutex
	windows []*Window // open windows, in opening order
	active  *Window   // default target of requests
	nextID  int
	actions []pendingAction // Close/Raise requested from UI code
}

func init() {
	switch v := os.Getenv("KEEL_AUTOMATION"); v {
	case "", "0":
	case "1", "true", "on":
		name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		auto.addr = filepath.Join(SocketDir(), fmt.Sprintf("%s-%d.sock", name, os.Getpid()))
	default:
		auto.addr = v
	}
	if auto.addr != "" {
		// Agents act on what a snapshot reports, so nothing may still be
		// sliding into place; the app can turn motion back on.
		theme.SetReducedMotion(true)
	}
}

// SocketDir is where apps started with KEEL_AUTOMATION=1 listen. keel-mcp
// computes the same path; keep the two in sync.
func SocketDir() string { return filepath.Join(os.TempDir(), "keel") }

func automating() bool { return auto.addr != "" }

func offScreen() bool { return automating() && os.Getenv("KEEL_HEADLESS") == "1" }

type virtual struct {
	id      string
	size    image.Point
	router  input.Router
	ops     op.Ops
	gpu     *headless.Window
	gpuSize image.Point
	refs    []Element // from the last snapshot, indexed by ref number
}

type pendingAction struct {
	w *Window
	a system.Action
}

// openVirtual gives w its twin. alone is true for headless windows, which
// have no real window to redraw.
func openVirtual(w *Window, alone bool) {
	auto.mu.Lock()
	auto.nextID++
	w.virt = &virtual{id: fmt.Sprintf("w%d", auto.nextID), size: image.Pt(w.opts.Width, w.opts.Height)}
	auto.windows = append(auto.windows, w)
	auto.active = w
	auto.mu.Unlock()
	if alone {
		loop.Register(w, func() {}) // frames are rendered on request, not on invalidation
	}
}

func (v *virtual) setSize(s image.Point) {
	auto.mu.Lock()
	v.size = s
	auto.mu.Unlock()
}

func (v *virtual) getSize() image.Point {
	auto.mu.Lock()
	defer auto.mu.Unlock()
	return v.size
}

// forgetVirtual drops a closed window from the automation list.
func forgetVirtual(w *Window) {
	auto.mu.Lock()
	auto.windows = slices.DeleteFunc(auto.windows, func(x *Window) bool { return x == w })
	if auto.active == w {
		auto.active = nil
		if n := len(auto.windows); n > 0 {
			auto.active = auto.windows[n-1]
		}
	}
	auto.mu.Unlock()
	if w.virt.gpu != nil {
		w.virt.gpu.Release()
	}
}

// queueAction defers Close and Raise until the current request has rendered:
// UI code calls them while holding the frame lock.
func queueAction(w *Window, a system.Action) {
	auto.mu.Lock()
	auto.actions = append(auto.actions, pendingAction{w, a})
	auto.mu.Unlock()
}

func runActions() {
	auto.mu.Lock()
	acts := auto.actions
	auto.actions = nil
	auto.mu.Unlock()
	for _, p := range acts {
		switch {
		case p.a&system.ActionClose != 0:
			closeVirtual(p.w)
		case p.a&system.ActionRaise != 0:
			auto.mu.Lock()
			if !p.w.closed {
				auto.active = p.w
			}
			auto.mu.Unlock()
		}
	}
}

func closeVirtual(w *Window) {
	if w.closed {
		return
	}
	w.finish()
	forgetVirtual(w)
}

// render lays out frames until nothing asks for an immediate redraw: a callback
// that changed state needs a second frame to show it. Timed redraws (caret
// blink, button ink) are left for later.
func (w *Window) render() {
	v := w.virt
	size := v.getSize()
	for range 10 {
		v.ops.Reset()
		gtx := layout.Context{Ops: &v.ops, Now: time.Now(), Source: v.router.Source(),
			Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
		loop.Lock()
		loop.Drain()
		w.layout(gtx)
		loop.Unlock()
		v.router.Frame(&v.ops)
		if t, ok := v.router.WakeupTime(); !ok || t.After(time.Now()) {
			break
		}
	}
	// The real windows share this state but only redraw when invalidated, and
	// core.Call only invalidates for components with a callback: a checkbox
	// without OnChange would stay stale on screen until the next mouse move.
	loop.InvalidateAll()
	runActions()
}

// Element is one node of a window's semantic tree, as reported to agents.
type Element struct {
	Ref      string `json:"ref"`
	Role     string `json:"role"` // see roleOf
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`    // textbox content, select choice, progress
	Checked  *bool  `json:"checked,omitempty"`  // checkbox, radio, switch
	Selected *bool  `json:"selected,omitempty"` // tab, row, option
	Disabled bool   `json:"disabled,omitempty"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

func (e Element) center() f32.Point {
	return f32.Pt(float32(e.X)+float32(e.Width)/2, float32(e.Y)+float32(e.Height)/2)
}

// snapshot renders and flattens the semantic tree into elements. An element
// absorbs the nodes inside it, so a button with a label is one element, not two.
func (w *Window) snapshot() []Element {
	w.render()
	v := w.virt
	nodes := v.router.AppendSemantics(nil)
	var out []Element
	var walk func(n input.SemanticNode, inControl bool, visible image.Rectangle)
	walk = func(n input.SemanticNode, inControl bool, visible image.Rectangle) {
		d := n.Desc
		if d.Description == "el-inert" {
			return
		}
		// Gio's bounds ignore scroll clipping: a row scrolled out of a table
		// still reports where it would be. Keep only what can be seen.
		b := d.Bounds.Intersect(visible)
		if b.Empty() {
			return
		}
		role, value := roleOf(d, inControl)
		if role != "" {
			e := Element{Role: role, Name: d.Label, Value: value, Disabled: d.Disabled,
				X: b.Min.X, Y: b.Min.Y, Width: b.Dx(), Height: b.Dy()}
			state := d.Selected
			switch role {
			case "checkbox", "radio", "switch":
				e.Checked = &state
			case "tab", "row", "option", "disclosure", "toggle", "tag", "gridcell":
				e.Selected = &state
			}
			if role != "text" && !containerRoles[role] && e.Name == "" {
				e.Name = childText(n)
			}
			e.Ref = fmt.Sprintf("e%d", len(out)+1)
			out = append(out, e)
		}
		// An element absorbs the nodes inside it (a button's label), except
		// containers, whose children are elements of their own.
		absorb := role != "" && !containerRoles[role] && !strings.HasPrefix(d.Description, "paragraph")
		if role != "" || !d.Bounds.Eq(nodes[0].Desc.Bounds) {
			visible = b // an element or clip area bounds what is inside it
		}
		for _, c := range n.Children {
			walk(c, inControl || absorb, visible)
		}
	}
	if len(nodes) > 0 {
		walk(nodes[0], false, image.Rectangle{Max: w.virt.getSize()})
	}
	v.refs = out
	return out
}

var containerRoles = map[string]bool{"dialog": true, "figure": true, "banner": true, "region": true, "toolbar": true, "tablist": true, "tabpanel": true, "log": true, "article": true, "tree": true, "navigation": true, "attachment": true, "menu": true, "alertdialog": true, "grid": true, "radiogroup": true, "listbox": true, "list": true, "form": true, "combobox": true, "table": true, "code": true, "footnotes": true, "accordion": true, "alert": true, "group": true, "status": true, "tag": true}

// roleOf maps a semantic node to an element role and value. Gio's classes
// give the common roles; core.Role descriptions ("row", "select:北京") the rest.
func roleOf(d input.SemanticDesc, inControl bool) (role, value string) {
	if inControl { // part of an element already listed, e.g. Gio's own node inside our button
		return "", ""
	}
	custom, val, _ := strings.Cut(d.Description, ":")
	switch d.Class {
	case semantic.Button:
		switch custom {
		case "button", "link", "tab", "columnheader", "select", "image", "disclosure", "toggle":
			return custom, val
		}
		return "button", ""
	case semantic.CheckBox:
		return "checkbox", val
	case semantic.RadioButton:
		return "radio", ""
	case semantic.Switch:
		return "switch", ""
	case semantic.Editor:
		return "textbox", d.Description
	}
	switch custom {
	case "", "el-inert":
	case "paragraph": // text with links: listed as text, its links after it
		return "text", ""
	default: // any role a component declares with core.Role or el's Role
		return custom, val
	}
	if d.Label != "" {
		return "text", ""
	}
	return "", ""
}

func childText(n input.SemanticNode) string {
	var parts []string
	for _, c := range n.Children {
		if c.Desc.Label != "" {
			parts = append(parts, c.Desc.Label) // a labelled node speaks for what is inside it
			continue
		}
		if t := childText(c); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

// find resolves a ref from the last snapshot, or searches a fresh one by text:
// an exact name match wins over a partial one, controls over plain text.
func (w *Window) find(ref, text string) (Element, error) {
	if ref != "" {
		var i int
		if _, err := fmt.Sscanf(ref, "e%d", &i); err != nil || i < 1 || i > len(w.virt.refs) {
			return Element{}, fmt.Errorf("unknown ref %q: take a new snapshot", ref)
		}
		return w.virt.refs[i-1], nil
	}
	els := w.snapshot()
	best, score := Element{}, 0
	for _, e := range els {
		s := 0
		switch {
		case e.Name == text || e.Value == text:
			s = 4
		case strings.Contains(e.Name, text) || strings.Contains(e.Value, text):
			s = 2
		}
		if s > 0 && e.Role != "text" {
			s++
		}
		if s > 0 && s >= score { // on a tie, the later one is drawn on top (dialogs, popups)
			best, score = e, s
		}
	}
	if score == 0 {
		return Element{}, fmt.Errorf("no element matches %q", text)
	}
	return best, nil
}

func (w *Window) click(p f32.Point) {
	w.render() // input only reaches handlers registered by a frame
	w.virt.router.Queue(
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p},
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p},
		offscreen, // leave, so the button does not stay hovered in the real window
	)
	w.render()
}

// offscreen moves the virtual pointer out of every area, ending hover states
// that the real window would otherwise keep showing.
var offscreen = pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(-1e6, -1e6)}

func (w *Window) scroll(p f32.Point, dy float32) {
	w.render()
	w.virt.router.Queue(
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p},
		pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: p, Scroll: f32.Pt(0, dy)},
		offscreen,
	)
	w.render()
}

// typeText inserts s at the caret of the focused text box, replacing any selection.
func (w *Window) typeText(s string) error {
	w.render()
	v := w.virt
	st := v.router.EditorState()
	v.router.Queue(key.EditEvent{Range: st.Selection.Range, Text: s})
	if _, handled := v.router.WakeupTime(); !handled {
		return fmt.Errorf("no text box has focus: click one first or pass its ref")
	}
	w.render()
	return nil
}

// press sends one key chord, e.g. "enter", "tab", "mod+a". Like a real window,
// an unhandled Tab or Shift+Tab moves focus.
func (w *Window) press(spec string) error {
	sc, err := parseShortcut(spec)
	if err != nil {
		return err
	}
	w.render() // e.g. the first request after launch: no frame has registered shortcuts yet
	v := w.virt
	var ev gioevent.Event = key.Event{Name: sc.name, Modifiers: sc.mods, State: key.Press}
	dir := key.FocusDirection(-1)
	if sc.name == key.NameTab && sc.mods == 0 {
		dir = key.FocusForward
	} else if sc.name == key.NameTab && sc.mods == key.ModShift {
		dir = key.FocusBackward
	}
	if dir != -1 {
		ev = input.SystemEvent{Event: ev}
	}
	v.router.Queue(ev)
	if _, handled := v.router.WakeupTime(); !handled && dir != -1 {
		v.router.MoveFocus(dir)
		v.router.RevealFocus(image.Rectangle{Max: v.getSize()})
	}
	v.router.Queue(key.Event{Name: sc.name, Modifiers: sc.mods, State: key.Release})
	w.render()
	return nil
}

func (w *Window) screenshot() ([]byte, error) {
	w.render()
	v := w.virt
	size := v.getSize()
	if v.gpu != nil && v.gpuSize != size { // the real window was resized
		v.gpu.Release()
		v.gpu = nil
	}
	if v.gpu == nil {
		g, err := headless.NewWindow(size.X, size.Y)
		if err != nil {
			return nil, err
		}
		v.gpu, v.gpuSize = g, size
	}
	if err := v.gpu.Frame(&v.ops); err != nil {
		return nil, err
	}
	img := image.NewRGBA(image.Rectangle{Max: size})
	if err := v.gpu.Screenshot(img); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}
