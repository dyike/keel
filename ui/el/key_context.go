package el

import (
	"cmp"
	"gioui.org/io/key"
	"slices"

	"github.com/dyike/keel/ui/core"
)

// KeyContext names a keymap scope inherited by descendants. Nested scopes
// override outer ones per action; an empty name adds no scope. Pair it with
// core.BindIn and Context.ActionAt for focused command handling.
func (s *Styled[T]) KeyContext(name string) *T { s.n.keyContext = name; return s.self }

type actionKeyHint struct {
	action, target string
	build          func(string) Element
}

// KeyHint builds a hint for the first chord resolved at targetID. Resolution
// happens after the current element tree is built, before measuring it, so
// initial-frame hints see the target's current KeyContext ancestry. Missing,
// hidden or disabled targets and unbound actions render no content. The builder
// must return display-only content, without declaring overlays or shortcuts.
func KeyHint(action, targetID string, build func(chord string) Element) *DivEl {
	e := Div()
	e.n.keyHint = &actionKeyHint{action: action, target: targetID, build: build}
	return e
}

type scopedAction struct {
	target, action string
	fn             func()
}

// ActionAt handles an action while focus is within targetID, resolving chords
// through its current KeyContext ancestry and then core.Bind. Declare it every
// frame. Do not also register the same action through the global Action method.
// The innermost focused target wins for repeated actions or chords; ties use
// declaration order. An inner empty binding also suppresses outer handlers.
// Perform runs the handler as if from targetID (see Perform).
func (cx *Context) ActionAt(targetID, action string, fn func()) {
	cx.actions = append(cx.actions, scopedAction{targetID, action, fn})
}

type actionBindingTarget struct {
	contexts  []string
	depth     int
	ancestors []string // IDs of enclosing elements, inner first
}

func (cx *Context) prepareScopedActions() {
	actions := slices.Clone(cx.actions)
	slices.SortStableFunc(actions, func(a, b scopedAction) int {
		return cmp.Compare(cx.bindingTargets[b.target].depth, cx.bindingTargets[a.target].depth)
	})
	seenActions := map[string]bool{}
	seenKeys := map[struct {
		name key.Name
		mods key.Modifiers
	}]bool{}
	for _, a := range actions {
		target, ok := cx.bindingTargets[a.target]
		if !ok || seenActions[a.action] || !cx.FocusWithin(a.target) {
			continue
		}
		seenActions[a.action] = true
		for _, chord := range core.BindingsIn(a.action, target.contexts...) {
			name, mods, _ := core.ParseShortcut(chord)
			k := struct {
				name key.Name
				mods key.Modifiers
			}{name, mods}
			if !seenKeys[k] {
				cx.Shortcut(chord, a.fn)
				seenKeys[k] = true
			}
		}
	}
}

func (cx *Context) prepareActionBindings(tree *Node) {
	// Collect every target before resolving any hint. A target may appear after
	// the menu or in another overlay's declaration in this same frame.
	cx.bindingTargets = map[string]actionBindingTarget{}
	roots := []*Node{tree}
	for _, d := range cx.layers {
		if d.layer.content != nil {
			roots = append(roots, d.layer.content.node())
		}
	}
	var collect func(*Node, []string, []string, int)
	collect = func(n *Node, contexts, ancestors []string, depth int) {
		if n.style.hidden || n.disabled {
			return
		}
		if n.keyContext != "" {
			contexts = append([]string{n.keyContext}, contexts...)
		}
		if n.id != "" {
			cx.bindingTargets[n.id] = actionBindingTarget{slices.Clone(contexts), depth, slices.Clone(ancestors)}
			ancestors = append([]string{n.id}, ancestors...)
		}
		for _, c := range n.children {
			collect(c.node(), contexts, ancestors, depth+1)
		}
	}
	for _, root := range roots {
		collect(root, nil, nil, 0)
	}
	var resolve func(*Node)
	resolve = func(n *Node) {
		if hint := n.keyHint; hint != nil {
			n.children = nil
			if target, ok := cx.bindingTargets[hint.target]; ok && hint.build != nil {
				if chords := core.BindingsIn(hint.action, target.contexts...); len(chords) > 0 {
					if content := hint.build(chords[0]); content != nil {
						n.children = []Element{content}
					}
				}
			}
			return
		}
		for _, c := range n.children {
			resolve(c.node())
		}
	}
	for _, root := range roots {
		resolve(root)
	}
}

// Perform runs the handler a key press bound to action would run with focus
// at targetID: the innermost ActionAt declared on targetID or an element
// enclosing it, else a global Action handler. It works without any key
// bound, which is what menus and command palettes need. Call it from a
// callback; it sees this frame's declarations and reports whether a
// handler ran. Hidden or disabled targets run nothing.
func (cx *Context) Perform(targetID, action string) bool {
	if target, ok := cx.bindingTargets[targetID]; ok {
		scope := append([]string{targetID}, target.ancestors...)
		for _, id := range scope { // inner first
			for _, a := range cx.actions {
				if a.action == action && a.target == id && a.fn != nil {
					core.Call(cx.root.e.gtx, a.fn)
					return true
				}
			}
		}
	} else if targetID != "" {
		return false
	}
	if fn := cx.globalActions[action]; fn != nil {
		core.Call(cx.root.e.gtx, fn)
		return true
	}
	return false
}
