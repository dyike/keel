package el

import (
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
)

// KeyState distinguishes key presses from releases.
type KeyState uint8

const (
	KeyPress KeyState = iota
	KeyRelease
)

// KeyEvent keeps the el API independent of the Gio event structure.
type KeyEvent struct {
	Name      string
	Modifiers key.Modifiers
	State     KeyState
}

const allKeyModifiers = key.ModCtrl | key.ModCommand | key.ModShift | key.ModAlt | key.ModSuper

// Focusable adds the element to the native Tab order and focuses it on press.
// Input and TextArea already manage their native editor focus.
func (s *Styled[T]) Focusable(on bool) *T {
	s.n.focusSet = true
	s.n.focusable = on
	return s.self
}

// OnKey handles keys from a focused Focusable element, bubbling through its
// ancestors. Return true to stop bubbling and suppress default activation.
// Tab remains platform focus navigation; use Context.Shortcut for global keys.
func (s *Styled[T]) OnKey(fn func(KeyEvent) bool) *T { s.n.onKey = fn; return s.self }

// FocusStyle sets a visual keyboard/programmatic focus style, like Hover.
// Pointer focus on non-input elements does not show it. Inputs always show it.
// It does not change layout.
// The default focus style is a 2dp Primary border inside the element bounds.
func (s *Styled[T]) FocusStyle(fn func(*Style)) *T { s.n.focus = fn; return s.self }

func (s *Styled[T]) Disabled(v bool) *T               { s.n.disabled = v; return s.self }
func (s *Styled[T]) DisabledStyle(fn func(*Style)) *T { s.n.disabledStyle = fn; return s.self }

// FocusWithin reports whether the element with id, or anything inside it,
// had focus in the last painted frame. Tooltips use it for keyboard focus.
func (cx *Context) FocusWithin(id string) bool {
	if id == "" {
		return false
	}
	r := cx.root
	var find func(*Node) *Node
	find = func(n *Node) *Node {
		if n == nil || n.style.hidden {
			return nil
		}
		if n.id == id {
			return n
		}
		for _, c := range n.children {
			if f := find(c.node()); f != nil {
				return f
			}
		}
		return nil
	}
	n := find(r.mainTree)
	for _, st := range r.layers {
		if n == nil {
			n = find(st.tree)
		}
	}
	var focused func(*Node) bool
	focused = func(n *Node) bool {
		if st := r.store.states[n.key]; st != nil && !st.disabled && (r.source.Focused(st) || r.source.Focused(&st.editor)) {
			return true
		}
		for _, c := range n.children {
			if focused(c.node()) {
				return true
			}
		}
		return false
	}
	return n != nil && focused(n)
}

// Enabled reports whether the last declared element with id accepts input,
// including ancestor disabled state and modal blocking. Missing IDs are false.
// During Render this describes the previous declaration, like FocusWithin.
func (cx *Context) Enabled(id string) bool {
	if id == "" {
		return false
	}
	for _, st := range cx.root.store.states {
		if st.id == id {
			return !st.disabled && !st.blocked
		}
	}
	return false
}

func (cx *Context) Focused(id string) bool {
	if id == "" {
		return false
	}
	for _, st := range cx.root.store.states {
		if st.id == id && !st.disabled && (cx.root.source.Focused(st) || cx.root.source.Focused(&st.editor)) {
			return true
		}
	}
	return false
}

// Focus requests focus by ID in this root. It is applied after painting.
// The first visible Focusable element or input with that ID wins. An empty ID clears
// focus; missing, hidden or off-screen targets leave the current focus alone.
// Call from Render or its callbacks; IDs should be unique within a root.
func (cx *Context) Focus(id string) {
	cx.root.focusID = id
	cx.root.focusPending = true
	cx.root.focusFromPointer = cx.root.pointerDispatch
}

