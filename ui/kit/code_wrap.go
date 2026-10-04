package kit

import "sort"

// Soft wrap splits a long line over several display rows. Rows are built on
// top of folding: rowOf/lineOf map lines to fold rows, and vrows maps those
// to visual rows, each a column range of one line. Without wrap there is one
// visual row per fold row, and every conversion below reduces to it.

// codeVRow is one visual row: columns [start, end) of a line, the last row
// of a line also holding its end.
type codeVRow struct{ line, start, end int }

type codeWrapKey struct {
	rev, rows      uint64
	width, px, tab int
}

// SoftWrap wraps lines at the editor's width instead of scrolling sideways.
// Breaks fall after spaces when a row has one, otherwise mid-word. Up and
// Down move by visual rows; line numbers show on a line's first row.
func (v *CodeEditorView) SoftWrap(on bool) *CodeEditorView {
	if v.wrap != on {
		v.wrap, v.vrows, v.scrollX = on, nil, 0
		v.reveal = true
	}
	return v
}

// buildVRows recomputes the visual rows when the text, folds or width changed.
func (v *CodeEditorView) buildVRows() {
	if !v.wrap {
		v.vrows = nil
		return
	}
	v.buildRows()
	m := v.metrics
	k := codeWrapKey{v.buf.revision, v.rowsGen, v.wrapWidth(), m.px, m.tab}
	if v.vrows != nil && v.vrowsKey == k {
		return
	}
	v.vrowsKey = k
	v.vrows = make([]codeVRow, 0, v.rowCount())
	next := 0 // index into v.rows of the next shown line, when folded
	v.buf.lines.each(0, func(line int, l *codeLine) bool {
		if v.rows != nil {
			if next >= len(v.rows) || v.rows[next] != line {
				return true // folded away
			}
			next++
		}
		for _, seg := range v.wrapLine(line, l.text, k.width) {
			v.vrows = append(v.vrows, codeVRow{line, seg[0], seg[1]})
		}
		return true
	})
}

// wrapWidth is the px a row of text may span.
func (v *CodeEditorView) wrapWidth() int {
	m := v.metrics
	return max(m.space*8, m.size.X-m.textOrigin-m.pad)
}

// wrapLine splits a line into column ranges no wider than width.
func (v *CodeEditorView) wrapLine(line int, l []rune, width int) [][2]int {
	// Most lines fit; bound their width without shaping them. A monospace
	// ASCII rune is a space wide, a tab at most a stop, anything else under
	// two ems.
	m, bound := v.metrics, 0
	for _, r := range l {
		switch {
		case r == '\t':
			bound += m.tab
		case r < 0x80:
			bound += m.space
		default:
			bound += 2 * m.px
		}
	}
	if bound <= width {
		return [][2]int{{0, len(l)}}
	}
	xs := v.colX(line)
	n := len(xs) - 1
	var out [][2]int
	start := 0
	for xs[n]-xs[start] > width {
		end := start + 1
		for end < n && xs[end+1]-xs[start] <= width {
			end++
		}
		brk := end
		for i := end; i > start+1; i-- {
			if l[i-1] == ' ' || l[i-1] == '\t' {
				brk = i
				break
			}
		}
		out = append(out, [2]int{start, brk})
		start = brk
	}
	return append(out, [2]int{start, n})
}

func (v *CodeEditorView) vrowCount() int {
	v.buildVRows()
	if v.vrows == nil {
		return v.rowCount()
	}
	return len(v.vrows)
}

// vrow is the visual row at index r, clamped.
func (v *CodeEditorView) vrow(r int) codeVRow {
	v.buildVRows()
	if v.vrows == nil {
		line := v.lineOf(r)
		return codeVRow{line, 0, len(v.buf.line(line))}
	}
	return v.vrows[min(max(r, 0), len(v.vrows)-1)]
}

// vrowOf is the visual row showing a position. A column at a break belongs
// to the row it starts; a hidden line maps to its fold's row.
func (v *CodeEditorView) vrowOf(p codePos) int {
	v.buildVRows()
	if v.vrows == nil {
		return v.rowOf(p.line)
	}
	line := v.lineOf(v.rowOf(p.line))
	if line != p.line {
		p = codePos{line, 0}
	}
	first := sort.Search(len(v.vrows), func(i int) bool { return v.vrows[i].line >= line })
	r := first
	for r+1 < len(v.vrows) && v.vrows[r+1].line == line && v.vrows[r+1].start <= p.col {
		r++
	}
	return min(r, len(v.vrows)-1)
}

// rowX is a column's x within its visual row.
func (v *CodeEditorView) rowX(p codePos) int {
	return v.colX(p.line)[p.col] - v.colX(p.line)[v.vrow(v.vrowOf(p)).start]
}

// colIn is the column of a visual row nearest x, measured from the row's
// start. Before a break the caret stops short of it, so it stays on the row.
func (v *CodeEditorView) colIn(vr codeVRow, x int) int {
	xs := v.colX(vr.line)
	limit := vr.end
	if limit < len(xs)-1 && limit > vr.start {
		limit--
	}
	x += xs[vr.start]
	col := vr.start
	for col < limit && x > (xs[col]+xs[col+1])/2 {
		col++
	}
	return col
}
