package el

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
)

// KeyEvent is a Gio key press or release, including its modifiers.
// Text input continues to use Input/TextArea rather than key events.
type KeyEvent = key.Event

const allKeyModifiers = key.ModCtrl | key.ModCommand | key.ModShift | key.ModAlt | key.ModSuper

// Focusable adds the element to the native Tab order and focuses it on press.
// Input and TextArea already manage their native editor focus.
func (s *Styled[T]) Focusable() *T { s.n.focusable = true; return s.self }

// OnKey handles keys from a focused Focusable element, bubbling through its
// ancestors. Return true to stop bubbling and suppress default activation.
// Tab remains platform focus navigation; use Context.Shortcut for global keys.
func (s *Styled[T]) OnKey(fn func(KeyEvent) bool) *T { s.n.onKey = fn; return s.self }

// Focus sets a visual focus style, like Hover. It does not change layout.
// The default focus style is a 2dp Primary border inside the element bounds.
func (s *Styled[T]) Focus(fn func(*Style)) *T { s.n.focus = fn; return s.self }

// Focus requests focus by ID in this root. It is applied after painting.
// The first visible Focusable element or input with that ID wins. An empty ID clears
// focus; missing, hidden or off-screen targets leave the current focus alone.
// Call from Render or its callbacks; IDs should be unique within a root.
func (cx *Context) Focus(id string) { cx.root.focusID = id; cx.root.focusPending = true }

func (r *RootWidget) prepareKeys(n *Node, parent *elemState) {
	if n.style.hidden {
		return
	}
	if st := r.store.states[n.key]; st != nil {
		st.focusable = false
		st.onKey, st.keyParent = nil, nil
	}
	if n.focusable || n.onKey != nil {
		st := r.store.get(n.key)
		st.focusable = n.focusable && n.input == nil
		st.onKey, st.keyParent = n.onKey, parent
		parent = st
	}
	for _, c := range n.children {
		r.prepareKeys(c.node(), parent)
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
		if n.style.hidden {
			return false
		}
		if n.id == r.focusID && (n.focusable || n.input != nil) {
			st := r.store.states[n.key]
			if st != nil && st.keyFrame == r.store.frame {
				if n.input != nil {
					gtx.Execute(key.FocusCmd{Tag: &st.editor})
				} else {
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
		if !st.focusable || st.keyFrame+1 != r.store.frame {
			continue
		}
		for {
			ev, ok := gtx.Event(key.FocusFilter{Target: st}, key.Filter{Focus: st, Optional: allKeyModifiers, Name: ""})
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
				handled := false
				for node := st; node != nil; node = node.keyParent {
					if node.onKey != nil {
						core.Call(gtx, func() { handled = node.onKey(ev) })
					}
					if handled {
						break
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
					core.Call(gtx, st.onClick)
				}
			}
		}
	}
}
