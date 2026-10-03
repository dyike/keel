package el

// contentBottom resolves after measurement, so the first frame and changing
// heights use current geometry. place only traverses descendants, never parents.
func (e *engine) contentBottom(n *Node) int {
	target := n.style.contentBottom
	if target == nil || target == n {
		return n.size.Y
	}
	e.place(n)
	var find func(*Node, int) (int, bool)
	find = func(current *Node, y int) (int, bool) {
		for _, child := range current.children {
			c := child.node()
			if c.style.hidden || c.style.absolute {
				continue
			}
			bottom := y + c.pos.Y + c.size.Y
			if c == target {
				return bottom, true
			}
			if result, ok := find(c, y+c.pos.Y); ok {
				return result, true
			}
		}
		return 0, false
	}
	if bottom, ok := find(n, 0); ok {
		return max(0, min(n.size.Y, bottom))
	}
	return n.size.Y
}
func (e *engine) contentBottomExtents(kids []*Node) (above, below int) {
	for _, c := range kids {
		_, top, _, bottom := e.edges(c.style.margin)
		anchor := e.contentBottom(c)
		above = max(above, top+anchor)
		below = max(below, bottom+c.size.Y-anchor)
	}
	return
}
