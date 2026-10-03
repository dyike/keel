package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Description pairs a label with text; use ItemView for rich values.
type Description struct{ Label, Text string }
type descriptionItem struct {
	Description
	view      el.View
	span      int
	separator bool
}
type DescriptionListView struct {
	items              []descriptionItem
	labelWidth         float32
	columns            int
	vertical, bordered bool
	size               float32
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
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.labelWidth = dp
	}
	return v
}

// Columns sets the number of equal-width item columns (at least one).
func (v *DescriptionListView) Columns(n int) *DescriptionListView { v.columns = max(1, n); return v }

// Span makes the most recently added item occupy n columns, clamped at render time.
func (v *DescriptionListView) Span(n int) *DescriptionListView {
	if len(v.items) > 0 && !v.items[len(v.items)-1].separator {
		v.items[len(v.items)-1].span = max(1, n)
	}
	return v
}

// Vertical places each label above its value.
func (v *DescriptionListView) Vertical() *DescriptionListView { v.vertical = true; return v }

// Separator inserts a full-width divider and starts subsequent items on a new row.
func (v *DescriptionListView) Separator() *DescriptionListView {
	v.items = append(v.items, descriptionItem{separator: true})
	return v
}

// Bordered draws a padded border around each item; the default is unbordered.
func (v *DescriptionListView) Bordered(on bool) *DescriptionListView { v.bordered = on; return v }

// Size sets text size in sp; use theme.TextSm/TextBody/TextLg for size tiers.
func (v *DescriptionListView) Size(sp float32) *DescriptionListView {
	if sp > 0 && finiteNumber(float64(sp)) {
		v.size = sp
	}
	return v
}
func (v *DescriptionListView) Render(cx *el.Context) el.Element {
	columns := max(1, v.columns)
	gap := float32(theme.SpaceLg)
	if v.size > 0 && v.size <= theme.TextMd {
		gap = theme.SpaceMd
	} else if v.size >= theme.TextLg {
		gap = theme.SpaceXl
	}
	box := el.Div().W(el.Full).Grid(columns).Gap(gap).Items(el.Stretch)
	if v.size > 0 {
		box.TextSize(v.size)
	}
	for _, item := range v.items {
		if item.separator {
			box.Child(el.Div().ColSpan(columns).H(el.Dp(1)).Bg(theme.Border).Role("separator"))
			continue
		}
		row := el.Div().W(el.Full).ColSpan(item.span).Gap(gap).Items(el.Start)
		label := el.Text(item.Label).TextColor(theme.Muted)
		if v.vertical {
			label.WFull()
		} else {
			row.Row()
			label.W(el.Dp(v.labelWidth))
		}
		if v.bordered {
			row.Border(1, theme.Border).Rounded(theme.RadiusMd).P(gap)
		}
		if item.view == nil {
			value := el.Text(item.Text).TextColor(theme.Text)
			if v.vertical {
				value.WFull()
			} else {
				value.Grow()
			}
			row.Role("text").Name(item.Label+"："+item.Text).Child(label, value)
		} else {
			value := el.Div().Child(item.view.Render(cx))
			if v.vertical {
				value.WFull()
			} else {
				value.Grow()
			}
			row.Child(label, value)
		}
		box.Child(row)
	}
	return box
}
