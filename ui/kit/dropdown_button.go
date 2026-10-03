package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

// DropdownButtonView is a button that opens a menu. With Split, the label
// runs a main action and a separate arrow opens the menu:
//
//	kit.DropdownButton("导出", formats)            // the whole button opens the menu
//	kit.DropdownButton("保存", more).Split(save)   // 保存 runs save; ▾ opens more
type DropdownButtonView struct {
	label      string
	menu       *MenuView
	action     func()
	variant    ButtonVariant
	disabled   bool
	button     *ButtonView
	variantSet bool
	size       float32
	loading    *bool
}

func DropdownButton(label string, menu *MenuView) *DropdownButtonView {
	if menu == nil {
		menu = Menu()
	}
	return &DropdownButtonView{label: label, menu: menu}
}
func (v *DropdownButtonView) Split(action func()) *DropdownButtonView { v.action = action; return v }
func (v *DropdownButtonView) Variant(b ButtonVariant) *DropdownButtonView {
	v.variant, v.variantSet = b, true
	return v
}

// Button supplies the main half of a split button. Render copies its configuration
// so the source can be reused. Nil restores the ordinary/Split configuration.
func (v *DropdownButtonView) Button(button *ButtonView) *DropdownButtonView {
	v.button = button
	return v
}

// Size sets both halves' height in dp; zero inherits the inner button/default.
func (v *DropdownButtonView) Size(dp float32) *DropdownButtonView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.size = dp
	}
	return v
}

// Loading overrides loading on the main action only. A split arrow remains usable.
func (v *DropdownButtonView) Loading(on bool) *DropdownButtonView { v.loading = &on; return v }

func (v *DropdownButtonView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.menu.SetValue(false)
	}
}

func (v *DropdownButtonView) Render(cx *el.Context) el.Element {
	id := autoID("dropdown-button", v)
	split := v.action != nil || v.button != nil
	main := Button(v.label, v.menu.Toggle)
	if v.button != nil {
		copy := *v.button
		main = &copy
	}
	if v.action != nil {
		main.onClick = v.action
	}
	if v.variantSet {
		main.Variant(v.variant)
	}
	if v.size > 0 {
		main.Size(v.size)
	}
	if v.loading != nil {
		main.Loading(*v.loading)
	}
	main.ID(id + "/main")
	main.SetDisabled(v.disabled || main.disabled)
	if !split {
		main.Icon(IconChevronDown)
		v.menu.Trigger(main)
		return el.Div().Disabled(v.disabled).Items(el.Start).Child(v.menu.Render(cx))
	}
	name := main.name
	if name == "" {
		name = main.text
	}
	arrow := Button("", v.menu.Toggle).ID(id + "/arrow").Variant(main.variant).Size(main.height).
		Name(locale.Current().Name(name, locale.Current().MoreOptions)).Icon(IconChevronDown)
	arrow.SetDisabled(v.disabled)
	v.menu.Trigger(arrow)
	return el.Div().Disabled(v.disabled).Row().Gap(1).Items(el.Start).Child(main.Render(cx), v.menu.Render(cx))
}
