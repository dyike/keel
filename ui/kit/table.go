package kit

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"gioui.org/io/key"
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
// built, so tables with many thousands of rows stay fast.
//
// Row indexes in callbacks, Value and SetValue are positions in the data given
// to SetRows, whatever the sort order.
type TableView struct {
	cols     []*ColumnSpec
	rows     [][]string
	order    []int // display position → data index
	sortCol  int
	desc     bool
	selected int
	empty    string
	loading  bool
	disabled bool
	widths   []float32 // painted column widths in dp, for resizing
	grab     float32   // pointer offset inside the resize handle
	list     *VirtualListView
	reveal   bool // scroll the selection into view on the next Render
	onChange func(row int)
	onActive func(row int)
}

func Table(cols ...*ColumnSpec) *TableView {
	v := &TableView{cols: cols, sortCol: -1, selected: -1, widths: make([]float32, len(cols))}
	v.list = VirtualList(0, 40, v.row)
	return v
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
func (v *TableView) SetLoading(on bool) { v.loading = on }

// SetRows replaces the data, keeping the sort column. The selection is
// cleared when its index no longer exists.
func (v *TableView) SetRows(rows [][]string) {
	v.rows = rows
	if v.selected >= len(rows) {
		v.selected = -1
	}
	v.resort()
}
func (v *TableView) Rows() [][]string { return v.rows }
func (v *TableView) Len() int         { return len(v.rows) }

// Row returns row i, or nil.
func (v *TableView) Row(i int) []string {
	if i < 0 || i >= len(v.rows) {
		return nil
	}
	return v.rows[i]
}

// Value is the selected row, or -1.
func (v *TableView) Value() int { return v.selected }

// SetValue selects row i (-1 clears) and scrolls it into view, without calling OnChange.
func (v *TableView) SetValue(i int) {
	if i < -1 || i >= len(v.rows) {
		i = -1
	}
	v.selected, v.reveal = i, true
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
		v.order = append(v.order, i)
	}
	v.list.SetCount(len(v.rows))
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
	label := el.Div().Row().Items(el.Center).Gap(4).Child(el.Text(col.title).Bold().TextSize(13).MaxLines(1))
	if v.sortCol == c {
		arrow := "↑"
		if v.desc {
			arrow = "↓"
		}
		label.Child(el.Text(arrow).TextSize(13).TextColor(theme.PrimaryText))
	}
	cell := v.sized(c, el.Div().Role("columnheader").Name(col.title).Py(10).Px(12).TextColor(theme.Muted)).Child(label)
	cell.Decorate(func(gtx core.C, draw func()) {
		if px := gtx.Metric.PxPerDp; px > 0 {
			v.widths[c] = float32(gtx.Constraints.Max.X) / px
		}
		draw()
	})
	if !col.noSort && !v.disabled {
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
	return el.Div().Row().Items(el.Stretch).Child(cell, handle).When(col.width > 0, func(d *el.DivEl) { d.NoShrink() }).
		When(col.width <= 0, func(d *el.DivEl) { d.Flex(col.flex).W(el.Dp(0)).MinW(el.Dp(minColumn)) })
}

func (v *TableView) row(cx *el.Context, p int) el.Element {
	data := v.order[p]
	cells := v.rows[data]
	on := data == v.selected
	r := el.Div().Role("row").Name(strings.Join(cells, " | ")).Selected(on).Row().Items(el.Center)
	if on {
		r.Bg(theme.Highlight)
	}
	for c, col := range v.cols {
		cell := v.sized(c, el.Div().Px(12).Justify(el.Center))
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
		r.Child(cell)
	}
	if !v.disabled {
		r.CursorPointer().
			OnClick(func() { v.choose(cx, data); cx.Focus(autoID("table", v)) }).
			OnDoubleClick(v.activate)
		if !on {
			r.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	return el.Div().Items(el.Stretch).Child(r.H(el.Dp(39)), el.Div().H(el.Dp(1)).Bg(theme.Border))
}

func (v *TableView) Render(cx *el.Context) el.Element {
	if v.reveal {
		v.list.ScrollTo(cx, v.position(v.selected))
		v.reveal = false
	}
	head := el.Div().Row().Items(el.Stretch).Bg(theme.Subtle)
	for c := range v.cols {
		head.Child(v.header(cx, c))
	}
	body := el.Div().Items(el.Stretch).Child(v.list.Render(cx))
	if len(v.rows) == 0 && !v.loading {
		empty := v.empty
		if empty == "" {
			empty = locale.Current().NoData
		}
		body.Child(el.Div().Absolute().Top(0).Left(0).Right(0).Bottom(0).Center().Child(el.Text(empty).TextColor(theme.Muted)))
	}
	if v.loading {
		body.Child(el.Div().Absolute().Top(0).Left(0).Right(0).Bottom(0).Center().Child(Spinner().Render(cx)))
	}
	return el.Div().ID(autoID("table", v)).Role("table").Value(locale.Current().Rows(len(v.rows))).Disabled(v.disabled).
		Rounded(6).Border(1, theme.Border).Bg(theme.Surface).Items(el.Stretch).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) == key.NameReturn {
				if e.State == el.KeyPress {
					v.activate()
				}
				return true
			}
			p, ok := listKeys(e.Name, v.position(v.selected), len(v.order), 8)
			if ok && e.State == el.KeyPress && len(v.order) > 0 {
				v.choose(cx, v.order[p])
			}
			return ok
		}).
		Child(head, el.Div().H(el.Dp(1)).Bg(theme.Border), body)
}
