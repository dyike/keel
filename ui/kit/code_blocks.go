package kit

import (
	"image"
	"sort"

	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/ui/core"
)

// Private source-anchored rows for composite editors. They never enter the
// text buffer or keyboard row navigation. Nil widgets reserve matching space.
type codeBlock struct {
	line   int
	widget core.Widget
}
type codeBlockRow struct{ index, row, extra int }
type codeBlocks struct {
	items                []codeBlock
	rows                 []codeBlockRow
	valid                bool
	generation           uint64
	wrapKey              codeWrapKey
	wrap                 bool
	height, cachedHeight int
}

func (v *CodeEditorView) setBlocks(items []codeBlock) {
	v.blocks.items = items
	v.blocks.valid = false
}
func (v *CodeEditorView) ensureBlocks() {
	b := &v.blocks
	if len(b.items) == 0 {
		b.rows = b.rows[:0]
		return
	}
	v.buildRows()
	v.buildVRows()
	if b.valid && b.generation == v.rowsGen && b.wrapKey == v.vrowsKey && b.wrap == v.wrap && b.cachedHeight == b.height {
		return
	}
	b.valid, b.generation, b.wrapKey, b.wrap, b.cachedHeight = true, v.rowsGen, v.vrowsKey, v.wrap, b.height
	b.rows = b.rows[:0]
	for i, item := range b.items {
		line := min(max(item.line, 0), v.buf.count()-1)
		if v.hidden(line) {
			continue
		}
		row := v.vrowOf(codePos{line, 0})
		b.rows = append(b.rows, codeBlockRow{index: i, row: row})
	}
	sort.SliceStable(b.rows, func(i, j int) bool { return b.rows[i].row < b.rows[j].row })
	for i := range b.rows {
		b.rows[i].extra = (i + 1) * b.height
	}
}
func (v *CodeEditorView) rowTop(row int) int {
	v.ensureBlocks()
	i := sort.Search(len(v.blocks.rows), func(i int) bool { return v.blocks.rows[i].row > row })
	extra := 0
	if i > 0 {
		extra = v.blocks.rows[i-1].extra
	}
	return row*v.metrics.lh + extra
}
func (v *CodeEditorView) rowAt(y float32) int {
	if len(v.blocks.items) == 0 {
		return min(max(int(y)/max(1, v.metrics.lh), 0), v.vrowCount()-1)
	}
	return min(sort.Search(v.vrowCount(), func(row int) bool { return float32(v.rowTop(row)+v.metrics.lh) > y }), v.vrowCount()-1)
}
func (v *CodeEditorView) contentHeight() int { return v.rowTop(v.vrowCount()) }
func (v *CodeEditorView) inBlock(y float32) bool {
	if len(v.blocks.items) == 0 {
		return false
	}
	row := v.rowAt(y)
	return y < float32(v.rowTop(row))
}
func (v *CodeEditorView) paintBlocks(gtx core.C) {
	v.ensureBlocks()
	first := sort.Search(len(v.blocks.rows), func(i int) bool {
		b := v.blocks.rows[i]
		return float32(b.row*v.metrics.lh+(i+1)*v.blocks.height) > v.scrollY
	})
	for i := first; i < len(v.blocks.rows); i++ {
		b := v.blocks.rows[i]
		top := b.row*v.metrics.lh + i*v.blocks.height - int(v.scrollY)
		if top >= v.metrics.size.Y {
			break
		}
		if widget := v.blocks.items[b.index].widget; widget != nil {
			offset := op.Offset(image.Pt(0, top)).Push(gtx.Ops)
			size := image.Pt(v.metrics.size.X, v.blocks.height)
			area := clip.Rect{Max: size}.Push(gtx.Ops)
			ctx := gtx
			ctx.Constraints = layout.Exact(size)
			widget.Layout(ctx)
			area.Pop()
			offset.Pop()
		}
	}
}

// rowStart includes controls preceding this row.
func (v *CodeEditorView) rowStart(row int) int {
	top := v.rowTop(row)
	first := sort.Search(len(v.blocks.rows), func(i int) bool { return v.blocks.rows[i].row >= row })
	last := sort.Search(len(v.blocks.rows), func(i int) bool { return v.blocks.rows[i].row > row })
	top -= (last - first) * v.blocks.height
	return top
}
