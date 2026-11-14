package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

// Popup changes the inline panel to a trigger. Narrow viewports use a bottom
// sheet; wider ones use an anchored popover. Choosing closes either surface.
func (v *EmojiPickerView) Popup(on bool) *EmojiPickerView {
	v.popup = on
	if !on {
		v.SetOpen(false)
	}
	return v
}
func (v *EmojiPickerView) Open() bool { return v.open }
func (v *EmojiPickerView) SetOpen(on bool) {
	v.open = on && !v.disabled
	if v.popover != nil {
		v.popover.SetValue(v.open)
	}
	if v.sheet != nil {
		v.sheet.SetValue(v.open)
	}
}
func (v *EmojiPickerView) renderPopup(cx *el.Context) el.Element {
	text := locale.Current()
	name := v.name
	if name == "" {
		name = text.EmojiPicker
	}
	if v.popover == nil {
		body := el.ViewFunc(v.panel)
		v.popover = Popover(body).Width(376).OnChange(func(open bool) { v.open = open })
		v.sheet = Sheet(el.Bottom, name).Body(body).OnClose(func() { v.open = false })
	}
	width, height := cx.ViewportSize()
	narrow := width < 600*max(float32(1), cx.TextScale())
	v.sheetPanel = narrow
	trigger := Button("☺", func() { v.SetOpen(!v.open) }).Name(name).Size(44).Variant(ButtonSecondary)
	trigger.SetDisabled(v.disabled)
	if narrow {
		v.popover.SetValue(false)
		v.sheet.SetTitle(name)
		v.sheet.Size(min(v.height+240*max(float32(1), cx.TextScale()), height*.85))
		v.sheet.SetValue(v.open)
		return el.Div().Disabled(v.disabled).Child(trigger.Render(cx), v.sheet.Render(cx))
	}
	v.sheet.SetValue(false)
	v.popover.SetValue(v.open)
	v.popover.Trigger(trigger)
	return el.Div().Disabled(v.disabled).Child(v.popover.Render(cx))
}
