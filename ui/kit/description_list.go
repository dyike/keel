package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Description pairs a label with text; use ItemView for rich values.
type Description struct{ Label, Text string }
type descriptionItem struct {
	Description
	view el.View
}
type DescriptionListView struct {
	items      []descriptionItem
	labelWidth float32
}

func DescriptionList() *DescriptionListView { return &DescriptionListView{labelWidth: 96} }
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
func (v *DescriptionListView) ItemView(label string, value el.View) *DescriptionListView {
	v.items = append(v.items, descriptionItem{Description: Description{Label: label}, view: value})
	return v
}
func (v *DescriptionListView) LabelWidth(dp float32) *DescriptionListView {
	if dp >= 0 {
		v.labelWidth = dp
	}
	return v
}
func (v *DescriptionListView) Render(cx *el.Context) el.Element {
	box := el.Div().W(el.Full).Gap(12)
	for _, item := range v.items {
		row := el.Div().W(el.Full).Row().Gap(12).Items(el.Start)
		label := el.Text(item.Label).W(el.Dp(v.labelWidth)).TextColor(theme.Muted)
		if item.view == nil {
			row.Role("text").Name(item.Label+"："+item.Text).Child(label, el.Text(item.Text).Grow().TextColor(theme.Text))
		} else {
			row.Child(label, el.Div().Grow().Child(item.view.Render(cx)))
		}
		box.Child(row)
	}
	return box
}
