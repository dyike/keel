package el

import "image"

func (e *engine) layoutChildren(n *Node, innerW, innerH, limW, limH int) image.Point {
	if n.style.grid > 0 {
		return e.grid(n, innerW, limW, limH)
	}
	if n.style.wrap {
		return e.wrap(n, innerW, limW, limH)
	}
	return e.flex(n, innerW, innerH, limW, limH)
}

func (e *engine) wrap(n *Node, width, limW, limH int) image.Point {
	kids := inFlow(n)
	n.flowPositions = make([]image.Point, len(kids))
	limit := limW
	if width >= 0 {
		limit = width
	}
	gap := e.dp(n.style.gap)
	result := image.Point{}
	for first := 0; first < len(kids); {
		last, used := first, 0
		for last < len(kids) {
			c := kids[last]
			c.forceW, c.forceH = -1, -1
			e.layoutChild(c, n, limit, limH)
			l, _, r, _ := e.edges(c.style.margin)
			basis := c.size.X + l + r
			if c.style.grow > 0 {
				basis = max(c.style.minW.px(e.m, limit), 0) + l + r
			}
			if last > first && used+gap+basis > limit {
				break
			}
			if last > first {
				used += gap
			}
			used += basis
			last++
		}
		line := Node{style: Style{row: true, gap: n.style.gap, align: n.style.align, alignSet: n.style.alignSet, justify: n.style.justify}, textStyle: n.textStyle}
		for _, c := range kids[first:last] {
			line.children = append(line.children, c)
		}
		available := limit
		if width < 0 && n.style.wrapFit {
			available = min(limit, used)
		}
		if available >= inf {
			available = -1
		}
		size := e.flex(&line, available, -1, limit, limH)
		if available >= 0 {
			size.X = max(size.X, available)
		}
		if first > 0 {
			result.Y += gap
		}
		positions := e.rowPositions(&line, size)
		for j, p := range positions {
			n.flowPositions[first+j] = p.Add(image.Pt(0, result.Y))
		}
		result.X = max(result.X, size.X)
		result.Y += size.Y
		first = last
	}
	return result
}

// rowPositions aligns a line's already measured children, without recursively
// placing descendants; the ordinary place pass does that once afterwards.
func (e *engine) rowPositions(n *Node, size image.Point) []image.Point {
	kids := inFlow(n)
	gap := e.dp(n.style.gap)
	used := max(len(kids)-1, 0) * gap
	for _, c := range kids {
		l, _, r, _ := e.edges(c.style.margin)
		used += l + c.size.X + r
	}
	free := max(size.X-used, 0)
	cursor, extra := 0, 0
	switch n.style.justify {
	case Center:
		cursor = free / 2
	case End:
		cursor = free
	case SpaceBetween:
		if len(kids) > 1 {
			extra = free / (len(kids) - 1)
		}
	case SpaceAround:
		if len(kids) > 0 {
			extra = free / len(kids)
			cursor = extra / 2
		}
	}
	anchor := 0
	if n.style.align == ContentBottom {
		anchor, _ = e.contentBottomExtents(kids)
	}
	out := make([]image.Point, len(kids))
	for i, c := range kids {
		l, t, r, b := e.edges(c.style.margin)
		y := t
		switch n.style.align {
		case Center:
			y = t + (size.Y-c.size.Y-t-b)/2
		case End:
			y = size.Y - c.size.Y - b
		case ContentBottom:
			y = anchor - e.contentBottom(c)
		}
		out[i] = image.Pt(cursor+l, y)
		cursor += l + c.size.X + r + gap + extra
	}
	return out
}

func (e *engine) grid(n *Node, width, limW, limH int) image.Point {
	kids := inFlow(n)
	n.flowPositions = make([]image.Point, len(kids))
	if len(kids) == 0 {
		return image.Point{}
	}
	columns := n.style.grid
	gap := e.dp(n.style.gap)
	widths := make([]int, columns)
	starts, spans, rows := make([]int, len(kids)), make([]int, len(kids)), make([]int, len(kids))
	column, row := 0, 0
	for i, c := range kids {
		span := min(max(c.style.colSpan, 1), columns)
		if column+span > columns {
			column, row = 0, row+1
		}
		starts[i], spans[i], rows[i] = column, span, row
		column += span
	}
	trackWidth := func(start, span int) int {
		total := (span - 1) * gap
		for _, w := range widths[start : start+span] {
			total += w
		}
		return total
	}
	for i, c := range kids {
		l, _, r, _ := e.edges(c.style.margin)
		minimum := max(c.style.minW.px(e.m, limW), 0)
		if c.style.w.kind == dpLen || c.style.w.kind == spLen {
			minimum = max(minimum, c.style.w.px(e.m, limW))
		}
		if limW >= inf {
			c.forceW, c.forceH = -1, -1
			e.layoutChild(c, n, limW, limH)
			minimum = max(minimum, c.size.X)
		}
		// Distribute a spanning cell's minimum across its covered tracks.
		start, span := starts[i], spans[i]
		deficit := max(minimum+l+r-trackWidth(start, span), 0)
		for j := 0; j < span; j++ {
			share := deficit / (span - j)
			widths[start+j] += share
			deficit -= share
		}
	}
	available := width
	if available < 0 {
		available = limW
	}
	if available < inf {
		space := max(available-(columns-1)*gap, 0)
		fixed := make([]bool, columns)
		remaining := columns
		for {
			changed := false
			for i, mn := range widths {
				if !fixed[i] && remaining > 0 && mn > space/remaining {
					fixed[i] = true
					space = max(space-mn, 0)
					remaining--
					changed = true
				}
			}
			if !changed {
				break
			}
		}
		for i := range widths {
			if !fixed[i] && remaining > 0 {
				widths[i] = space / remaining
				space -= widths[i]
				remaining--
			}
		}
	}
	result := image.Point{X: (columns - 1) * gap}
	for _, w := range widths {
		result.X += w
	}
	for first := 0; first < len(kids); {
		last := first + 1
		for last < len(kids) && rows[last] == rows[first] {
			last++
		}
		rowHeight := 0
		for j, c := range kids[first:last] {
			i := first + j
			cellWidth := trackWidth(starts[i], spans[i])
			l, t, r, b := e.edges(c.style.margin)
			c.forceW, c.forceH = -1, -1
			if c.style.w.kind == autoLen {
				c.forceW = max(cellWidth-l-r, 0)
			}
			e.layoutChild(c, n, cellWidth, limH)
			rowHeight = max(rowHeight, t+c.size.Y+b)
		}
		if first > 0 {
			result.Y += gap
		}
		x := 0
		for j, c := range kids[first:last] {
			i := first + j
			cellWidth := trackWidth(starts[i], spans[i])
			l, t, _, b := e.edges(c.style.margin)
			if (!n.style.alignSet || n.style.align == Stretch) && c.style.h.kind == autoLen {
				c.forceH = max(rowHeight-t-b, 0)
				e.layoutChild(c, n, cellWidth, limH)
			}
			y := t
			switch n.style.align {
			case Center:
				y = t + (rowHeight-c.size.Y-t-b)/2
			case End:
				y = rowHeight - c.size.Y - b
			}
			n.flowPositions[i] = image.Pt(x+l, result.Y+y)
			x += cellWidth + gap
		}
		result.Y += rowHeight
		first = last
	}
	return result
}
