package el

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"sort"
)

type tabTarget struct {
	tag   event.Tag
	state *elemState
	index int
}

func (r *RootWidget) tabTargets(cx *Context) ([]tabTarget, bool) {
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
	return r.collectTabTargets(roots)
}

func (r *RootWidget) collectTabTargets(roots []*Node) ([]tabTarget, bool) {
	var targets []tabTarget
	configured := false
	var visit func(*Node)
	visit = func(n *Node) {
		configured = configured || n.tabConfigured
		if n.effectiveDisabled || n.style.hidden {
			return
		}
		if s := r.store.states[n.key]; s != nil && !s.blocked && s.keyFrame+1 >= r.store.frame && s.keyFrame > 0 && !n.tabSkip && n.tabIndex >= 0 {
			if n.input != nil {
				targets = append(targets, tabTarget{&s.editor, s, n.tabIndex})
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

func (r *RootWidget) dispatchTab(cx *Context) {
	targets, configured := r.tabTargets(cx)
	if !configured {
		return
	}
	for {
		ev, ok := r.e.gtx.Event(key.Filter{Name: key.NameTab, Optional: key.ModShift})
		if !ok {
			break
		}
		k, ok := ev.(key.Event)
		if !ok || k.State != key.Press {
			continue
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
		t.state.pointerFocus = false
		r.e.gtx.Execute(key.FocusCmd{Tag: t.tag})
	}
}
