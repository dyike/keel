package kit

import (
	"cmp"
	"math"
	"slices"
	"strconv"
	"strings"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ColumnSpec describes a Table column. Create one with Col.
type ColumnSpec struct {
	title   string
	flex    float32 // share of the free width, when width is 0
	width   float32 // fixed width in dp
	numeric bool
	noSort  bool
	cell    func(cx *el.Context, row int) el.Element
}

// Col creates a column that takes an equal share of the width.
func Col(title string) *ColumnSpec { return &ColumnSpec{title: title, flex: 1} }

// Flex sets the column's share of the free width relative to other flexible columns.
func (c *ColumnSpec) Flex(w float32) *ColumnSpec { c.flex, c.width = max(w, 0.1), 0; return c }

// Width fixes the column width in dp. Users can still resize it.
func (c *ColumnSpec) Width(dp float32) *ColumnSpec { c.width = max(dp, minColumn); return c }

// Numeric right-aligns the column and sorts it as numbers.
func (c *ColumnSpec) Numeric() *ColumnSpec { c.numeric = true; return c }

// NoSort stops the header from sorting by this column.
func (c *ColumnSpec) NoSort() *ColumnSpec { c.noSort = true; return c }

// Cell renders the column's cells with fn instead of plain text; row indexes
// the data given to SetRows. Agents still see the row's text.
func (c *ColumnSpec) Cell(fn func(cx *el.Context, row int) el.Element) *ColumnSpec {
	c.cell = fn
	return c
}

const minColumn = 40

// TableView shows rows of text in columns. Click a header to sort by it (again
// to reverse); drag a header's right edge to resize the column. Click a row,
// or use ↑ ↓ Home End PageUp PageDown once the table has focus, to select it;
// double-click or press Enter to activate it. Only rows near the viewport are
// built, so tables with many thousands of rows stay fast. Columns wider than
// the viewport scroll horizontally together with the header.
//
// Row indexes in callbacks, Value and SetValue are positions in the data given
// to SetRows, whatever the sort order.
type TableView struct {
	rowMenu                 func(int) *MenuView
	cellMenu                func(int, int) *MenuView
	contextMenu             *MenuView
	contextCell             TableCell
	filter                  func([]string) bool
	hasMore, loadRequested  bool
	loadError               string
	onLoadMore              func()
	cellMode                bool
	cells                   map[TableCell]bool
	activeCell, cellAnchor  TableCell
	onCells                 func([]TableCell)
	multi                   bool
	selection               map[int]bool
	anchor                  int
	onSelection             func([]int)
	frozenLeft, frozenRight int
	cols                    []*ColumnSpec
	columns                 []int // display position → source column
	hidden                  []bool
	rows                    [][]string
	order                   []int // display position → data index
	sortCol                 int
	desc                    bool
	selected                int
	empty                   string
	loading                 bool
	disabled                bool
	widths                  []float32 // painted column widths in dp, for resizing
	grab                    float32   // pointer offset inside the resize handle
	list                    *VirtualListView
	reveal                  bool // scroll the selection into view on the next Render
	onChange                func(row int)
	onActive                func(row int)
}

func Table(cols ...*ColumnSpec) *TableView {
	owned := make([]*ColumnSpec, len(cols))
	for i, col := range cols {
		if col == nil {
			panic("kit.Table: nil column")
		}
		copy := *col
		owned[i] = &copy
	}
	v := &TableView{cols: owned, sortCol: -1, selected: -1, anchor: -1, widths: make([]float32, len(cols))}
	v.hidden = make([]bool, len(cols))
	for c := range cols {
		v.columns = append(v.columns, c)
	}
	v.list = VirtualList(0, 40, v.row).ItemKey(func(position int) string { return strconv.Itoa(v.order[position]) })
	return v
}

// FrozenColumns pins the first left and last right columns. Counts are
// clamped; left columns take priority. Flexible frozen columns become 120dp.
func (v *TableView) FrozenColumns(left, right int) *TableView {
	v.frozenLeft = max(0, min(left, len(v.cols)))
	v.frozenRight = max(0, min(right, len(v.cols)-v.frozenLeft))
	v.ensureFrozenWidths()
	return v
}
func (v *TableView) pin(c int, cell *el.DivEl) *el.DivEl {
	columns := v.visibleColumns()
	left, right := v.frozenCounts(columns)
	pos := slices.Index(columns, c)
	if pos < left {
		var offset float32
		for _, i := range columns[:pos] {
			offset += v.cols[i].width
		}
		cell.PinLeft(offset)
	} else if pos >= len(columns)-right {
		var offset float32
		for _, i := range columns[pos+1:] {
			offset += v.cols[i].width
		}
		cell.PinRight(offset)
	}
	return cell
}

// Height sets the height of the rows' viewport in dp, 320 by default; Fill grows instead.
func (v *TableView) Height(dp float32) *TableView           { v.list.Height(dp); return v }
func (v *TableView) Fill() *TableView                       { v.list.Fill(); return v }
func (v *TableView) OnChange(fn func(row int)) *TableView   { v.onChange = fn; return v }
func (v *TableView) OnActivate(fn func(row int)) *TableView { v.onActive = fn; return v }
func (v *TableView) SetDisabled(on bool)                    { v.disabled = on }

// Empty sets the text shown when there are no rows; "" uses the locale's NoData.
func (v *TableView) Empty(s string) *TableView { v.empty = s; return v }

// SetLoading shows a spinner over the rows while data loads elsewhere; deliver
// the rows with core.Update, then SetRows and SetLoading(false).
func (v *TableView) SetLoading(on bool) {
	v.loading = on
	if on {
		v.loadError = ""
	}
}

// SetRows copies the data, keeping the sort column. The selection is
// cleared when its index no longer exists.
func (v *TableView) SetRows(rows [][]string) {
	v.rows = cloneTableRows(rows)
	v.loadRequested = false
	if v.selected >= len(rows) {
		v.selected = -1
	}
	for row := range v.selection {
		if row >= len(rows) {
			delete(v.selection, row)
		}
	}
	if v.anchor >= len(rows) {
		v.anchor = -1
	}
	for cell := range v.cells {
		if cell.Row >= len(rows) {
			delete(v.cells, cell)
		}
	}
	if v.activeCell.Row >= len(rows) {
		v.activeCell = TableCell{-1, -1}
	}
	if v.cellAnchor.Row >= len(rows) {
		v.cellAnchor = TableCell{-1, -1}
	}
	v.resort()
}

// Rows returns a deep copy; use SetRows to replace data and refresh sorting.
func (v *TableView) Rows() [][]string { return cloneTableRows(v.rows) }
func (v *TableView) Len() int         { return len(v.rows) }

// Row returns a copy of row i, or nil.
func (v *TableView) Row(i int) []string {
	if i < 0 || i >= len(v.rows) {
		return nil
	}
	return slices.Clone(v.rows[i])
}

func cloneTableRows(rows [][]string) [][]string {
	out := slices.Clone(rows)
	for i, row := range rows {
		out[i] = slices.Clone(row)
	}
	return out
}

// SetColumnWidth changes this table's column width in dp (minimum 40),
// independently of the ColumnSpec used to construct it. Invalid indexes and
// non-finite widths are ignored. It does not change sorting or selection.
func (v *TableView) SetColumnWidth(column int, dp float32) {
	if column < 0 || column >= len(v.cols) || math.IsNaN(float64(dp)) || math.IsInf(float64(dp), 0) {
		return
	}
	v.cols[column].Width(dp)
}

// Value is the selected row, or -1.
func (v *TableView) Value() int { return v.selected }

// SetValue selects row i (-1 clears) and scrolls it into view, without calling OnChange.
func (v *TableView) SetValue(i int) {
	if v.cellMode {
		columns := v.visibleColumns()
		if len(columns) > 0 {
			v.SetSelectedCells([]TableCell{{i, columns[0]}})
		} else {
			v.SetSelectedCells(nil)
		}
		return
	}
	if i < -1 || i >= len(v.rows) {
		i = -1
	}
	v.selected, v.anchor, v.reveal = i, i, true
	v.selection = make(map[int]bool)
	if i >= 0 {
		v.selection[i] = true
	}
}

// SortBy sorts by column col, descending if desc; col -1 restores data order.
func (v *TableView) SortBy(col int, desc bool) {
	if col >= len(v.cols) {
		col = -1
	}
	v.sortCol, v.desc = col, desc
	v.resort()
}

func (v *TableView) resort() {
	v.order = v.order[:0]
	for i := range v.rows {
		if v.filter == nil || v.filter(slices.Clone(v.rows[i])) {
			v.order = append(v.order, i)
		}
	}
	v.list.SetCount(len(v.order))
	c := v.sortCol
	if c < 0 {
		return
	}
	text := func(i int) string {
		if c < len(v.rows[i]) {
			return v.rows[i][c]
		}
		return ""
	}
	slices.SortStableFunc(v.order, func(a, b int) int {
		x, y := text(a), text(b)
		r := 0
		fx, ex := strconv.ParseFloat(strings.TrimSpace(x), 64)
		fy, ey := strconv.ParseFloat(strings.TrimSpace(y), 64)
		if ex == nil && ey == nil {
			r = cmp.Compare(fx, fy)
		} else {
			r = strings.Compare(x, y)
		}
		if v.desc {
			r = -r
		}
		return r
	})
}

func (v *TableView) position(data int) int {
	for p, i := range v.order {
		if i == data {
			return p
		}
	}
	return -1
}

func (v *TableView) choose(cx *el.Context, data int) {
	v.list.ScrollTo(cx, v.position(data))
	if data == v.selected {
		return
	}
	v.selected = data
	if v.onChange != nil {
		v.onChange(data)
	}
}

func (v *TableView) activate() {
	if v.selected >= 0 && v.onActive != nil {
		v.onActive(v.selected)
	}
}

// sized applies the column's width rule to a header or body cell.
func (v *TableView) sized(c int, cell *el.DivEl) *el.DivEl {
	col := v.cols[c]
	if col.width > 0 {
		cell.W(el.Dp(col.width)).NoShrink()
	} else {
		cell.Flex(col.flex).W(el.Dp(0)).MinW(el.Dp(minColumn))
	}
	if col.numeric {
		cell.Items(el.End)
	}
	return cell
}

func (v *TableView) header(cx *el.Context, c int) el.Element {
	col := v.cols[c]
	label := el.Div().Row().Items(el.Center).Gap(4).Child(el.Text(col.title).Bold().TextSize(theme.TextMd).MaxLines(1))
	if v.sortCol == c {
		arrow := "↑"
		if v.desc {
			arrow = "↓"
		}
		label.Child(el.Text(arrow).TextSize(theme.TextMd).TextColor(theme.PrimaryText))
	}
	cell := v.sized(c, el.Div().Role("columnheader").Name(col.title).Py(10).Px(12).TextColor(theme.Muted)).Child(label)
	cell.Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			v.widths[c] = float32(gtx.Constraints.Max.X) / px
		}
		draw()
	})
	if v.cellMode && !v.disabled {
		cell.CursorPointer().OnClick(func() { v.chooseColumn(cx, c, cx.ClickModifiers()) })
		if !col.noSort {
			cell.OnDoubleClick(func() { v.SortBy(c, v.sortCol == c && !v.desc) })
		}
	} else if !col.noSort && !v.disabled {
		cell.CursorPointer().Focusable(false).OnClick(func() {
			v.SortBy(c, v.sortCol == c && !v.desc)
		})
	}
	// A thin handle on the right edge resizes the column; the column becomes
	// fixed width from then on.
	handle := el.Div().Absolute().Top(0).Bottom(0).Right(0).W(el.Dp(6)).OnDrag(func(e el.DragEvent) {
		if e.Kind == el.DragStart {
			v.grab = e.X
			return
		}
		col.width = max(minColumn, v.widths[c]+e.X-v.grab)
	})
	return v.pin(c, el.Div().ID(autoID("table", v)+"/header/"+strconv.Itoa(c)).Row().Items(el.Stretch).Bg(theme.Subtle).Child(cell, handle)).When(col.width > 0, func(d *el.DivEl) { d.NoShrink() }).
		When(col.width <= 0, func(d *el.DivEl) { d.Flex(col.flex).W(el.Dp(0)).MinW(el.Dp(minColumn)) })
}

