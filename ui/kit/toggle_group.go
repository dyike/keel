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
}

func ToggleGroup(options ...string) *ToggleGroupView  { return &ToggleGroupView{options: options} }
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
	row := el.Div().Role("group").Row().Gap(theme.SpaceXs)
	for i, o := range v.options {
		o := o
		row.Child(toggleButton(cx, id+"/"+strconv.Itoa(i), o, nil, slices.Contains(v.value, o), v.disabled, func() { v.toggle(o) }))
	}
	return row
}
