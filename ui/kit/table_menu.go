package kit

import (
	"github.com/dyike/keel/ui/el"
	"strconv"
)

// RowMenu creates a menu for a source row when requested, not on every frame.
func (v *TableView) RowMenu(build func(row int) *MenuView) *TableView { v.rowMenu = build; return v }

// CellMenu takes precedence over RowMenu for cell context actions. Returning
// nil falls back to RowMenu. Indexes refer to the source data, not visual order.
func (v *TableView) CellMenu(build func(row, column int) *MenuView) *TableView {
	v.cellMenu = build
	return v
}
func (v *TableView) cellID(row, column int) string {
	return autoID("table", v) + "/cell/" + strconv.Itoa(row) + "/" + strconv.Itoa(column)
}
func (v *TableView) openMenu(cx *el.Context, row, column int) {
	if v.disabled || row < 0 || row >= len(v.rows) {
		return
	}
	var menu *MenuView
	if v.cellMenu != nil && column >= 0 {
		menu = v.cellMenu(row, column)
	}
	if menu == nil && v.rowMenu != nil {
		menu = v.rowMenu(row)
	}
	if menu == nil {
		return
	}
	// Preserve an existing multi-selection when opening its context menu.
	if v.cellMode {
		if !v.cells[TableCell{row, column}] {
			v.chooseCell(cx, row, column, 0)
		}
	} else if !v.rowSelected(row) {
		v.chooseRows(cx, row, 0)
	}
	v.contextMenu, v.contextCell = menu, TableCell{row, column}
	menu.SetValue(true)
}
func (v *TableView) renderContextMenu(cx *el.Context) {
	menu := v.contextMenu
	if menu == nil || !menu.Value() {
		return
	}
	cell := v.contextCell
	if v.disabled || v.position(cell.Row) < 0 || cell.Column < 0 || cell.Column >= len(v.cols) || v.hidden[cell.Column] {
		menu.SetValue(false)
		return
	}
	anchor := v.cellID(cell.Row, cell.Column)

	cx.Overlay(autoID("menu", menu), el.Anchored(anchor, menu.panel(cx)).Placement(el.Bottom, el.Start).Modal().TrapFocus().OnDismiss(func() { menu.SetValue(false) }))
	menu.renderSub(cx)
}
