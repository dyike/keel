package el

// A view often queries hover/focus for every row. Index the last declared
// state once per render pass instead of traversing it separately per ID.
// Reset after input dispatch and before rerendering so callbacks see fresh
// pointer state, and retain duplicate-ID behavior (any hovered/focused state;
// the first declared subtree for FocusWithin).
func (cx *Context) resetInteractionQueries() {
	cx.root.queryRevision++
	cx.refreshInteractionQueries()
}

func (cx *Context) refreshInteractionQueries() {
	if cx.queryRevision != cx.root.queryRevision {
		cx.queryRevision = cx.root.queryRevision
		cx.queryHover, cx.queryFocus, cx.queryNodes = nil, nil, nil
		cx.queryNodesPrepared = false
	}
}

func (cx *Context) prepareInteractionQueries() {
	cx.refreshInteractionQueries()
	if cx.queryHover != nil {
		return
	}
	cx.queryHover, cx.queryFocus = map[string][]*elemState{}, map[string][]*elemState{}
	for _, st := range cx.root.store.states {
		if st.id == "" {
			continue
		}
		if st.hovered {
			cx.queryHover[st.id] = append(cx.queryHover[st.id], st)
		}
		if cx.root.source.Focused(st) || (st.editor != nil && cx.root.source.Focused(st.editor)) {
			cx.queryFocus[st.id] = append(cx.queryFocus[st.id], st)
		}
	}
}

type focusWithinState struct {
	focused, visible bool
}

func (cx *Context) prepareInteractionNodes() {
	cx.refreshInteractionQueries()
	if cx.queryNodesPrepared {
		return
	}
	cx.queryNodesPrepared = true
	if cx.root.focusedTag() == nil {
		return
	}
	cx.queryNodes = map[string]focusWithinState{}
	var visit func(*Node) focusWithinState
	visit = func(n *Node) focusWithinState {
		if n == nil || n.style.hidden {
			return focusWithinState{}
		}
		_, seen := cx.queryNodes[n.id]
		if n.id != "" && !seen {
			cx.queryNodes[n.id] = focusWithinState{}
		}
		state := focusWithinState{}
		if st := cx.root.store.of(n); st != nil && !st.disabled {
			state.focused = cx.root.source.Focused(st) || st.editor != nil && cx.root.source.Focused(st.editor)
			state.visible = !st.pointerFocus && cx.root.source.Focused(st) || st.inputRing && st.editor != nil && cx.root.source.Focused(st.editor)
		}
		for _, child := range n.children {
			descendant := visit(child.node())
			state.focused = state.focused || descendant.focused
			state.visible = state.visible || descendant.visible
		}
		if n.id != "" && !seen {
			cx.queryNodes[n.id] = state
		}
		return state
	}
	visit(cx.root.mainTree)
	if cx.layers != nil {
		for _, declaration := range cx.layers {
			visit(declaration.layer.content.node())
		}
	} else {
		for _, layer := range cx.root.layers {
			visit(layer.tree)
		}
	}
}
