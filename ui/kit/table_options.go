package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
)

func (c *ColumnSpec) Selectable(on bool) *ColumnSpec { c.noSelect = !on; return c }
func (c *ColumnSpec) Resizable(on bool) *ColumnSpec  { c.noResize = !on; return c }

// Movable controls MoveColumn; a locked column also prevents other columns crossing it.
// Explicit layout restoration remains application-controlled.
func (c *ColumnSpec) Movable(on bool) *ColumnSpec { c.noMove = !on; return c }
func (v *TableView) Stripe(on bool) *TableView    { v.stripe = on; return v }

// RowHeight sets uniform row height including the separator. Zero restores 40dp.
func (v *TableView) RowHeight(dp float32) *TableView {
	if dp == 0 {
		dp = 40
	}
	if dp >= 24 && dp <= 256 {
		v.list.rowH = dp
		v.reveal = v.selected >= 0
	}
	return v
}

// ColumnSelect enables independent whole-column selection, retained across SetRows.
func (v *TableView) ColumnSelect() *TableView {
	v.cellMode = false
	v.columnMode = true
	v.cells = nil
	v.selection = nil
	v.selected = -1
	v.selectedColumns = make(map[int]bool)
	v.columnAnchor = -1
	v.activeColumn = -1
	return v
}
func (v *TableView) OnColumnSelectionChange(fn func([]int)) *TableView { v.onColumns = fn; return v }
func (v *TableView) SelectedColumns() []int {
	var result []int
	for _, c := range v.columns {
		if v.selectedColumns[c] {
			result = append(result, c)
		}
	}
	return result
}
func (v *TableView) SetSelectedColumns(columns []int) {
	if !v.columnMode {
		return
	}
	v.selectedColumns = make(map[int]bool)
	v.activeColumn = -1
	for _, c := range columns {
		if c >= 0 && c < len(v.cols) && !v.cols[c].noSelect {
			v.selectedColumns[c] = true
			v.activeColumn = c
		}
	}
	v.columnAnchor = v.activeColumn
}
func (v *TableView) selectableColumns() []int {
	return slices.DeleteFunc(v.visibleColumns(), func(c int) bool { return v.cols[c].noSelect })
}
func (v *TableView) notifyColumns(before []int) {
	after := v.SelectedColumns()
	if !slices.Equal(before, after) && v.onColumns != nil {
		v.onColumns(after)
	}
}
func (v *TableView) chooseWholeColumn(cx *el.Context, c int, mods key.Modifiers) {
	columns := v.selectableColumns()
	pos := slices.Index(columns, c)
	if v.disabled || pos < 0 {
		return
	}
	before := v.SelectedColumns()
	add := mods.Contain(key.ModShortcut)
	anchor := slices.Index(columns, v.columnAnchor)
	wasSelected := v.selectedColumns[c]
	if !add {
		v.selectedColumns = make(map[int]bool)
	}
	if mods.Contain(key.ModShift) && anchor >= 0 {
		for _, column := range columns[min(anchor, pos) : max(anchor, pos)+1] {
			v.selectedColumns[column] = true
		}
	} else {
		if add && wasSelected {
			delete(v.selectedColumns, c)
		} else {
			v.selectedColumns[c] = true
		}
		v.columnAnchor = c
	}
	v.activeColumn = c
	v.activeCell.Column = c
	v.revealCell(cx)
	cx.Focus(autoID("table", v))
	v.notifyColumns(before)
}
func (v *TableView) columnKey(cx *el.Context, e el.KeyEvent) bool {
	columns := v.selectableColumns()
	if len(columns) == 0 {
		return false
	}
	pos := slices.Index(columns, v.activeColumn)
	switch key.Name(e.Name) {
	case key.NameLeftArrow:
		if pos < 0 {
			pos = 0
		} else {
			pos--
		}
	case key.NameRightArrow:
		pos++
	case key.NameHome:
		pos = 0
	case key.NameEnd:
		pos = len(columns) - 1
	default:
		return false
	}
	if e.State == el.KeyPress {
		v.chooseWholeColumn(cx, columns[min(max(pos, 0), len(columns)-1)], e.Modifiers&key.ModShift)
	}
	return true
}
