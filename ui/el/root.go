package el

import (
	"image"

	"github.com/dyike/keel/ui/locale"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/op/clip"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// View is anything that renders an element tree: usually a struct holding the
// state of a screen, whose event handlers change its fields directly.
//
//	type Counter struct{ n int }
//
//	func (c *Counter) Render(cx *el.Context) el.Element {
//		return el.Div().Child(el.Text(strconv.Itoa(c.n)), el.Div().OnClick(func() { c.n++ }).Child(el.Text("+1")))
//	}
//
// Render runs every frame, after event handlers, under the UI lock (see
// ui/core). Code on other goroutines changes views through core.Update.
type View interface {
	Render(cx *Context) Element
}

// ViewFunc adapts a render function to a View.
type ViewFunc func(*Context) Element

func (f ViewFunc) Render(cx *Context) Element { return f(cx) }

// Context is passed to Render.
type Context struct {
	layers         []overlayDecl
	shortcuts      []viewShortcut
	actions        []scopedAction
	globalActions  map[string]func() // by name, for Perform
	bindingTargets map[string]actionBindingTarget
	root           *RootWidget
}

// ClickModifiers reports modifier keys during the current pointer click
// callback (including double click); outside that callback it returns zero.
func (cx *Context) ClickModifiers() key.Modifiers { return cx.root.clickModifiers }

// Cache returns the element built for key, calling build only when key was
// not used in the previous frame. While the width it is given stays the same,
// the element's layout is reused too, so long, mostly unchanging content (a
// chat history, a rendered document) costs little per frame. key must be
// comparable and change whenever the element would look different; the
// element must not depend on anything else. Entries unused for a frame are
// dropped. Applying a theme also rebuilds cached elements.
func (cx *Context) Cache(key any, build func() Element) Element {
	c := &cx.root.cache
	if rev := [2]uint64{theme.Revision(), locale.Revision()}; c.revision != rev {
		clear(c.entries)
		c.revision = rev
	}
	if e, ok := c.entries[key]; ok {
		e.frame = cx.root.store.frame
		return e.el
	}
	el := build()
	if el == nil {
		return nil
	}
	el.node().cached = true
	c.entries[key] = &cacheEntry{el: el, frame: cx.root.store.frame}
	return el
}

type elementCache struct {
	entries  map[any]*cacheEntry
	revision [2]uint64 // theme and locale: either change rebuilds cached elements
}

type cacheEntry struct {
	el    Element
	frame uint64
}

func (c *elementCache) sweep(frame uint64) {
	for k, e := range c.entries {
		if e.frame != frame {
			delete(c.entries, k)
		}
	}
}

type viewShortcut struct {
	name key.Name
	mods key.Modifiers
	fn   func()
}

// Shortcut binds a key chord such as "mod+s" to fn while the window has focus
// and this view is rendered. It panics on an invalid chord.
func (cx *Context) Shortcut(chord string, fn func()) {
	name, mods, err := core.ParseShortcut(chord)
	if err != nil {
		panic("el: " + err.Error())
	}
	cx.shortcuts = append(cx.shortcuts, viewShortcut{name, mods, fn})
}

// Action handles a named action from the keymap (core.Bind) while the window
// has focus and this view is rendered: every chord bound to name runs fn.
// Rebinding takes effect on the next frame; an unbound action does nothing.
func (cx *Context) Action(name string, fn func()) {
	if cx.globalActions == nil {
		cx.globalActions = map[string]func(){}
	}
	if _, ok := cx.globalActions[name]; !ok {
		cx.globalActions[name] = fn
	}
	for _, chord := range core.Bindings(name) {
		if k, mods, err := core.ParseShortcut(chord); err == nil {
			cx.shortcuts = append(cx.shortcuts, viewShortcut{k, mods, fn})
		}
	}
}

// RootWidget renders a View as a core.Widget.
type RootWidget struct {
	mounts         []mount // views rendered after the root view; see Mount
	clickModifiers key.Modifiers
	requestedFocus event.Tag
	mainTree       *Node
	layers         map[any]*layerState
	activeLayers   []*layerState
	layerSerial    uint64
	timerEpoch     uint64
	timers         map[any]*viewTimer
	polls          map[any]*viewPoll
	source         input.Source
	callbacks      bool
	bg             int // tag of the area under everything; see blur
	view           View
	fill           bool
	store          *store
	cache          elementCache
	e              engine
	focusID        string
	focusPending   bool

	pointerDispatch  bool
	focusFromPointer bool
	// keyboardModality is whether the last press came from the keyboard. Focus
	// returned by a closing overlay shows its ring only then, like
	// :focus-visible.
	keyboardModality bool
}

// Root makes v the whole content of a window: it fills the window, with the
// theme background, and the window adds no padding or scrolling of its own.
//
//	window.Open(window.Options{Title: "Orders", Content: el.Root(&Orders{})})
func Root(v View) *RootWidget {
	return &RootWidget{view: v, fill: true, store: newStore(), cache: elementCache{entries: map[any]*cacheEntry{}}}
}

// Embed renders v as an ordinary widget sized to its content, e.g. inside
// hand-written Gio layout or a window that pads and scrolls its content.
func Embed(v View) *RootWidget {
	return &RootWidget{view: v, store: newStore(), cache: elementCache{entries: map[any]*cacheEntry{}}}
}

// FillsWindow tells ui/window to give the root the whole window.
func (r *RootWidget) FillsWindow() bool { return r.fill }

// SetTextAtlas enables an optional image cache for simple numeric labels.
// The root must be drawn at an integral pixel translation, without additional
// scaling or rotation. Decorated and scroll subtrees, overlays, rich text and
// editors retain vector drawing. Pass nil to disable the cache.
//
// Use a separate atlas for each root, call this from UI code, and call the
// atlas's Release when the root is discarded. The caller owns the atlas.
func (r *RootWidget) SetTextAtlas(atlas *theme.GlyphAtlas) {
	r.e.textAtlas = atlas
	r.e.textAtlasActive = false
	r.e.atlasLabels = nil
	r.e.atlasGlyphs = nil
}

func (r *RootWidget) Layout(gtx core.C) core.D {
	// A missing input source can mean measurement or a disabled parent.
	// Both render real state without advancing its lifecycle.
	live := gtx.Enabled()
	st := r.store
	if live {
		st.frame++
		theme.BeginTextFrame()
	}
	e := &r.e
	e.beginTextMeasurements()
	e.gtx, e.m, e.store = gtx, gtx.Metric, st

	if gtx.Enabled() {
		r.source = gtx.Source
	}
	if live {
		r.beginTimers()
	}
	cx := Context{root: r}
	tree := r.renderTree(&cx)
	cx.prepareActionBindings(tree)
	st.assignKeys(tree, 1)
	r.prepareKeys(tree, nil, false)
	r.mainTree = tree
	r.prepareLayers(&cx)
	r.callbacks = false
	r.requestedFocus = nil
	if live {
		r.dispatchLayers(&cx)
		r.dispatchHover(gtx)
		r.blur(gtx)
		r.dispatchTab(&cx)
		r.dispatchKeys(gtx)
		r.dispatch(gtx)
		flushClipboard(gtx)
	}
	if r.callbacks {
		r.beginTimers()
		cx.shortcuts = nil
		cx.actions = nil
		cx.globalActions = nil
		cx.layers = nil
		tree = r.renderTree(&cx)
		cx.prepareActionBindings(tree)
	}

	cx.prepareScopedActions()
	for _, s := range cx.shortcuts {
		for {
			ev, ok := gtx.Event(key.Filter{Name: s.name, Required: s.mods})
			if !ok {
				break
			}
			if k, ok := ev.(key.Event); ok && k.State == key.Press {
				r.keyboardModality = true
				core.Call(gtx, s.fn)
			}
		}
	}

	st.assignKeys(tree, 1)
	r.prepareKeys(tree, nil, false)
	r.mainTree = tree
	r.prepareLayers(&cx)
	priorFocus := r.focusedTag()
	if r.requestedFocus != nil {
		priorFocus = r.requestedFocus
	}
	base := textStyle{color: &theme.Text, size: theme.BodySize}
	max := gtx.Constraints.Max
	if r.fill {
		tree.forceW, tree.forceH = max.X, max.Y
		if tree.style.bg == nil && tree.style.gradient == nil {
			tree.style.bg = &theme.Bg
			tree.style.BgGradient(theme.BgGradient)
		}
	}
	e.layout(tree, max.X, max.Y, base)
	e.place(tree)
	e.origin, e.visible = image.Point{}, image.Rectangle{Max: max}
	e.prepareTextAtlas(tree)
	// Everything is painted inside an area that sees every press, so a click
	// on empty space can take focus away from inputs and selected text.
	area := clip.Rect{Max: max}.Push(gtx.Ops)
	event.Op(gtx.Ops, &r.bg)
	e.anchors = map[string]image.Rectangle{}
	e.blockInput, e.blockFocus = false, false
	for _, d := range cx.layers {
		e.blockInput = e.blockInput || (d.eligible && d.layer.modal)
		e.blockFocus = e.blockFocus || (d.eligible && d.layer.trap)
	}
	if e.blockInput {
		core.Role("el-inert").Add(gtx.Ops)
	}
	e.paint(tree)
	e.textAtlasActive = false
	area.Pop()
	e.blockInput, e.blockFocus = false, false
	r.paintLayers(&cx, base, priorFocus)
	r.finishTimers()
	if live {
		flushClipboard(gtx) // from shortcuts and widget callbacks during this frame
		st.sweep()
		r.cache.sweep(st.frame)
	}
	if r.fill {
		return core.D{Size: max}
	}
	return core.D{Size: tree.size}
}

// dispatch runs click handlers for input that arrived since the last frame,
// before Render, so the frame being drawn already shows their effect.
func (r *RootWidget) dispatch(gtx core.C) {
	r.pointerDispatch = true
	defer func() { r.pointerDispatch = false }()
	var focusTarget, editorTarget *elemState
	for _, st := range r.store.states {
		if st.pressable && !st.disabled && !st.blocked && st.frame == r.store.frame {
			for {
				ev, ok := gtx.Event(pointer.Filter{Target: &st.pressTag, Kinds: pointer.Press})
				if !ok {
					break
				}
				if ev, ok := ev.(pointer.Event); !ok || ev.Kind != pointer.Press {
					continue
				}
				r.keyboardModality = false
				if st.pressEditor {
					editorTarget = st
				} else {
					r.focusID, r.focusPending, r.focusFromPointer = st.pressFocus, true, true
				}
			}
		}
		if st.onScroll != nil && !st.disabled && !st.blocked && st.frame == r.store.frame {
			for {
				ev, ok := gtx.Event(st.onScroll.filter(&st.scrollTag, gtx.Metric))
				if !ok {
					break
				}
				if e, ok := ev.(pointer.Event); ok && e.Kind == pointer.Scroll {
					scale := gtx.Metric.PxPerDp
					if scale <= 0 {
						scale = 1
					}
					fn := st.onScroll.fn
					core.Call(gtx, func() { r.callbacks = true; fn(ScrollEvent{X: e.Scroll.X / scale, Y: e.Scroll.Y / scale}) })
				}
			}
		}
		if !st.clickable || st.disabled || st.blocked || st.frame != r.store.frame {
			continue
		}
		if st.onContextMenu != nil {
			for {
				ev, ok := gtx.Event(pointer.Filter{Target: &st.contextTag, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
				if !ok {
					break
				}
				e, ok := ev.(pointer.Event)
				if !ok {
					continue
				}
				if e.Kind == pointer.Press && e.Source == pointer.Mouse && e.Buttons == st.contextButton {
					r.clickModifiers = e.Modifiers
					core.Call(gtx, func() { r.callbacks = true; st.onContextMenu() })
					r.clickModifiers = 0
				}
				st.trackTouchHold(gtx, e)
			}
			// A finger held still stands in for the secondary button.
			if h := &st.touchHold; h.active && st.contextButton == pointer.ButtonSecondary {
				if gtx.Now.Before(h.due) {
					gtx.Execute(op.InvalidateCmd{At: h.due})
				} else {
					h.active, st.holdFired = false, true
					core.Call(gtx, func() { r.callbacks = true; st.onContextMenu() })
				}
			}
		}
		for {
			ev, ok := st.click.Update(gtx.Source)
			if !ok {
				break
			}
			if ev.Kind == gesture.KindPress {
				r.keyboardModality = false
				if st.focusable {
					st.pointerFocus = true
					focusTarget = st
				}
			}
			if ev.Kind != gesture.KindClick {
				continue
			}
			if st.holdFired {
				st.holdFired = false
				continue // lifting the finger after a long press
			}
			r.clickModifiers = ev.Modifiers
			if ev.NumClicks >= 2 && st.onDoubleClick != nil {
				core.Call(gtx, func() { r.callbacks = true; st.onDoubleClick() })
			}
			if st.onClick != nil {
				core.Call(gtx, func() { r.callbacks = true; st.onClick() })
			}
			r.clickModifiers = 0
		}
		for st.onDrag != nil {
			var ev pointer.Event
			var ok bool
			if st.dragAccept != nil {
				ev, ok = st.conditionalDrag.update(gtx.Metric, gtx.Source, st.dragAccept)
			} else {
				ev, ok = st.drag.Update(gtx.Metric, gtx.Source, gesture.Both)
			}
			if !ok {
				break
			}
			kind := DragMove
			switch ev.Kind {
			case pointer.Press:
				kind = DragStart
				r.keyboardModality = false
				if st.focusable {
					st.pointerFocus = true
					focusTarget = st
				}
			case pointer.Release, pointer.Cancel:
				kind = DragEnd
			}
			px := r.e.m.PxPerDp
			if px == 0 {
				px = 1
			}
			de := DragEvent{Kind: kind, Canceled: ev.Kind == pointer.Cancel, X: ev.Position.X / px, Y: ev.Position.Y / px, W: float32(st.size.X) / px, H: float32(st.size.Y) / px}
			fn := st.onDrag
			core.Call(gtx, func() { r.callbacks = true; fn(de) })
		}
	}
	if focusTarget != nil {
		r.requestedFocus = focusTarget
		gtx.Execute(key.FocusCmd{Tag: focusTarget})
	}
	if editorTarget != nil {
		gtx.Execute(key.FocusCmd{Tag: &editorTarget.editor})
	}
}

// blur clears keyboard focus on a press that no focusable element claims. The
// root's area encloses everything, so it sees every press; an input or text
// that was pressed asks for focus later in the frame, and that request wins.
// Listen for the whole pointer sequence: if a disabled control leaves only
// this background handler, an unhandled release must not strand pointer capture.
func (r *RootWidget) blur(gtx core.C) {
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &r.bg, Kinds: pointer.Press | pointer.Release})
		if !ok {
			break
		}
		if e, ok := ev.(pointer.Event); ok && e.Kind == pointer.Press {
			gtx.Execute(key.FocusCmd{})
		}
	}
}
