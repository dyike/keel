package kit

import (
	"slices"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ToggleGroupView is a row of toggles: one choice at a time (the default) or,
// with Multiple, any number. Value lists the pressed options in option order.
type ToggleGroupView struct {
	options            []string
	value              []string
	multiple, disabled bool
	onChange           func([]string)
	appearance         toggleAppearance
	segmented          bool
	gap                *float32
	items              map[string]ToggleView
}

func ToggleGroup(options ...string) *ToggleGroupView {
	return &ToggleGroupView{options: slices.Clone(options)}
}
func (v *ToggleGroupView) Multiple() *ToggleGroupView { v.multiple = true; return v }
func (v *ToggleGroupView) OnChange(fn func(values []string)) *ToggleGroupView {
	v.onChange = fn
	return v
}

// Value returns a copy of the pressed options.
func (v *ToggleGroupView) Value() []string { return slices.Clone(v.value) }

// SetValue presses exactly these options without calling OnChange; a single
// group keeps only the first.
func (v *ToggleGroupView) SetValue(values ...string) {
	v.value = nil
	for _, o := range v.options {
		if slices.Contains(values, o) && (v.multiple || len(v.value) == 0) {
			v.value = append(v.value, o)
		}
	}
}
func (v *ToggleGroupView) SetDisabled(on bool) { v.disabled = on }

func (v *ToggleGroupView) toggle(o string) {
	on := slices.Contains(v.value, o)
	switch {
	case !v.multiple && on:
		v.value = nil
	case !v.multiple:
		v.value = []string{o}
	case on:
		v.value = slices.DeleteFunc(v.value, func(s string) bool { return s == o })
	default:
		v.SetValue(append(v.value, o)...)
	}
	if v.onChange != nil {
		v.onChange(v.Value())
	}
}

func (v *ToggleGroupView) Render(cx *el.Context) el.Element {
	id := autoID("togglegroup", v)
	gap := float32(theme.SpaceXs)
	if v.segmented {
		gap = 0
	}
	if v.gap != nil {
		gap = *v.gap
	}
	row := el.Div().Role("group").Row().Gap(gap)
	for i, o := range v.options {
		o := o
		itemID := id + "/" + strconv.Itoa(i)
		on := slices.Contains(v.value, o)
		text, icon, disabled, appearance := o, (*IconView)(nil), v.disabled, v.appearance
		if item, ok := v.items[o]; ok {
			text, icon, disabled = item.text, item.icon, disabled || item.disabled
			if item.appearance.variantSet {
				appearance.variant = item.appearance.variant
			}
			if item.appearance.sizeSet {
				appearance.size = item.appearance.size
			}
		}
		button := styledToggleButton(cx, appearance, itemID, text, icon, on, disabled, func() { v.toggle(o) })
		if text == "" {
			button.Name(o)
		}
		if v.segmented && gap == 0 {
			decorateToggleSegment(button, cx, itemID, i == 0, i == len(v.options)-1, on, appearance.variant)
		}
		row.Child(button)
	}
	return row
}
