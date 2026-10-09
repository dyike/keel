package kit

import (
	"encoding/csv"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
	"strings"
)

// MultiSelect enables additive and range row selection. Value is the active
// row; SelectedRows reports membership in current display order.
func (v *TableView) MultiSelect() *TableView {
	rows := v.SelectedRows()
	v.columnMode = false
	v.selectedColumns = nil
	v.cellMode = false
	v.cells = nil
	v.multi = true
	v.SetSelectedRows(rows)
	return v
}
func (v *TableView) OnSelectionChange(fn func([]int)) *TableView { v.onSelection = fn; return v }
func (v *TableView) SelectedRows() []int {
	if v.columnMode {
		if len(v.SelectedColumns()) > 0 {
			return slices.Clone(v.order)
		}
		return nil
	}
	if v.cellMode {
		set := make(map[int]bool)
		for cell := range v.cells {
			set[cell.Row] = true
		}
		var rows []int
		for _, r := range v.order {
			if set[r] {
				rows = append(rows, r)
			}
		}
		return rows
	}
	if !v.multi {
		if v.selected >= 0 && v.position(v.selected) >= 0 {
			return []int{v.selected}
		}
		return nil
	}
	var rows []int
	for _, row := range v.order {
		if v.selection[row] {
			rows = append(rows, row)
		}
	}
	return rows
}

// SetSelectedRows replaces row selection without callbacks. Invalid indexes
// are ignored. In single selection mode only the first valid row is selected.
func (v *TableView) SetSelectedRows(rows []int) {
	if v.columnMode {
		return
	}
	if v.cellMode {
		var cells []TableCell
		for _, r := range rows {
			for _, c := range v.visibleColumns() {
				cells = append(cells, TableCell{r, c})
			}
		}
		v.SetSelectedCells(cells)
		return
	}
	v.selection = make(map[int]bool)
	active := -1
	for _, row := range rows {
		if row >= 0 && row < len(v.rows) {
			v.selection[row] = true
			active = row
			if !v.multi {
				break
			}
		}
	}
	v.selected, v.anchor, v.reveal = active, active, true
}
func (v *TableView) rowSelected(row int) bool {
	if v.cellMode || v.columnMode {
		return false
	}
	if v.multi {
		return v.selection[row]
	}
	return row == v.selected
}

func (v *TableView) chooseRows(cx *el.Context, row int, mods key.Modifiers) {
	before := v.SelectedRows()
	if v.multi {
		additive := mods.Contain(key.ModShortcut)
		if mods.Contain(key.ModShift) && v.anchor >= 0 && v.position(v.anchor) >= 0 {
			if !additive {
				v.selection = make(map[int]bool)
			}
			a, b := v.position(v.anchor), v.position(row)
			for pos := min(a, b); pos <= max(a, b); pos++ {
				v.selection[v.order[pos]] = true
			}
		} else {
			if !additive {
				v.selection = make(map[int]bool)
			}
			if additive && v.selection[row] {
				delete(v.selection, row)
			} else {
				v.selection[row] = true
			}
			v.anchor = row
		}
	}
	v.choose(cx, row)
	if !slices.Equal(before, v.SelectedRows()) && v.onSelection != nil {
		v.onSelection(v.SelectedRows())
	}
}

// SelectionText copies visible columns of selected rows as TSV, in current
// display order. Tabs, newlines and quotes inside values are CSV-escaped.
func (v *TableView) SelectionText() string {
	var out strings.Builder
	writer := csv.NewWriter(&out)
	writer.Comma = '\t'
	columns := v.visibleColumns()
	if v.columnMode {
		columns = slices.DeleteFunc(columns, func(c int) bool { return !v.selectedColumns[c] })
	}
	if v.cellMode {
		columns = slices.DeleteFunc(columns, func(c int) bool {
			for cell := range v.cells {
				if cell.Column == c {
					return false
				}
			}
			return true
		})
	}
	if len(columns) == 0 {
		return ""
	}
	for _, row := range v.SelectedRows() {
		values := make([]string, len(columns))
		for i, c := range columns {
			if c < len(v.rows[row]) && (!v.cellMode || v.cells[TableCell{row, c}]) {
				values[i] = v.rows[row][c]
			}
		}
		_ = writer.Write(values)
	}
	writer.Flush()
	return strings.TrimSuffix(out.String(), "\n")
}
func (v *TableView) selectionKey(e el.KeyEvent) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	switch key.Name(e.Name) {
	case "C":
		if e.State == el.KeyPress {
			if len(v.SelectedRows()) > 0 && len(v.visibleColumns()) > 0 {
				el.WriteClipboard(v.SelectionText())
			}
		}
		return true
	case "A":
		if v.columnMode {
			if e.State == el.KeyPress {
				before := v.SelectedColumns()
				v.SetSelectedColumns(v.selectableColumns())
				v.notifyColumns(before)
			}
			return true
		}
		if v.cellMode {
			if e.State == el.KeyPress {
				before := v.SelectedCells()
				v.cells = make(map[TableCell]bool)
				for _, row := range v.order {
					for _, c := range v.selectableColumns() {
						v.cells[TableCell{row, c}] = true
					}
				}
				v.notifyCells(before)
			}
			return true
		}
		if !v.multi {
			return false
		}
		if e.State == el.KeyPress {
			before := v.SelectedRows()
			v.selection = make(map[int]bool, len(v.rows))
			for _, row := range v.order {
				v.selection[row] = true
			}
			if !slices.Equal(before, v.SelectedRows()) && v.onSelection != nil {
				v.onSelection(v.SelectedRows())
			}
		}
		return true
	}
	return false
}
