package kit

import "sort"

// Folding works on indentation, the structure every language shares: a
// region is a line and the more-indented lines after it (blank lines
// included). A closing bracket back at the header's indentation stays
// visible, so a folded Go function shows `func f() {` … `}`.
//
// Folded lines leave the display: rows count displayed lines, and every
// conversion between screen and text goes through rowOf and lineOf.

// spliced keeps folds and caches in step with edits: lines from on were
// replaced, removed by added (removed < 0: the whole text).
func (v *CodeEditorView) spliced(from, removed, added int) {
	clear(v.foldEnds)
	if removed < 0 {
		clear(v.folds)
		return
	}
	if len(v.folds) == 0 {
		return
	}
	next := map[int]bool{}
	for start := range v.folds {
		switch {
		case start < from:
			next[start] = true
		case start >= from+removed:
			next[start+added-removed] = true
		}
		// Folds whose header was edited open: the region may have changed.
	}
	v.folds = next
}

// indentWidth is a line's indentation in columns, tabs to the tab stop.
func (v *CodeEditorView) indentWidth(line int) int {
	w := 0
	for _, r := range v.buf.line(line) {
		switch r {
		case ' ':
			w++
		case '\t':
			w = (w/v.tabSize + 1) * v.tabSize
		default:
			return w
		}
	}
	return w
}

// foldEnd is the last line of the region starting at line, or line if it
// starts none.
func (v *CodeEditorView) foldEnd(line int) int {
	if v.foldRev != v.buf.revision {
		clear(v.foldEnds)
		v.foldRev = v.buf.revision
	}
	if e, ok := v.foldEnds[line]; ok {
		return e
	}
	end := line
	if line >= 0 && line < v.buf.count()-1 && !v.buf.blank(line) {
		base := v.indentWidth(line)
		last := line
		v.buf.lines.each(line+1, func(i int, l *codeLine) bool {
			if v.buf.blank(i) {
				return true
			}
			if v.indentWidth(i) <= base {
				return false
			}
			last = i
			return true
		})
		end = last
	}
	v.foldEnds[line] = end
	return end
}

// buildRows lists the displayed lines when something is folded.
func (v *CodeEditorView) buildRows() {
	key := len(v.folds)
	if v.rowsRev == v.buf.revision && v.rowsKey == key && (v.rows != nil) == (len(v.folds) > 0) {
		return
	}
	v.rowsRev, v.rowsKey = v.buf.revision, key
	if len(v.folds) == 0 {
		v.rows = nil
		return
	}
	starts := make([]int, 0, len(v.folds))
	for s := range v.folds {
		starts = append(starts, s)
	}
	sort.Ints(starts)
	rows := make([]int, 0, v.buf.count())
	next := 0
	for _, s := range starts {
		if s < next || s >= v.buf.count() {
			continue
		}
		end := v.foldEnd(s)
		for i := next; i <= s; i++ {
			rows = append(rows, i)
		}
		next = max(end+1, s+1)
	}
	for i := next; i < v.buf.count(); i++ {
		rows = append(rows, i)
	}
	v.rows = rows
}

// invalidateRows rebuilds the row map on the next use, after folds changed.
func (v *CodeEditorView) invalidateRows() { v.rowsKey = -1 }

func (v *CodeEditorView) rowCount() int {
	v.buildRows()
	if v.rows == nil {
		return v.buf.count()
	}
	return len(v.rows)
}

// lineOf is the line shown at a display row.
func (v *CodeEditorView) lineOf(row int) int {
	v.buildRows()
	if v.rows == nil {
		return min(max(row, 0), v.buf.count()-1)
	}
	return v.rows[min(max(row, 0), len(v.rows)-1)]
}

// rowOf is the display row of a line; a hidden line maps to its fold's row.
func (v *CodeEditorView) rowOf(line int) int {
	v.buildRows()
	if v.rows == nil {
		return line
	}
	r := sort.SearchInts(v.rows, line)
	if r < len(v.rows) && v.rows[r] == line {
		return r
	}
	return max(0, r-1)
}

// hidden reports whether a line is inside a folded region.
func (v *CodeEditorView) hidden(line int) bool {
	v.buildRows()
	if v.rows == nil {
		return false
	}
	r := sort.SearchInts(v.rows, line)
	return r >= len(v.rows) || v.rows[r] != line
}

// keepCaretsVisible opens folds that hide a caret.
func (v *CodeEditorView) keepCaretsVisible() {
	v.invalidateRows()
	for _, s := range v.sels {
		for start := range v.folds {
			if c := s.caret.line; c > start && c <= v.foldEnd(start) {
				delete(v.folds, start)
			}
		}
	}
	v.invalidateRows()
}

// toggleFold folds or opens the region at a line, or the one containing it.
func (v *CodeEditorView) toggleFold(line int) {
	if v.folds[line] {
		delete(v.folds, line)
	} else if v.foldEnd(line) > line {
		v.folds[line] = true
	} else {
		// Fold the innermost region holding the line.
		for l := line - 1; l >= 0 && line-l < 5000; l-- {
			if v.foldEnd(l) >= line {
				v.folds[l] = true
				break
			}
		}
	}
	v.invalidateRows()
	v.keepCaretsVisible()
}
