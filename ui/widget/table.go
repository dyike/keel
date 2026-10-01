package widget

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
)

// Column describes a table column. Width is a weight: a column of width 2 gets
// twice the space of a column of width 1. Zero means 1.
type Column struct {
	Title string
	Width float32
}

// Col is shorthand for Column{title, width}.
func Col(title string, width float32) Column { return Column{Title: title, Width: width} }

// TableView shows rows of text in columns. Click a header to sort by it (again
// to reverse); click a row, or use ↑ ↓ Home End once the table has focus, to
// select it; double-click or press Enter to activate it. Rows scroll inside a
// fixed height and are only laid out while visible, so large tables stay fast.
type TableView struct {
	cols       []Column
	rows       [][]string
	order      []int // row indexes in display order
	sortCol    int   // -1: unsorted
	desc       bool
	selected   int // row index, -1: none
	height     unit.Dp
	empty      string
	onSelect   func(row int)
	onActivate func(row int)

	list     widget.List
	reveal0  bool // scroll the selection into view at the next layout
	heads    []widget.Clickable
	clicks   []widget.Clickable // per row index
	focused  bool
	disabled bool
}

// Table creates an empty table with the given columns, 320dp tall.
func Table(cols ...Column) *TableView {
	t := &TableView{cols: cols, sortCol: -1, selected: -1, height: 320, empty: "暂无数据"}
	t.heads = make([]widget.Clickable, len(cols))
	t.list.Axis = giolayout.Vertical
	return t
}

func (t *TableView) Height(dp int) *TableView               { t.height = unit.Dp(dp); return t }
func (t *TableView) Empty(text string) *TableView           { t.empty = text; return t }
func (t *TableView) OnSelect(fn func(row int)) *TableView   { t.onSelect = fn; return t }
func (t *TableView) OnActivate(fn func(row int)) *TableView { t.onActivate = fn; return t }

// SetRows replaces the data, keeping the sort column. Rows are indexed in the
// order given here; callbacks and Selected use those indexes.
func (t *TableView) SetRows(rows [][]string) {
	t.rows = rows
	if len(t.clicks) < len(rows) {
		t.clicks = append(t.clicks, make([]widget.Clickable, len(rows)-len(t.clicks))...)
	}
	if t.selected >= len(rows) {
		t.selected = -1
	}
	t.resort()
}

func (t *TableView) Rows() [][]string { return t.rows }
func (t *TableView) Len() int         { return len(t.rows) }

// Row returns row i, or nil if out of range.
func (t *TableView) Row(i int) []string {
	if i < 0 || i >= len(t.rows) {
		return nil
	}
	return t.rows[i]
}

// Selected returns the selected row index, or -1.
func (t *TableView) Selected() int { return t.selected }
func (t *TableView) SetDisabled(v bool) {
	t.disabled = v
	if v {
		t.focused = false
	}
}

// SetSelected selects row i (-1 clears) and scrolls it into view, without
// calling OnSelect.
func (t *TableView) SetSelected(i int) {
	if i < -1 || i >= len(t.rows) {
		i = -1
	}
	t.selected = i
	t.reveal0 = i >= 0
}

// SortBy sorts by column col, descending if desc; col -1 restores data order.
func (t *TableView) SortBy(col int, desc bool) {
	t.sortCol, t.desc = col, desc
	t.resort()
}

// resort orders rows by the sort column, comparing numbers as numbers.
func (t *TableView) resort() {
	t.order = t.order[:0]
	for i := range t.rows {
		t.order = append(t.order, i)
	}
	c := t.sortCol
	if c < 0 || c >= len(t.cols) {
		return
	}
	cell := func(i int) string {
		if c < len(t.rows[i]) {
			return t.rows[i][c]
		}
		return ""
	}
	slices.SortStableFunc(t.order, func(a, b int) int {
		x, y := cell(a), cell(b)
		r := 0
		if fx, err1 := strconv.ParseFloat(strings.ReplaceAll(x, ",", ""), 64); err1 == nil {
			if fy, err2 := strconv.ParseFloat(strings.ReplaceAll(y, ",", ""), 64); err2 == nil {
				r = cmp.Compare(fx, fy)
			}
		}
		if r == 0 {
			r = strings.Compare(x, y)
		}
		if t.desc {
			r = -r
		}
		return r
	})
}

