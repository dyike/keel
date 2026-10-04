package kit

import (
	"fmt"
	"math"
	"slices"
)

// TableColumnState identifies a source column, independently of its display order.
// Width zero uses Flex; Hidden columns retain their position and width.
type TableColumnState struct {
	Column int     `json:"column"`
	Width  float32 `json:"width"`
	Flex   float32 `json:"flex"`
	Hidden bool    `json:"hidden"`
}

// TableLayout is a caller-owned, JSON-serializable column layout. Restore it
// only to a table with the same source-column schema.
type TableLayout struct {
	Columns     []TableColumnState `json:"columns"`
	FrozenLeft  int                `json:"frozen_left"`
	FrozenRight int                `json:"frozen_right"`
}

func (v *TableView) visibleColumns() []int {
	columns := make([]int, 0, len(v.columns))
	for _, c := range v.columns {
		if !v.hidden[c] {
			columns = append(columns, c)
		}
	}
	return columns
}

// MoveColumn moves a source column to a display position (including hidden
// columns). Invalid indexes are ignored. It never moves row data.
func (v *TableView) MoveColumn(column, position int) {
	if column < 0 || column >= len(v.cols) || position < 0 || position >= len(v.cols) {
		return
	}
	old := slices.Index(v.columns, column)
	if !v.canMoveColumn(column, position) {
		return
	}
	v.columnDrag = nil
	v.columns = slices.Delete(v.columns, old, old+1)
	v.columns = slices.Insert(v.columns, position, column)
	v.ensureFrozenWidths()
}

// SetColumnVisible hides or shows a source column without changing its data,
// sort order, or configured position. All columns may be hidden.
func (v *TableView) SetColumnVisible(column int, visible bool) {
	if column < 0 || column >= len(v.cols) {
		return
	}
	v.columnDrag = nil
	v.hidden[column] = !visible
	v.ensureFrozenWidths()
}

// LayoutState returns a deep snapshot suitable for encoding with encoding/json.
func (v *TableView) LayoutState() TableLayout {
	state := TableLayout{FrozenLeft: v.frozenLeft, FrozenRight: v.frozenRight}
	for _, c := range v.columns {
		col := v.cols[c]
		state.Columns = append(state.Columns, TableColumnState{c, col.width, col.flex, v.hidden[c]})
	}
	return state
}

// SetLayoutState validates the entire snapshot before applying any changes.
// It does not change sorting, row selection, or invoke user callbacks.
func (v *TableView) SetLayoutState(state TableLayout) error {
	if len(state.Columns) != len(v.cols) || state.FrozenLeft < 0 || state.FrozenRight < 0 || state.FrozenLeft > len(v.cols) || state.FrozenRight > len(v.cols)-state.FrozenLeft {
		return fmt.Errorf("kit.Table: incompatible column layout")
	}
	seen := make([]bool, len(v.cols))
	for _, col := range state.Columns {
		if col.Column < 0 || col.Column >= len(v.cols) || seen[col.Column] || math.IsNaN(float64(col.Width)) || math.IsInf(float64(col.Width), 0) || col.Width < 0 || col.Width > 0 && col.Width < minColumn || !(col.Flex > 0) || math.IsInf(float64(col.Flex), 0) {
			return fmt.Errorf("kit.Table: invalid column state")
		}
		seen[col.Column] = true
	}
	v.columnDrag = nil
	for pos, col := range state.Columns {
		v.columns[pos] = col.Column
		v.cols[col.Column].width = col.Width
		v.cols[col.Column].flex = col.Flex
		v.hidden[col.Column] = col.Hidden
	}
	v.frozenLeft, v.frozenRight = state.FrozenLeft, state.FrozenRight
	v.ensureFrozenWidths()
	return nil
}

func (v *TableView) frozenCounts(columns []int) (int, int) {
	left := min(v.frozenLeft, len(columns))
	return left, min(v.frozenRight, len(columns)-left)
}
func (v *TableView) ensureFrozenWidths() {
	columns := v.visibleColumns()
	left, right := v.frozenCounts(columns)
	for pos, c := range columns {
		if (pos < left || pos >= len(columns)-right) && v.cols[c].width <= 0 {
			v.cols[c].Width(120)
		}
	}
}
