package el

import (
	"image"
	"slices"
)

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
	var scratch [16]*Node
	kids := inFlow(n, scratch[:0])
	n.flowPositions = slices.Grow(n.flowPositions[:0], len(kids))[:len(kids)]
	limit := limW
	if width >= 0 {
		limit = width
	}
	colGapDp := n.style.colGapDp()
	colGap := e.dp(colGapDp)
	rowGap := e.dp(n.style.rowGapDp())
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
			if last > first && used+colGap+basis > limit {
				break
			}
			if last > first {
				used += colGap
			}
			used += basis
			last++
		}
		line := Node{style: Style{row: true, gap: colGapDp, align: n.style.align, alignSet: n.style.alignSet, justify: n.style.justify}, textStyle: n.textStyle}
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
			result.Y += rowGap
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
	var scratch [16]*Node
	kids := inFlow(n, scratch[:0])
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
	align := n.style.crossAlign()
	anchor := 0
	if align == ContentBottom {
		anchor, _ = e.contentBottomExtents(kids)
	}
	if align == FirstLine {
		anchor, _ = e.firstLineExtents(kids)
	}
	out := make([]image.Point, len(kids))
	for i, c := range kids {
		l, t, r, b := e.edges(c.style.margin)
		y := t
		switch c.style.itemAlign(align) {
		case Center:
			y = t + (size.Y-c.size.Y-t-b)/2
		case End:
			y = size.Y - c.size.Y - b
		case ContentBottom:
			y = anchor - e.contentBottom(c)
		case FirstLine:
			y = anchor - e.firstLineOf(c)
		}
		out[i] = image.Pt(cursor+l, y)
		cursor += l + c.size.X + r + gap + extra
	}
	return out
}

