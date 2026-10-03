package kit

import "github.com/dyike/keel/ui/el"

func (v *TabsView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.dragID = 0
		v.more.SetValue(false)
	}
}

// Reorderable enables dragging visible tabs; Move can reorder hidden tabs.
func (v *TabsView) Reorderable(fn func(from, to int)) *TabsView {
	v.reorder = true
	v.onMove = fn
	return v
}

// Move preserves the selected page and does not invoke callbacks.
func (v *TabsView) Move(from, to int) {
	if from < 0 || to < 0 || from >= len(v.pages) || to >= len(v.pages) || from == to {
		return
	}
	selected := v.pages[v.current].id
	page, width := v.pages[from], v.widths[from]
	if from < to {
		copy(v.pages[from:to], v.pages[from+1:to+1])
		copy(v.widths[from:to], v.widths[from+1:to+1])
	} else {
		copy(v.pages[to+1:from+1], v.pages[to:from])
		copy(v.widths[to+1:from+1], v.widths[to:from])
	}
	v.pages[to], v.widths[to] = page, width
	for i, p := range v.pages {
		if p.id == selected {
			v.current = i
			break
		}
	}
}
func (v *TabsView) dragTab(cx *el.Context, i int, visible []int, e el.DragEvent) {
	if e.Kind == el.DragStart {
		v.dragID = v.pages[i].id
		return
	}
	if e.Kind != el.DragEnd {
		return
	}
	id := v.dragID
	v.dragID = 0
	if e.Canceled || v.disabled || id == 0 {
		return
	}
	from := -1
	for j, p := range v.pages {
		if p.id == id {
			from = j
			break
		}
	}
	if from < 0 || v.pages[from].Disabled {
		return
	}
	x := v.offsets[from] + e.X
	to := visible[0]
	for _, j := range visible {
		if x >= v.offsets[j] {
			to = j
		}
	}
	if from == to {
		return
	}
	v.Move(from, to)
	v.focusPending = true
	if v.onMove != nil {
		v.onMove(from, to)
	}
}

func (v *TabsView) Leading(view el.View) *TabsView  { v.leading = view; return v }
func (v *TabsView) Trailing(view el.View) *TabsView { v.trailing = view; return v }
func (v *TabsView) Size(height float32) *TabsView {
	if height >= 24 && finiteNumber(float64(height)) {
		v.height = height
	}
	return v
}
