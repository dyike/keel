package el

// MatchText pairs descendant Text elements within a SharedLayout handoff by
// their relative paths (IDs when supplied, sibling positions otherwise). It
// adds a Morph content transition where none was declared. Put stable IDs on
// reordered labels. Explicit SharedLayout descendants start independent scopes.
// This opt-in does not match strings heuristically or retain old element trees.
func (s *Styled[T]) MatchText() *T {
	s.node().sharedText = true
	return s.self
}

type sharedTextIdentity struct {
	parent any
	path   stateKey
}

func (s *store) assignSharedKeys(n *Node, key stateKey, scope *sharedLayoutSpec, path stateKey) {
	if n.sharedAuto {
		n.sharedLayout, n.sharedAuto = nil, false
	}
	if n.contentAuto {
		n.content, n.contentAuto = nil, false
	}
	if n.sharedLayout != nil {
		scope = nil
	}
	if n.sharedText {
		if n.sharedLayout == nil {
			panic("el: MatchText requires SharedLayout")
		}
		scope, path = n.sharedLayout, 0
	} else if scope != nil && n.isText && n.shimmer == nil {
		n.sharedLayout = &sharedLayoutSpec{identity: sharedTextIdentity{scope.identity, path}, animation: scope.animation}
		n.sharedAuto = true
		if n.content == nil {
			n.content = &ContentTransition{Morph: true, Animation: scope.animation}
			n.contentAuto = true
		}
	}
	s.assignNodeKeys(n, key)
	for i, c := range n.children {
		cn := c.node()
		relative := stateKey(0)
		if scope != nil {
			relative = childKey(path, cn.id, i)
		}
		s.assignSharedKeys(cn, childKey(key, cn.id, i), scope, relative)
		n.sharedSubtree = n.sharedSubtree || cn.sharedSubtree
	}
}