func (e *engine) grid(n *Node, width, limW, limH int) image.Point {
	var scratch [16]*Node
	kids := inFlow(n, scratch[:0])
	n.flowPositions = slices.Grow(n.flowPositions[:0], len(kids))[:len(kids)]
	if len(kids) == 0 {
		return image.Point{}
	}
	columns := n.style.grid
	colGap := e.dp(n.style.colGapDp())
	rowGap := e.dp(n.style.rowGapDp())

	// 2D auto-placement supporting both ColSpan and RowSpan.
	starts := make([]int, len(kids))
	spans := make([]int, len(kids))
	rowStarts := make([]int, len(kids))
	rowSpans := make([]int, len(kids))
	var occupied [][]bool
	ensureRow := func(r int) {
		for len(occupied) <= r {
			occupied = append(occupied, make([]bool, columns))
		}
	}
	fits := func(r, col, cSpan, rSpan int) bool {
		if col+cSpan > columns {
			return false
		}
		for rr := r; rr < r+rSpan; rr++ {
			if rr >= len(occupied) {
				continue
			}
			for cc := col; cc < col+cSpan; cc++ {
				if occupied[rr][cc] {
					return false
				}
			}
		}
		return true
	}
	cursorRow, cursorCol, totalRows := 0, 0, 0
	for i, c := range kids {
		cSpan := min(max(c.style.colSpan, 1), columns)
		rSpan := max(c.style.rowSpan, 1)
		r, col := cursorRow, cursorCol
		for !fits(r, col, cSpan, rSpan) {
			col++
			if col+cSpan > columns {
				col = 0
				r++
			}
		}
		ensureRow(r + rSpan - 1)
		for rr := r; rr < r+rSpan; rr++ {
			for cc := col; cc < col+cSpan; cc++ {
				occupied[rr][cc] = true
			}
		}
		starts[i], spans[i] = col, cSpan
		rowStarts[i], rowSpans[i] = r, rSpan
		totalRows = max(totalRows, r+rSpan)
		cursorRow, cursorCol = r, col+cSpan
		if cursorCol >= columns {
			cursorRow++
			cursorCol = 0
		}
	}

	available := width
	if available < 0 {
		available = limW
	}

	// Column track sizing: equal flexible tracks by default, or explicit GridCols.
	widths := make([]int, columns)
	flexWeight := make([]float32, columns)
	isAutoTrack := make([]bool, columns)
	isFixedTrack := make([]bool, columns)
	if len(n.style.gridCols) == 0 {
		for i := range flexWeight {
			flexWeight[i] = 1
		}
	} else {
		for i := 0; i < columns; i++ {
			t := n.style.gridCols[i]
			switch t.kind {
			case dpLen, spLen, fracLen:
				if px := t.px(e.m, available); px >= 0 {
					widths[i] = px
					isFixedTrack[i] = true
				} else {
					isAutoTrack[i] = true
				}
			case frLen:
				flexWeight[i] = t.value
			default: // autoLen
				isAutoTrack[i] = true
			}
		}
	}

	trackWidth := func(start, span int) int {
		total := (span - 1) * colGap
		for _, w := range widths[start : start+span] {
			total += w
		}
		return total
	}
	spansAuto := func(start, span int) bool {
		for j := start; j < start+span; j++ {
			if isAutoTrack[j] {
				return true
			}
		}
		return false
	}

	for i, c := range kids {
		l, _, r, _ := e.edges(c.style.margin)
		minimum := max(c.style.minW.px(e.m, limW), 0)
		if c.style.w.kind == dpLen || c.style.w.kind == spLen {
			minimum = max(minimum, c.style.w.px(e.m, limW))
		}
		if limW >= inf || spansAuto(starts[i], spans[i]) {
			c.forceW, c.forceH = -1, -1
			e.layoutChild(c, n, limW, limH)
			minimum = max(minimum, c.size.X)
		}
		// Distribute a cell's minimum deficit across its covered tracks,
		// preferring non-fixed tracks when any are present.
		start, span := starts[i], spans[i]
		deficit := max(minimum+l+r-trackWidth(start, span), 0)
		if deficit <= 0 {
			continue
		}
		nonFixed := 0
		for j := start; j < start+span; j++ {
			if !isFixedTrack[j] {
				nonFixed++
			}
		}
		if nonFixed > 0 {
			rem := nonFixed
			for j := start; j < start+span; j++ {
				if isFixedTrack[j] {
					continue
				}
				share := deficit / rem
				widths[j] += share
				deficit -= share
				rem--
			}
		} else {
			for j := 0; j < span; j++ {
				share := deficit / (span - j)
				widths[start+j] += share
				deficit -= share
			}
		}
	}

	if available < inf {
		space := available - (columns-1)*colGap
		fixed := make([]bool, columns)
		var totalWeight float32
		for i := range widths {
			if flexWeight[i] <= 0 {
				fixed[i] = true
				space -= widths[i]
			} else {
				totalWeight += flexWeight[i]
			}
		}
		space = max(space, 0)
		for {
			changed := false
			for i, mn := range widths {
				if !fixed[i] && totalWeight > 0 && float32(mn) > float32(space)*flexWeight[i]/totalWeight {
					fixed[i] = true
					space = max(space-mn, 0)
					totalWeight -= flexWeight[i]
					changed = true
				}
			}
			if !changed {
				break
			}
		}
		for i := range widths {
			if !fixed[i] && totalWeight > 0 {
				w := int(float32(space) * flexWeight[i] / totalWeight)
				widths[i] = w
				space -= w
				totalWeight -= flexWeight[i]
			}
		}
	}

	colX := make([]int, columns)
	result := image.Point{}
	for i, w := range widths {
		colX[i] = result.X
		result.X += w
		if i+1 < columns {
			result.X += colGap
		}
	}

	// Pass 1: measure every child at its column track width; size single-row tracks.
	rowHeights := make([]int, totalRows)
	for i, c := range kids {
		cellWidth := trackWidth(starts[i], spans[i])
		l, t, r, b := e.edges(c.style.margin)
		c.forceW, c.forceH = -1, -1
		if c.style.w.kind == autoLen {
			c.forceW = max(cellWidth-l-r, 0)
		}
		e.layoutChild(c, n, cellWidth, limH)
		if rowSpans[i] == 1 {
			rowHeights[rowStarts[i]] = max(rowHeights[rowStarts[i]], t+c.size.Y+b)
		}
	}

	trackHeight := func(start, span int) int {
		total := (span - 1) * rowGap
		for _, h := range rowHeights[start : start+span] {
			total += h
		}
		return total
	}

	// Pass 2: expand spanned rows to fit multi-row items.
	for i, c := range kids {
		if rowSpans[i] <= 1 {
			continue
		}
		_, t, _, b := e.edges(c.style.margin)
		deficit := max(t+c.size.Y+b-trackHeight(rowStarts[i], rowSpans[i]), 0)
		if deficit <= 0 {
			continue
		}
		rStart, rSpan := rowStarts[i], rowSpans[i]
		for j := 0; j < rSpan; j++ {
			share := deficit / (rSpan - j)
			rowHeights[rStart+j] += share
			deficit -= share
		}
	}

	rowY := make([]int, totalRows)
	for r, h := range rowHeights {
		rowY[r] = result.Y
		result.Y += h
		if r+1 < totalRows {
			result.Y += rowGap
		}
	}

	// Pass 3: vertical stretch (honoring Self) and final cell placement.
	parentAlign := n.style.crossAlign()
	for i, c := range kids {
		cellWidth := trackWidth(starts[i], spans[i])
		cellHeight := trackHeight(rowStarts[i], rowSpans[i])
		l, t, _, b := e.edges(c.style.margin)
		childAlign := c.style.itemAlign(parentAlign)
		if childAlign == Stretch && c.style.h.kind == autoLen {
			wantH := max(cellHeight-t-b, 0)
			if c.size.Y != wantH {
				c.forceH = wantH
				e.layoutChild(c, n, cellWidth, limH)
			}
		}
		y := t
		switch childAlign {
		case Center:
			y = t + (cellHeight-c.size.Y-t-b)/2
		case End:
			y = cellHeight - c.size.Y - b
		}
		n.flowPositions[i] = image.Pt(colX[starts[i]]+l, rowY[rowStarts[i]]+y)
	}
	return result
}
