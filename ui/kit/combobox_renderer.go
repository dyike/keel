package kit

import "github.com/dyike/keel/ui/el"

// RenderItem replaces candidate label content, leaving selection, disabled
// state and the trailing check under Combobox control. Nil callbacks or nil
// results restore the label. Only visible virtual rows invoke the callback.
// Reuse stateful Views by stable item Value; do not mutate selection in Render.
func (v *ComboboxView) RenderItem(fn func(ComboboxItem, bool) el.View) *ComboboxView {
	v.itemRenderer = fn
	return v
}

// RowHeight sets the full candidate slot height in dp, including row margins.
// All candidates use this height. Zero restores 30dp scaled with Size;
// positive explicit heights do not scale with Size. Invalid values are ignored.
func (v *ComboboxView) RowHeight(dp float32) *ComboboxView {
	if dp >= 0 && finiteNumber(float64(dp)) && v.rowHeight != dp {
		v.rowHeight = dp
		if v.active >= 0 {
			v.virtual.reveal = v.displayIndex(v.active)
		}
	}
	return v
}
func (v *ComboboxView) optionHeight() float32 {
	if v.rowHeight > 0 {
		return max(1+2*v.sizeRatio(), v.rowHeight)
	}
	return max(1+2*v.sizeRatio(), 30*v.sizeRatio())
}
func (v *ComboboxView) renderItem(cx *el.Context, value string, selected bool) el.Element {
	var content el.View
	if v.itemRenderer != nil {
		content = v.itemRenderer(ComboboxItem{Value: value, Label: v.optionLabel(value), Disabled: v.optionDisabled(value)}, selected)
	}
	if content == nil {
		return el.Div().ID("content").Grow().MinW(el.Dp(0)).Child(el.Text(v.optionLabel(value)).MaxLines(1))
	}
	return el.Div().ID("content").Grow().MinW(el.Dp(0)).Child(content.Render(cx))
}
