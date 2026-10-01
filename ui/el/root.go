package el

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
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

// Context is passed to Render.
type Context struct {
	shortcuts []viewShortcut
	root      *RootWidget
}

// Cache returns the element built for key, calling build only when key was
// not used in the previous frame. While the width it is given stays the same,
// the element's layout is reused too, so long, mostly unchanging content (a
// chat history, a rendered document) costs little per frame. key must be
// comparable and change whenever the element would look different; the
// element must not depend on anything else. Entries unused for a frame are
// dropped. Applying a theme also rebuilds cached elements.
func (cx *Context) Cache(key any, build func() Element) Element {
	c := &cx.root.cache
	if c.themeRevision != theme.Revision() {
		clear(c.entries)
		c.themeRevision = theme.Revision()
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
	entries       map[any]*cacheEntry
	themeRevision uint64
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

// RootWidget renders a View as a core.Widget.
type RootWidget struct {
	bg    int // tag of the area under everything; see blur
	view  View
	fill  bool
	store *store
	cache elementCache
	e     engine
}

// Root makes v the whole content of a window: it fills the window, with the
// theme background, and the window adds no padding or scrolling of its own.
//
//	window.Open(window.Options{Title: "Orders", Content: el.Root(&Orders{})})
func Root(v View) *RootWidget {
	return &RootWidget{view: v, fill: true, store: newStore(), cache: elementCache{entries: map[any]*cacheEntry{}}}
}

// Embed renders v as an ordinary widget sized to its content, e.g. inside a
// ui/layout.Column during migration.
func Embed(v View) *RootWidget {
	return &RootWidget{view: v, store: newStore(), cache: elementCache{entries: map[any]*cacheEntry{}}}
}

// FillsWindow tells ui/window to give the root the whole window.
func (r *RootWidget) FillsWindow() bool { return r.fill }

func (r *RootWidget) Layout(gtx core.C) core.D {
	st := r.store
	st.frame++
	e := &r.e
	e.gtx, e.m, e.store = gtx, gtx.Metric, st

	r.blur(gtx)
	r.dispatch(gtx)
	flushClipboard(gtx)
	cx := Context{root: r}
	tree := r.view.Render(&cx).node()
	for _, s := range cx.shortcuts {
		for {
			ev, ok := gtx.Event(key.Filter{Name: s.name, Required: s.mods})
			if !ok {
				break
			}
			if k, ok := ev.(key.Event); ok && k.State == key.Press {
				core.Call(gtx, s.fn)
			}
		}
	}

	st.assignKeys(tree, 1)
	base := textStyle{color: &theme.Text, size: theme.BodySize}
	max := gtx.Constraints.Max
	if r.fill {
		tree.forceW, tree.forceH = max.X, max.Y
		if tree.style.bg == nil {
			tree.style.bg = &theme.Bg
		}
	}
	e.layout(tree, max.X, max.Y, base)
	e.place(tree)
	e.origin, e.visible = image.Point{}, image.Rectangle{Max: max}
	// Everything is painted inside an area that sees every press, so a click
	// on empty space can take focus away from inputs and selected text.
	area := clip.Rect{Max: max}.Push(gtx.Ops)
	event.Op(gtx.Ops, &r.bg)
	e.paint(tree)
	area.Pop()
	flushClipboard(gtx) // from shortcuts and widget callbacks during this frame
	st.sweep()
	r.cache.sweep(st.frame)
	if r.fill {
		return core.D{Size: max}
	}
	return core.D{Size: tree.size}
}

// dispatch runs click handlers for input that arrived since the last frame,
// before Render, so the frame being drawn already shows their effect.
func (r *RootWidget) dispatch(gtx core.C) {
	for _, st := range r.store.states {
		if !st.clickable {
			continue
		}
		for {
			ev, ok := st.click.Update(gtx.Source)
			if !ok {
				break
			}
			if ev.Kind != gesture.KindClick {
				continue
			}
			if ev.NumClicks >= 2 && st.onDoubleClick != nil {
				core.Call(gtx, st.onDoubleClick)
			}
			if st.onClick != nil {
				core.Call(gtx, st.onClick)
			}
		}
	}
}

// blur clears keyboard focus on a press that no focusable element claims. The
// root's area encloses everything, so it sees every press; an input or text
// that was pressed asks for focus later in the frame, and that request wins.
func (r *RootWidget) blur(gtx core.C) {
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &r.bg, Kinds: pointer.Press})
		if !ok {
			break
		}
		if e, ok := ev.(pointer.Event); ok && e.Kind == pointer.Press {
			gtx.Execute(key.FocusCmd{})
		}
	}
}