func (v *TableView) row(cx *el.Context, p int) el.Element {
	data := v.order[p]
	cells := v.rows[data]
	on := v.rowSelected(data)
	r := el.Div().Role("row").Name(strings.Join(cells, " | ")).Selected(on).Row().Items(el.Center)
	if v.cellMode {
		r.Items(el.Stretch)
	}
	if on {
		r.Bg(theme.Highlight)
	}
	for _, c := range v.visibleColumns() {
		col := v.cols[c]
		cell := v.sized(c, el.Div().ID(v.cellID(data, c)).Px(12).Justify(el.Center))
		if col.width <= 0 {
			cell.Items(el.Start)
			if col.numeric {
				cell.Items(el.End)
			}
		}
		switch {
		case col.cell != nil:
			cell.Child(col.cell(cx, data))
		case c < len(cells):
			cell.Child(el.Text(cells[c]).MaxLines(1))
		}
		if v.cellMode {
			selected := v.cells[TableCell{data, c}]
			cell.Role("gridcell").Name("cell " + strconv.Itoa(data) + "," + strconv.Itoa(c)).Selected(selected)
			if selected {
				cell.Bg(theme.Highlight)
			}
			if !v.disabled {
				cell.OnClick(func() { v.chooseCell(cx, data, c, cx.ClickModifiers()) })
			}
		}
		if v.cellMenu != nil || v.rowMenu != nil {
			cell.OnContextMenu(func() { v.openMenu(cx, data, c) })
		}
		v.pin(c, cell)
		r.Child(cell)
	}
	if !v.disabled && !v.cellMode {
		r.CursorPointer().
			OnClick(func() { v.chooseRows(cx, data, cx.ClickModifiers()) }).
			OnDoubleClick(v.activate)
		if !on {
			r.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	return el.Div().Items(el.Stretch).Child(r.H(el.Dp(39)), el.Div().H(el.Dp(1)).Bg(theme.Border))
}

func (v *TableView) Render(cx *el.Context) el.Element {
	v.loadNearEnd(cx)
	v.renderContextMenu(cx)
	if v.reveal {
		v.list.ScrollTo(cx, v.position(v.selected))
		if v.cellMode {
			v.revealCell(cx)
		}
		v.reveal = false
		if v.cellMode {
			if _, view, _ := cx.ScrollStateX(autoID("table", v)); view == 0 {
				v.reveal = true
				cx.After(revealKey{autoID("table", v)}, 0, func() {})
			}
		}
	}
	head := el.Div().Row().NoShrink().Items(el.Stretch).Bg(theme.Subtle)
	var minWidth float32
	for _, c := range v.visibleColumns() {
		minWidth += max(v.cols[c].width, minColumn)
	}
	for _, c := range v.visibleColumns() {
		head.Child(v.header(cx, c))
	}
	body := el.Div().Items(el.Stretch).When(v.list.fill, func(d *el.DivEl) { d.Grow() }).Child(v.list.Render(cx))
	var status el.Element
	if len(v.order) == 0 && !v.loading && v.loadError == "" {
		empty := v.empty
		if empty == "" {
			empty = locale.Current().NoData
		}
		status = el.Text(empty).TextColor(theme.Muted)
	}
	if v.loadError != "" {
		retry := Button(locale.Current().Retry, v.requestMore)
		retry.SetDisabled(v.onLoadMore == nil)
		status = el.Div().Gap(8).Items(el.Center).Child(el.Text(v.loadError).TextColor(theme.Danger), retry.Render(cx))
	}
	if v.loading {
		status = Spinner().Render(cx)
	}
	content := el.Div().MinW(el.Dp(minWidth)).Items(el.Stretch).
		When(v.list.fill, func(d *el.DivEl) { d.Grow() }).
		Child(head, el.Div().H(el.Dp(1)).NoShrink().Bg(theme.Border), body)
	table := el.Div().ID(autoID("table", v)).ScrollX().Role("table").Value(locale.Current().Rows(len(v.order))).Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() }).
		Rounded(theme.RadiusMd).Border(1, theme.Border).Bg(theme.Surface).Items(el.Stretch).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) == key.NameF10 && e.Modifiers == key.ModShift && (v.rowMenu != nil || v.cellMenu != nil) {
				if e.State == el.KeyPress {
					columns := v.visibleColumns()
					if len(columns) > 0 {
						column := columns[0]
						if v.cellMode && v.activeCell.Column >= 0 && !v.hidden[v.activeCell.Column] {
							column = v.activeCell.Column
						}
						v.openMenu(cx, v.selected, column)
					}
				}
				return true
			}
			if v.selectionKey(e) {
				return true
			}
			if key.Name(e.Name) == key.NameReturn {
				if e.State == el.KeyPress {
					v.activate()
				}
				return true
			}
			if v.cellMode {
				return v.cellKey(cx, e)
			}
			p, ok := base.List{Count: len(v.order), Page: 8}.Key(e.Name, v.position(v.selected))
			if ok && e.State == el.KeyPress && len(v.order) > 0 {
				v.chooseRows(cx, v.order[p], e.Modifiers)
				cx.Focus(autoID("table", v))
			}
			return ok
		}).
		Child(content)
	wrapper := el.Div().Items(el.Stretch).Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() }).Child(table)
	if status == nil {
		return wrapper
	}
	// Status belongs to the viewport, not the horizontally scrolling content.
	overlay := el.Div().Absolute().Top(0).Left(0).Right(0).Bottom(0).Center().Child(status)
	if v.loadError != "" {
		overlay.Bg(theme.Surface).Rounded(theme.RadiusMd).Border(1, theme.Border)
	}
	return wrapper.Child(overlay)
}
