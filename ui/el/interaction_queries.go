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
		if cx.root.source.Focused(st) {
			cx.queryFocus[st.id] = append(cx.queryFocus[st.id], st)
		}
	}
}

func (cx *Context) prepareInteractionNodes() {
	cx.refreshInteractionQueries()
	if cx.queryNodes != nil {
		return
	}
	cx.queryNodes = map[string]*Node{}
	var visit func(*Node)
	visit = func(n *Node) {
		if n == nil || n.style.hidden {
			return
		}
		if n.id != "" && cx.queryNodes[n.id] == nil {
			cx.queryNodes[n.id] = n
		}
		for _, child := range n.children {
			visit(child.node())
		}
	}
	visit(cx.root.mainTree)
	for _, layer := range cx.root.layers {
		visit(layer.tree)
	}
}