func (r *RootWidget) prepareKeys(n *Node, parent *elemState, disabled bool) {
	n.effectiveDisabled = disabled || n.disabled || n.style.hidden || n.style.revealSet && n.style.reveal <= 0
	n.disabledRoot = n.effectiveDisabled && !disabled
	st := r.store.states[n.key]
	if n.id != "" || n.isFocusable() || n.onKey != nil || n.input != nil || n.interactive() || n.style.scrollX || n.style.scrollY {
		st = r.store.get(n.key)
	}
	if st != nil && r.e.gtx.Enabled() {
		st.id = n.id
		st.scrollableX, st.scrollableY = n.style.scrollX, n.style.scrollY
		st.disabled = n.effectiveDisabled
		st.focusable = n.isFocusable() && n.input == nil
		st.onKey, st.keyParent = n.onKey, parent
		st.onContextMenu = n.onContextMenu
		st.onClick, st.onDoubleClick, st.onDrag = n.onClick, n.onDoubleClick, n.onDrag
		if st.disabled {
			st.pressedKey = ""
			st.click = gesture.Click{}
			st.drag = gesture.Drag{}
			st.scrollbarX, st.scrollbarY = scrollbarState{}, scrollbarState{}
			st.fresh = true
			if r.e.gtx.Focused(st) || r.e.gtx.Focused(&st.editor) {
				r.e.gtx.Execute(key.FocusCmd{})
			}
		}
	}
	if st != nil && (n.isFocusable() || n.onKey != nil || n.style.scrollX || n.style.scrollY) {
		parent = st
	}
	for _, c := range n.children {
		r.prepareKeys(c.node(), parent, n.effectiveDisabled)
	}
}
func (r *RootWidget) applyFocus(gtx core.C, n *Node) {
	if !r.focusPending || !gtx.Enabled() {
		return
	}
	r.focusPending = false
	if r.focusID == "" {
		gtx.Execute(key.FocusCmd{})
		return
	}
	var find func(*Node) bool
	find = func(n *Node) bool {
		if n.effectiveDisabled {
			return false
		}
		if n.id == r.focusID && (n.isFocusable() || n.input != nil) {
			st := r.store.states[n.key]
			if st != nil && st.keyFrame == r.store.frame {
				if n.input != nil {
					gtx.Execute(key.FocusCmd{Tag: &st.editor})
				} else {
					st.pointerFocus = r.focusFromPointer
					gtx.Execute(key.FocusCmd{Tag: st})
				}
				return true
			}
		}
		for _, c := range n.children {
			if find(c.node()) {
				return true
			}
		}
		return false
	}
	find(n)
}
func (r *RootWidget) dispatchKeys(gtx core.C) {
	for _, st := range r.store.states {
		// Clear the previous pointer origin before dispatch, never in prepareKeys:
		// a pointer FocusCmd may still be pending during the second preparation pass.
		if !gtx.Focused(st) {
			st.pointerFocus = false
		}
		if !st.focusable || st.disabled || st.blocked || st.frame != r.store.frame || st.keyFrame+1 != r.store.frame {
			continue
		}
		for {
			ev, ok := gtx.Event(focusFilters(st)...)
			if !ok {
				break
			}
			switch ev := ev.(type) {
			case key.FocusEvent:
				st.pressedKey = ""
			case key.Event:
				if !gtx.Focused(st) {
					continue
				}
				st.pointerFocus = false
				handled := false
				for node := st; node != nil; node = node.keyParent {
					if node.onKey != nil {
						core.Call(gtx, func() {
							r.callbacks = true
							state := KeyPress
							if ev.State == key.Release {
								state = KeyRelease
							}
							handled = node.onKey(KeyEvent{Name: string(ev.Name), Modifiers: ev.Modifiers, State: state})
						})
					}
					if handled {
						break
					}
				}
				// Explicit handlers get first refusal across the whole focus chain.
				// Otherwise a nested viewport consumes navigation meant for its owner.
				if !handled {
					for node := st; node != nil; node = node.keyParent {
						if !node.disabled && !node.blocked {
							handled = node.scrollKey(gtx, ev)
						}
						if handled {
							if ev.State == key.Press {
								r.callbacks = true
							}
							break
						}
					}
				}
				if handled {
					st.pressedKey = ""
					continue
				}
				activation := ev.Name == key.NameSpace || ev.Name == key.NameReturn || ev.Name == key.NameEnter
				if !activation || ev.Modifiers != 0 {
					continue
				}
				if ev.State == key.Press {
					st.pressedKey = ev.Name
				} else if st.pressedKey == ev.Name {
					st.pressedKey = ""
					core.Call(gtx, func() {
						if st.onClick != nil {
							r.callbacks = true
							st.onClick()
						}
					})
				}
			}
		}
	}
}

func focusFilters(st *elemState) []event.Filter {
	filters := []event.Filter{key.FocusFilter{Target: st}}
	scrolling := false
	for p := st; p != nil; p = p.keyParent {
		scrolling = scrolling || p.scrollableX || p.scrollableY
		if p.onKey != nil {
			return append(filters, key.Filter{Focus: st, Optional: allKeyModifiers})
		}
	}
	if scrolling {
		for _, name := range []key.Name{key.NameLeftArrow, key.NameRightArrow, key.NameUpArrow, key.NameDownArrow, key.NamePageUp, key.NamePageDown, key.NameHome, key.NameEnd} {
			filters = append(filters, key.Filter{Focus: st, Name: name, Optional: key.ModShift})
		}
	}
	for _, name := range []key.Name{key.NameSpace, key.NameReturn, key.NameEnter} {
		filters = append(filters, key.Filter{Focus: st, Name: name})
	}
	return filters
}
