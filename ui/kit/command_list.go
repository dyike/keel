package kit

import "github.com/dyike/keel/ui/el"

// AutoRowHeight measures custom row contents and virtualizes using cached
// heights. RowHeight becomes the minimum slot height and initial estimate.
// False restores uniform slots. Width/scale and model changes invalidate the
// measurements; offscreen content changes can call InvalidateRows.
func (v *CommandView) AutoRowHeight(on bool) *CommandView {
	if v.autoRows != on {
		v.autoRows = on
		v.variable.Invalidate()
		v.revealActive = true
	}
	return v
}

// InvalidateRows discards measured heights after application-owned offscreen
// content or typography changes. Visible rows are measured every frame.
func (v *CommandView) InvalidateRows() { v.variable.Invalidate(); v.revealActive = true }

func (v *CommandView) listID() string {
	if v.autoRows {
		return v.variable.ID()
	}
	return v.list.ID()
}
func (v *CommandView) scrollTo(cx *el.Context, i int) {
	if v.autoRows {
		v.variable.ScrollTo(cx, i)
	} else {
		v.list.ScrollTo(cx, i)
	}
}
func (v *CommandView) renderList(cx *el.Context, viewport float32) el.Element {
	if v.autoRows {
		height := max(1, v.variable.sums.prefix(len(v.rows)))
		return v.variable.Height(min(viewport, height)).Render(cx)
	}
	v.list.SetCount(len(v.rows))
	return v.list.Height(min(viewport, float32(len(v.rows))*v.itemHeight())).Render(cx)
}
