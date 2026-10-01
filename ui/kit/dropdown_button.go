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
	label    string
	menu     *MenuView
	action   func()
	variant  ButtonVariant
	disabled bool
}

func DropdownButton(label string, menu *MenuView) *DropdownButtonView {
	return &DropdownButtonView{label: label, menu: menu}
}
func (v *DropdownButtonView) Split(action func()) *DropdownButtonView { v.action = action; return v }
func (v *DropdownButtonView) Variant(b ButtonVariant) *DropdownButtonView {
	v.variant = b
	return v
}
func (v *DropdownButtonView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.menu.SetValue(false)
	}
}

func (v *DropdownButtonView) Render(cx *el.Context) el.Element {
	if v.action == nil {
		v.menu.Trigger(dropdownPart(v.label, v.menu.Toggle, v).Icon(IconChevronDown))
		return v.menu.Render(cx)
	}
	v.menu.Trigger(dropdownPart("", v.menu.Toggle, v).Name(locale.Current().Name(v.label, locale.Current().MoreOptions)).Icon(IconChevronDown))
	return el.Div().Row().Gap(1).Items(el.Start).Child(dropdownPart(v.label, v.action, v).Render(cx), v.menu.Render(cx))
}

func dropdownPart(label string, fn func(), v *DropdownButtonView) *ButtonView {
	b := Button(label, fn).Variant(v.variant)
	b.SetDisabled(v.disabled)
	return b
}