func (t *TableView) Layout(gtx C) D {
	if t.disabled {
		gtx = gtx.Disabled()
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	t.handleKeys(gtx)
	for i := range t.heads {
		for t.heads[i].Clicked(gtx) {
			if t.sortCol == i {
				t.desc = !t.desc
			} else {
				t.sortCol, t.desc = i, false
			}
			t.resort()
			core.Call(gtx, nil) // redraw other windows showing this table
		}
	}
	for pos, i := range t.order {
		for {
			c, ok := t.clicks[i].Update(gtx)
			if !ok {
				break
			}
			gtx.Execute(key.FocusCmd{Tag: t})
			t.choose(gtx, pos)
			if c.NumClicks >= 2 && t.onActivate != nil {
				row := i
				core.Call(gtx, func() { t.onActivate(row) })
			}
		}
	}
	border := theme.Border
	if t.focused {
		border = theme.Primary
	}
	summary := fmt.Sprintf("%d 行", len(t.rows))
	return core.Semantic(gtx, func(gtx C) D {
		return layout.Frame(gtx, theme.Surface, border, 6, giolayout.Inset{Top: 1, Bottom: 1, Left: 1, Right: 1}, func(gtx C) D {
			return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx,
				giolayout.Rigid(t.header),
				giolayout.Rigid(layout.Divider().Layout),
				giolayout.Rigid(t.body),
			)
		})
	}, core.Role("table", summary), semantic.EnabledOp(gtx.Enabled()))
}

// handleKeys moves the selection with the keyboard while the table has focus.
func (t *TableView) handleKeys(gtx C) {
	filters := []event.Filter{key.FocusFilter{Target: t}}
	for _, n := range []key.Name{key.NameUpArrow, key.NameDownArrow, key.NameHome, key.NameEnd,
		key.NamePageUp, key.NamePageDown, key.NameReturn, key.NameEnter} {
		filters = append(filters, key.Filter{Focus: t, Name: n})
	}
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		switch e := ev.(type) {
		case key.FocusEvent:
			t.focused = e.Focus
		case key.Event:
			if e.State != key.Press || len(t.order) == 0 {
				continue
			}
			pos := slices.Index(t.order, t.selected)
			page := max(int(t.height/40)-1, 1) // rows are about 40dp
			switch e.Name {
			case key.NameUpArrow:
				pos = max(pos-1, 0)
			case key.NameDownArrow:
				pos = min(pos+1, len(t.order)-1)
			case key.NamePageUp:
				pos = max(pos-page, 0)
			case key.NamePageDown:
				pos = min(max(pos, 0)+page, len(t.order)-1)
			case key.NameHome:
				pos = 0
			case key.NameEnd:
				pos = len(t.order) - 1
			case key.NameReturn, key.NameEnter:
				if t.selected >= 0 && t.onActivate != nil {
					row := t.selected
					core.Call(gtx, func() { t.onActivate(row) })
				}
				continue
			}
			t.choose(gtx, pos)
			t.reveal(pos)
		}
	}
}

// choose selects the row at display position pos and reports it.
func (t *TableView) choose(gtx C, pos int) {
	if pos < 0 || pos >= len(t.order) || t.order[pos] == t.selected {
		return
	}
	t.selected = t.order[pos]
	row := t.selected
	core.Call(gtx, func() {
		if t.onSelect != nil {
			t.onSelect(row)
		}
	})
}

