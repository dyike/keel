package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Description pairs a label with text; use ItemElement for rich values.
type Description struct{ Label, Text string }
type descriptionItem struct {
	Description
	element el.Element
}
type DescriptionListView struct {
	items      []descriptionItem
	labelWidth float32
}

func DescriptionList(items ...Description) *DescriptionListView {
	v := &DescriptionListView{labelWidth: 96}
	v.SetItems(items...)
	return v
}
func (v *DescriptionListView) SetItems(items ...Description) {
	v.items = nil
	for _, i := range items {
		v.Item(i.Label, i.Text)
	}
}
func (v *DescriptionListView) Item(label, value string) *DescriptionListView {
	v.items = append(v.items, descriptionItem{Description: Description{label, value}})
	return v
}
func (v *DescriptionListView) ItemElement(label string, value el.Element) *DescriptionListView {
	v.items = append(v.items, descriptionItem{Description: Description{Label: label}, element: value})
	return v
}
func (v *DescriptionListView) LabelWidth(dp float32) *DescriptionListView {
	if dp >= 0 {
		v.labelWidth = dp
	}
	return v
}
func (v *DescriptionListView) Render(*el.Context) el.Element {
	box := el.Div().W(el.Full).Gap(12)
	for _, item := range v.items {
		row := el.Div().W(el.Full).Row().Gap(12).Items(el.Start)
		label := el.Text(item.Label).W(el.Dp(v.labelWidth)).TextColor(theme.Muted)
		if item.element == nil {
			row.Role("text").Name(item.Label+"："+item.Text).Child(label, el.Text(item.Text).Grow().TextColor(theme.Text))
		} else {
			row.Child(label, el.Div().Grow().Child(item.element))
		}
		box.Child(row)
	}
	return box
}
