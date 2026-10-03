package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// AttachmentGroupView arranges attachments in a horizontally scrollable row.
// Upload state, selection, opening and removal remain owned by the application.
type AttachmentGroupView struct {
	items    []el.View
	gap      float32
	disabled bool
	name     string
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

func (v *AttachmentGroupView) Render(cx *el.Context) el.Element {
	row := el.Div().ID(autoID("attachment-group", v)).Role("group").Name(v.name).
		Disabled(v.disabled).W(el.Full).ScrollX().Row().Items(el.Start).Gap(v.gap).Pb(theme.SpaceSm)
	for _, item := range v.items {
		row.Child(el.Div().NoShrink().Child(item.Render(cx)))
	}
	return row
}
