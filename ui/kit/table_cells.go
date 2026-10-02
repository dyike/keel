package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
)

// TableCell is a source row/column address, unchanged by display sorting.
type TableCell struct{ Row, Column int }

// CellSelect enables cell rectangles and whole-column selection. Click a
// header to select its column; double-click it to sort. Row selection is cleared.
func (v *TableView) CellSelect() *TableView {
	v.cellMode = true
	v.cells = make(map[TableCell]bool)
	v.activeCell, v.cellAnchor = TableCell{-1, -1}, TableCell{-1, -1}
	v.selected = -1
	return v
}
func (v *TableView) OnCellSelectionChange(fn func([]TableCell)) *TableView { v.onCells = fn; return v }

// SelectedCells returns a snapshot in display order, including hidden columns.
func (v *TableView) SelectedCells() []TableCell {
	var cells []TableCell
	for _, r := range v.order {
		for _, c := range v.columns {
			cell := TableCell{r, c}
			if v.cells[cell] {
				cells = append(cells, cell)
			}
		}
	}
	return cells
}
func (v *TableView) SetSelectedCells(cells []TableCell) {
	if !v.cellMode {
		return
	}
	v.cells = make(map[TableCell]bool)
	v.activeCell = TableCell{-1, -1}
	for _, cell := range cells {
		if cell.Row >= 0 && cell.Row < len(v.rows) && cell.Column >= 0 && cell.Column < len(v.cols) {
			v.cells[cell] = true
			v.activeCell = cell
		}
	}
	v.cellAnchor = v.activeCell
	v.selected, v.reveal = v.activeCell.Row, true
}
func (v *TableView) notifyCells(before []TableCell) {
	after := v.SelectedCells()
	if !slices.Equal(before, after) && v.onCells != nil {
		v.onCells(after)
	}
}
func (v *TableView) chooseCell(cx *el.Context, row, column int, mods key.Modifiers) {
	if row < 0 || row >= len(v.rows) || column < 0 || column >= len(v.cols) {
		return
	}
	before := v.SelectedCells()
	columns := v.visibleColumns()
	add := mods.Contain(key.ModShortcut)
	a, b := v.position(v.cellAnchor.Row), slices.Index(columns, v.cellAnchor.Column)
	target := TableCell{row, column}
	if mods.Contain(key.ModShift) && a >= 0 && b >= 0 {
		if !add {
			v.cells = make(map[TableCell]bool)
		}
		endRow, endCol := v.position(row), slices.Index(columns, column)
		for r := min(a, endRow); r <= max(a, endRow); r++ {
			for c := min(b, endCol); c <= max(b, endCol); c++ {
				v.cells[TableCell{v.order[r], columns[c]}] = true
			}
		}
	} else {
		if !add {
			v.cells = make(map[TableCell]bool)
		}
		if add && v.cells[target] {
			delete(v.cells, target)
		} else {
			v.cells[target] = true
		}
		v.cellAnchor = target
	}
	v.activeCell = target
	v.choose(cx, row)
	v.revealCell(cx)
	v.notifyCells(before)
}
func (v *TableView) chooseColumn(cx *el.Context, column int, mods key.Modifiers) {
	if len(v.rows) == 0 {
		return
	}
	before := v.SelectedCells()
	columns := v.visibleColumns()
	a, b := slices.Index(columns, column), slices.Index(columns, v.cellAnchor.Column)
	if a < 0 {
		return
	}
	if !mods.Contain(key.ModShift) || b < 0 {
		b = a
	}
	add := mods.Contain(key.ModShortcut)
	all := true
	for _, r := range v.order {
		if !v.cells[TableCell{r, column}] {
			all = false
			break
		}
	}
	if !add {
		v.cells = make(map[TableCell]bool)
	}
	for c := min(a, b); c <= max(a, b); c++ {
		for _, r := range v.order {
			cell := TableCell{r, columns[c]}
			if add && all && !mods.Contain(key.ModShift) {
				delete(v.cells, cell)
			} else {
				v.cells[cell] = true
			}
		}
	}
	row := v.selected
	if row < 0 {
		row = v.order[0]
	}
	v.activeCell = TableCell{row, column}
	if !mods.Contain(key.ModShift) || v.cellAnchor.Row < 0 {
		v.cellAnchor = v.activeCell
	}
	v.choose(cx, row)
	v.revealCell(cx)
	v.notifyCells(before)
}
func (v *TableView) revealCell(cx *el.Context) {
	columns := v.visibleColumns()
	pos := slices.Index(columns, v.activeCell.Column)
	if pos < 0 {
		return
	}
	left, right := v.frozenCounts(columns)
	if pos < left || pos >= len(columns)-right {
		return
	}
	var x, leftWidth, rightWidth, width float32
	for i, c := range columns {
		w := v.cols[c].width
		if w <= 0 {
			w = max(v.widths[c], minColumn)
		}
		if i < pos {
			x += w
		}
		if i == pos {
			width = w
		}
		if i < left {
			leftWidth += w
		}
		if i >= len(columns)-right {
			rightWidth += w
		}
	}
	cx.ScrollIntoViewX(autoID("table", v), x-leftWidth, x+width+rightWidth)
}
func (v *TableView) cellKey(cx *el.Context, e el.KeyEvent) bool {
	columns := v.visibleColumns()
	if len(columns) == 0 || len(v.order) == 0 {
		return false
	}
	r, c := v.position(v.activeCell.Row), slices.Index(columns, v.activeCell.Column)
	if r < 0 {
		r = 0
	}
	if c < 0 {
		c = 0
	}
	switch key.Name(e.Name) {
	case key.NameLeftArrow:
		c--
	case key.NameRightArrow:
		c++
	case key.NameUpArrow:
		r--
	case key.NameDownArrow:
		r++
	case key.NamePageUp:
		r -= 8
	case key.NamePageDown:
		r += 8
	case key.NameHome:
		c = 0
		if e.Modifiers.Contain(key.ModShortcut) {
			r = 0
		}
	case key.NameEnd:
		c = len(columns) - 1
		if e.Modifiers.Contain(key.ModShortcut) {
			r = len(v.order) - 1
		}
	default:
		return false
	}
	if e.State == el.KeyPress {
		v.chooseCell(cx, v.order[max(0, min(r, len(v.order)-1))], columns[max(0, min(c, len(columns)-1))], e.Modifiers&key.ModShift)
		cx.Focus(autoID("table", v))
	}
	return true
}