// reveal scrolls so that display position pos is visible.
func (t *TableView) reveal(pos int) {
	first := t.list.Position.First
	visible := t.list.Position.Count
	if visible == 0 {
		return
	}
	if pos < first {
		t.list.ScrollTo(pos)
	} else if pos >= first+visible-1 {
		t.list.ScrollTo(max(pos-visible+2, 0))
	}
}

func (t *TableView) weights() []float32 {
	w := make([]float32, len(t.cols))
	for i, c := range t.cols {
		w[i] = c.Width
		if w[i] <= 0 {
			w[i] = 1
		}
	}
	return w
}

func (t *TableView) header(gtx C) D {
	weights := t.weights()
	cells := make([]giolayout.FlexChild, len(t.cols))
	for i, c := range t.cols {
		title := c.Title
		if t.sortCol == i {
			title += map[bool]string{false: " ↑", true: " ↓"}[t.desc]
		}
		cells[i] = giolayout.Flexed(weights[i], func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return t.heads[i].Layout(gtx, func(gtx C) D {
				pointer.CursorPointer.Add(gtx.Ops)
				return core.Semantic(gtx, func(gtx C) D {
					return cellInset.Layout(gtx, func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						lb := material.Label(theme.Material, theme.SmallSize, title)
						lb.Color, lb.Font.Weight, lb.MaxLines = theme.Muted, font.Bold, 1
						return lb.Layout(gtx)
					})
				}, semantic.Button, semantic.LabelOp(c.Title), core.Role("columnheader"), semantic.EnabledOp(gtx.Enabled()))
			})
		})
	}
	return giolayout.Flex{}.Layout(gtx, cells...)
}

var cellInset = giolayout.Inset{Top: 8 + theme.CJKNudge, Bottom: 8 - theme.CJKNudge, Left: 12, Right: 12}

func (t *TableView) body(gtx C) D {
	gtx.Constraints.Min.Y = gtx.Dp(t.height)
	gtx.Constraints.Max.Y = gtx.Constraints.Min.Y
	// Register the table as a key handler so it can take focus. Its own area:
	// on the table's semantic area it would hide that node from agents.
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	event.Op(gtx.Ops, t)
	if len(t.order) == 0 {
		return giolayout.Center.Layout(gtx, Muted(t.empty).Layout)
	}
	if t.reveal0 {
		t.reveal0 = false
		t.list.ScrollTo(max(slices.Index(t.order, t.selected)-2, 0))
	}
	weights := t.weights()
	lst := material.List(theme.Material, &t.list)
	lst.AnchorStrategy = material.Overlay // take no width, so columns line up with the header
	return lst.Layout(gtx, len(t.order), func(gtx C, pos int) D {
		i := t.order[pos]
		row := t.rows[i]
		selected := i == t.selected
		return t.clicks[i].Layout(gtx, func(gtx C) D {
			return core.Semantic(gtx, func(gtx C) D {
				cells := make([]giolayout.FlexChild, len(t.cols))
				for c := range t.cols {
					text := ""
					if c < len(row) {
						text = row[c]
					}
					cells[c] = giolayout.Flexed(weights[c], func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return cellInset.Layout(gtx, func(gtx C) D {
							lb := material.Label(theme.Material, theme.BodySize, text)
							lb.Color, lb.MaxLines = theme.Text, 1
							if !gtx.Enabled() {
								lb.Color = theme.Muted
							}
							return lb.Layout(gtx)
						})
					})
				}
				// Lay out the cells first so the highlight matches the row's real height.
				m := op.Record(gtx.Ops)
				d := giolayout.Flex{}.Layout(gtx, cells...)
				content := m.Stop()
				if selected {
					paint.FillShape(gtx.Ops, theme.Highlight, clip.Rect{Max: d.Size}.Op())
				}
				content.Add(gtx.Ops)
				return d
			}, core.Role("row"), semantic.LabelOp(strings.Join(row, " | ")), semantic.SelectedOp(selected), semantic.EnabledOp(gtx.Enabled()))
		})
	})
}
