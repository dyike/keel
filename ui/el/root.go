package el

import (
	"gioui.org/gesture"
	"gioui.org/io/key"

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
	view  View
	fill  bool
	store *store
	e     engine
}

// Root makes v the whole content of a window: it fills the window, with the
// theme background, and the window adds no padding or scrolling of its own.
//
//	window.Open(window.Options{Title: "Orders", Content: el.Root(&Orders{})})
func Root(v View) *RootWidget { return &RootWidget{view: v, fill: true, store: newStore()} }

// Embed renders v as an ordinary widget sized to its content, e.g. inside a
// ui/layout.Column during migration.
func Embed(v View) *RootWidget { return &RootWidget{view: v, store: newStore()} }

// FillsWindow tells ui/window to give the root the whole window.
func (r *RootWidget) FillsWindow() bool { return r.fill }

func (r *RootWidget) Layout(gtx core.C) core.D {
	st := r.store
	st.frame++
	e := &r.e
	e.gtx, e.m, e.store = gtx, gtx.Metric, st

	r.dispatch(gtx)
	var cx Context
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

	assignKeys(tree, 1)
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
	e.paint(tree)
	st.sweep()
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
