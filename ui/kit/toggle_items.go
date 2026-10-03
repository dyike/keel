package kit

import "slices"

// Item snapshots a Toggle's presentation for an existing option. The group owns
// selected state, IDs and change callbacks; the source Value and OnChange are ignored.
// Explicit item Variant/Size override the group, unset ones inherit it.
// nil clears the override. Unknown options are ignored. Reapply Item after edits.
func (v *ToggleGroupView) Item(option string, toggle *ToggleView) *ToggleGroupView {
	if !slices.Contains(v.options, option) {
		return v
	}
	if toggle == nil {
		delete(v.items, option)
		return v
	}
	if v.items == nil {
		v.items = make(map[string]ToggleView)
	}
	item := *toggle
	if item.icon != nil {
		icon := *item.icon
		item.icon = &icon
	}
	item.onChange = nil
	v.items[option] = item
	return v
}
