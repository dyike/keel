package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// MessageGroupView stacks complete message rows. The application decides which
// messages belong together; the group does not infer authors or change alignment.
type MessageGroupView struct {
	items    []el.View
	gap      float32
	name     string
	disabled bool
	style    func(*el.DivEl)
}

// MessageGroup copies the supplied views, ignoring nil entries. Reuse each view
// instance at most once in a group to preserve its element identity.
func MessageGroup(items ...el.View) *MessageGroupView {
	v := &MessageGroupView{gap: theme.SpaceMd}
	v.SetItems(items...)
	return v
}
func (v *MessageGroupView) SetItems(items ...el.View) {
	v.items = nil
	for _, item := range items {
		if item != nil {
			v.items = append(v.items, item)
		}
	}
}
func (v *MessageGroupView) Items() []el.View { return slices.Clone(v.items) }
func (v *MessageGroupView) Gap(dp float32) *MessageGroupView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.gap = dp
	}
	return v
}
func (v *MessageGroupView) Name(name string) *MessageGroupView { v.name = name; return v }
func (v *MessageGroupView) SetDisabled(on bool)                { v.disabled = on }

// Style refines a fresh group element after defaults. Nil clears the refinement.
func (v *MessageGroupView) Style(fn func(*el.DivEl)) *MessageGroupView { v.style = fn; return v }
func (v *MessageGroupView) Render(cx *el.Context) el.Element {
	box := el.Div().WFull().MaxW(el.Full).Items(el.Stretch).Gap(v.gap)
	if v.style != nil {
		v.style(box)
	}
	box.ID(autoID("message-group", v)).Role("group").Name(v.name).Disabled(v.disabled)
	for _, item := range v.items {
		box.Child(item.Render(cx))
	}
	return box
}
