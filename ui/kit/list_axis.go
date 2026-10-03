package kit

import "github.com/dyike/keel/ui/el"

func virtualState(cx *el.Context, id string, horizontal bool) (float32, float32, float32) {
	if horizontal {
		return cx.ScrollStateX(id)
	}
	return cx.ScrollState(id)
}
func virtualScroll(cx *el.Context, id string, horizontal bool, offset float32) {
	if horizontal {
		cx.ScrollToX(id, offset)
	} else {
		cx.ScrollTo(id, offset)
	}
}
func virtualReveal(cx *el.Context, id string, horizontal bool, start, end float32) {
	if horizontal {
		cx.ScrollIntoViewX(id, start, end)
	} else {
		cx.ScrollIntoView(id, start, end)
	}
}
func virtualSpacer(horizontal bool, size float32) *el.DivEl {
	d := el.Div().NoShrink()
	if horizontal {
		return d.W(el.Dp(size))
	}
	return d.H(el.Dp(size))
}
func virtualExtent(horizontal bool, width, height float32) float32 {
	if horizontal {
		if width > 0 {
			return width
		}
		return 320
	}
	return height
}

// Horizontal switches the virtualization axis. Item size is the constructor's
// rowHeight on either axis. It preserves the current leading item's offset;
// a pending ScrollTo takes precedence. False restores vertical rendering.
func (v *VirtualListView) Horizontal(on bool) *VirtualListView {
	if v.horizontal != on {
		v.horizontal = on
		v.axisChanged = true
		v.axisOffset = v.lastOffset
	}
	return v
}

// Width sets viewport width in dp. In horizontal mode it replaces Fill; in
// vertical mode it constrains the cross axis. Horizontal default width is 320.
func (v *VirtualListView) Width(dp float32) *VirtualListView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.width = dp
		if v.horizontal {
			v.fill = false
		}
	}
	return v
}

// ScrollToEnd reveals the last item on the current axis.
func (v *VirtualListView) ScrollToEnd(cx *el.Context) { v.ScrollTo(cx, v.count-1) }

// Horizontal switches the virtualization axis, keeping the leading stable key
// and discarding measurements from the old axis. Pending reveals are retained.
func (v *VariableListView) Horizontal(on bool) *VariableListView {
	if v.horizontal != on {
		v.horizontal = on
		v.anchorDelta = 0
		v.width = 0
		v.Invalidate()
	}
	return v
}

// Width sets viewport width. In horizontal mode it replaces Fill; in vertical
// mode it constrains the cross axis. Horizontal default width is 320dp.
func (v *VariableListView) Width(dp float32) *VariableListView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.viewportWidth = dp
		if v.horizontal {
			v.fill = false
		}
	}
	return v
}

// ScrollToEnd reveals the final stable key on the current axis.
func (v *VariableListView) ScrollToEnd(cx *el.Context) { v.ScrollTo(cx, len(v.keys)-1) }
