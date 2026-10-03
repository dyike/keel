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
// This does not dispatch menu clicks; Menu.ActionItem owns its callback.
func (cx *Context) ActionAt(targetID, action string, fn func()) {
	cx.actions = append(cx.actions, scopedAction{targetID, action, fn})
}

type actionBindingTarget struct {
	contexts []string
	depth    int
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
	var collect func(*Node, []string, int)
	collect = func(n *Node, contexts []string, depth int) {
		if n.style.hidden || n.disabled {
			return
		}
		if n.keyContext != "" {
			contexts = append([]string{n.keyContext}, contexts...)
		}
		if n.id != "" {
			cx.bindingTargets[n.id] = actionBindingTarget{slices.Clone(contexts), depth}
		}
		for _, c := range n.children {
			collect(c.node(), contexts, depth+1)
		}
	}
	for _, root := range roots {
		collect(root, nil, 0)
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
