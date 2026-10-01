package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"strconv"
)

// Description pairs a field name with its display value.
type Description struct{ Label, Text string }
type DescriptionListView struct{ items []Description }

func DescriptionList(items ...Description) *DescriptionListView {
	v := &DescriptionListView{}
	v.SetItems(items...)
	return v
}
func (v *DescriptionListView) SetItems(items ...Description) {
	v.items = append([]Description(nil), items...)
}
func (v *DescriptionListView) Render(*el.Context) el.Element {
	box := el.Div().Role("descriptionlist").Value(strconv.Itoa(len(v.items))).Gap(12)
	for _, item := range v.items {
		box.Child(el.Div().Gap(4).Child(el.Text(item.Label).TextColor(theme.Muted).TextSize(float32(theme.SmallSize)), el.Text(item.Text).TextColor(theme.Text)))
	}
	return box
}
