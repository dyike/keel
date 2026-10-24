package el

import (
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"slices"
	"sort"
)

type tabTarget struct {
	tag   event.Tag
	state *elemState
	index int
}

func (r *RootWidget) tabRoots(cx *Context) []*Node {
	var roots []*Node
	trap := -1
	for i, d := range cx.layers {
		if d.eligible && (d.layer.trap || d.layer.modal) {
			trap = i
		}
	}
	if trap < 0 {
		roots = append(roots, r.mainTree)
	}
	for i, d := range cx.layers {
		if d.eligible && i >= trap {
			roots = append(roots, d.layer.content.node())
		}
	}
	return roots
}

func (r *RootWidget) tabTargets(roots []*Node) ([]tabTarget, bool) {
	if scope := r.focusedContainerTrap(roots); scope != nil {
		targets, _ := r.collectTabTargets([]*Node{scope})
		return targets, true
	}
	return r.collectTabTargets(roots)
}

func (r *RootWidget) collectTabTargets(roots []*Node) ([]tabTarget, bool) {
	var targets []tabTarget
	configured := false
	var visit func(*Node)
	visit = func(n *Node) {
		configured = configured || n.tabConfigured || n.focusTrap
		if n.effectiveDisabled || n.style.hidden {
			return
		}
		if s := r.store.of(n); s != nil && !s.blocked && s.keyFrame+1 >= r.store.frame && s.keyFrame > 0 && !n.tabSkip && n.tabIndex >= 0 {
			if n.input != nil {
				targets = append(targets, tabTarget{s.editor, s, n.tabIndex})
			} else if n.isFocusable() {
				targets = append(targets, tabTarget{s, s, n.tabIndex})
			}
		}
		for _, c := range n.children {
			visit(c.node())
		}
	}
	for _, n := range roots {
		visit(n)
	}
	sort.SliceStable(targets, func(i, j int) bool { return targets[i].index < targets[j].index })
	return targets, configured
}

// tabConfigured reports what collectTabTargets would: whether any reachable
// node takes part in Keel's Tab order. It reads no state, so a frame without
// a Tab press does not pay for collecting targets.
func tabConfigured(n *Node) bool {
	if n.tabConfigured || n.focusTrap {
		return true
	}
	if n.effectiveDisabled || n.style.hidden {
		return false
	}
	for _, c := range n.children {
		if tabConfigured(c.node()) {
			return true
		}
	}
	return false
}

func (r *RootWidget) dispatchTab(cx *Context) {
	roots := r.tabRoots(cx)
	// A trap scope needs a reachable trap node, which tabConfigured counts.
	if !slices.ContainsFunc(roots, tabConfigured) {
		return
	}
	var targets []tabTarget
	collected := false
	for {
		ev, ok := r.e.gtx.Event(key.Filter{Name: key.NameTab, Optional: key.ModShift})
		if !ok {
			break
		}
		k, ok := ev.(key.Event)
		if !ok || k.State != key.Press {
			continue
		}
		if !collected {
			targets, _ = r.tabTargets(roots)
			collected = true
		}
		if len(targets) == 0 {
			r.e.gtx.Execute(key.FocusCmd{})
			continue
		}
		current := -1
		for i, t := range targets {
			if r.source.Focused(t.tag) {
				current = i
				break
			}
		}
		next := current + 1
		if k.Modifiers == key.ModShift {
			if current < 0 {
				next = len(targets) - 1
			} else {
				next = current - 1
			}
		}
		next = (next + len(targets)) % len(targets)
		t := targets[next]
		t.state.focusFrom(false)
		r.e.gtx.Execute(key.FocusCmd{Tag: t.tag})
	}
}

// Locate the closest trap ancestor of the focused node. Search only roots that
// are eligible under the active overlay, so a background trap cannot steal Tab.
func (r *RootWidget) focusedContainerTrap(roots []*Node) *Node {
	var visit func(*Node, *Node) (*Node, bool)
	visit = func(n, scope *Node) (*Node, bool) {
		if n == nil || n.effectiveDisabled || n.style.hidden {
			return nil, false
		}
		if n.focusTrap {
			scope = n
		}
		if state := r.store.of(n); state != nil && !state.blocked && (r.source.Focused(state) || (state.editor != nil && r.source.Focused(state.editor))) {
			return scope, true
		}
		for _, child := range n.children {
			if trap, found := visit(child.node(), scope); found {
				return trap, true
			}
		}
		return nil, false
	}
	for _, root := range roots {
		if trap, found := visit(root, nil); found {
			return trap
		}
	}
	return nil
}
