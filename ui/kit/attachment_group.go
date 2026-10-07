package kit

import (
	"image/color"
	"slices"

	"gioui.org/op"
	"github.com/dyike/keel/ui/core"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// AttachmentGroupView arranges attachments in a horizontally scrollable row.
// Upload state, selection, opening and removal remain owned by the application.
type AttachmentGroupView struct {
	items         []el.View
	gap           float32
	disabled      bool
	name          string
	scrollTarget  float32
	scrollPending bool
	edgeFade      *color.NRGBA
}

// AttachmentGroup copies the items; individual view instances remain shared.
// Nil entries are ignored. Each view instance should occur only once in the group.
func AttachmentGroup(items ...el.View) *AttachmentGroupView {
	v := &AttachmentGroupView{gap: theme.SpaceSm}
	v.SetItems(items...)
	return v
}

func (v *AttachmentGroupView) SetItems(items ...el.View) {
	v.items = nil
	for _, item := range items {
		if item != nil {
			v.items = append(v.items, item)
		}
	}
}
func (v *AttachmentGroupView) Items() []el.View                      { return slices.Clone(v.items) }
func (v *AttachmentGroupView) Name(name string) *AttachmentGroupView { v.name = name; return v }
func (v *AttachmentGroupView) SetDisabled(on bool)                   { v.disabled = on }

// Gap sets the nonnegative spacing in dp; invalid values are ignored.
func (v *AttachmentGroupView) Gap(dp float32) *AttachmentGroupView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.gap = dp
	}
	return v
}

// ScrollTo requests an absolute horizontal offset in dp. It also works before
// first paint; the next painted content clamps it to the available range.
// Negative offsets mean the start. Non-finite values are ignored.
func (v *AttachmentGroupView) ScrollTo(dp float32) {
	if !finiteNumber(float64(dp)) {
		return
	}
	v.scrollTarget = max(0, dp)
	v.scrollPending = true
}

// ScrollState reports offset, viewport width and content width in dp for this root.
// Before first paint the values are zero. Programmatic scrolling does not select items.
func (v *AttachmentGroupView) ScrollState(cx *el.Context) (offset, viewport, content float32) {
	return cx.ScrollStateX(autoID("attachment-group", v))
}

func (v *AttachmentGroupView) Render(cx *el.Context) el.Element {
	row := el.Div().ID(autoID("attachment-group", v)).Role("group").Name(v.name).
		Disabled(v.disabled).W(el.Full).ScrollX().Row().Items(el.Start).Gap(v.gap).Pb(scrollbarGutter)
	for _, item := range v.items {
		row.Child(el.Div().NoShrink().Child(item.Render(cx)))
	}
	return row.Decorate(func(gtx core.C, draw func()) {
		draw()
		v.paintEdgeFade(cx, gtx)
		if !v.scrollPending || !gtx.Enabled() {
			return
		}
		before, view, content := v.ScrollState(cx)
		if view <= 0 {
			return
		}
		target := min(v.scrollTarget, max(0, content-view))
		cx.ScrollIntoViewX(autoID("attachment-group", v), target, target+view)
		v.scrollPending = false
		after, _, _ := v.ScrollState(cx)
		if before != after {
			gtx.Execute(op.InvalidateCmd{})
		}
	})
}
