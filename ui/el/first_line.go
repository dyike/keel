package el

// firstLine returns the center of n's first line of text, from the top of
// its border box: where text and input ink is centered, so a label beside a
// control lines up with the control's text whatever frames, rows or padding
// sit around it. Containers take their first in-flow descendant with text.
// Like contentBottom it resolves after measurement and only reads descendants.
func (e *engine) firstLine(n *Node) (int, bool) {
	bw := e.dp(n.style.borderWidth)
	_, pt, _, pb := e.edges(n.style.pad)
	top := bw + pt
	switch {
	case n.isText:
		return top + e.textLineBox(n, scriptOf(n.text))/2, true
	case n.input != nil:
		inner := n.size.Y - top - pb - bw
		if !n.input.multiline && n.input.line > 0 && inner > n.input.line {
			return top + inner/2, true
		}
		return top + n.input.line/2, true
	}
	e.place(n)
	for _, child := range n.children {
		c := child.node()
		if c.style.hidden || c.style.absolute {
			continue
		}
		if y, ok := e.firstLine(c); ok {
			return c.pos.Y + y, true
		}
	}
	return 0, false
}

// firstLineExtents is how far a FirstLine row reaches above and below the
// shared line: children without text center on it, like icons.
func (e *engine) firstLineExtents(kids []*Node) (above, below int) {
	for _, c := range kids {
		_, top, _, bottom := e.edges(c.style.margin)
		line := e.firstLineOf(c)
		above = max(above, top+line)
		below = max(below, bottom+c.size.Y-line)
	}
	return
}

// firstLineOf is c's first line, or its middle when it has no text.
func (e *engine) firstLineOf(c *Node) int {
	if y, ok := e.firstLine(c); ok {
		return y
	}
	return c.size.Y / 2
}
