package kit

import (
	"slices"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// BubbleGroupView stacks conversation surfaces. The application decides which
// bubbles belong together; the group does not infer authors or change alignment.
type BubbleGroupView struct {
	items    []el.View
	gap      float32
	name     string
	disabled bool
	style    func(*el.DivEl)
}

// BubbleGroup copies the supplied views, ignoring nil entries. Reuse each view
// instance at most once in a group to preserve its element identity.
func BubbleGroup(items ...el.View) *BubbleGroupView {
	v := &BubbleGroupView{gap: theme.SpaceMd}
	v.SetItems(items...)
	return v
}
func (v *BubbleGroupView) SetItems(items ...el.View) {
	v.items = nil
	for _, item := range items {
		if item != nil {
			v.items = append(v.items, item)
		}
	}
}
func (v *BubbleGroupView) Items() []el.View { return slices.Clone(v.items) }
func (v *BubbleGroupView) Gap(dp float32) *BubbleGroupView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.gap = dp
	}
	return v
}
func (v *BubbleGroupView) Name(name string) *BubbleGroupView { v.name = name; return v }
func (v *BubbleGroupView) SetDisabled(on bool)               { v.disabled = on }

// Style refines a fresh group element after defaults. Nil clears the refinement.
func (v *BubbleGroupView) Style(fn func(*el.DivEl)) *BubbleGroupView { v.style = fn; return v }
func (v *BubbleGroupView) Render(cx *el.Context) el.Element {
	box := el.Div().WFull().MaxW(el.Full).Items(el.Stretch).Gap(v.gap)
	if v.style != nil {
		v.style(box)
	}
	box.ID(autoID("bubble-group", v)).Role("group").Name(v.name).Disabled(v.disabled)
	for _, item := range v.items {
		box.Child(item.Render(cx))
	}
	return box
}
