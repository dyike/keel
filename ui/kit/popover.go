package kit

import "github.com/dyike/keel/ui/el"

// PopoverView shows content next to a trigger. The trigger opens it itself,
// so it keeps its own Tab stop and keyboard handling:
//
//	filters := kit.Popover(form)
//	filters.Trigger(kit.Button("筛选", filters.Toggle).Variant(kit.ButtonSecondary))
//
// Clicking outside or pressing Esc closes it. It does not move focus.
type PopoverView struct {
	trigger, content el.View
	open             bool
	side             el.Side
	align            el.Align
	width            float32
	onChange         func(bool)
}

func Popover(content el.View) *PopoverView { return &PopoverView{content: content} }

// Trigger sets the view the popover is anchored to.
func (v *PopoverView) Trigger(t el.View) *PopoverView { v.trigger = t; return v }
func (v *PopoverView) Placement(side el.Side, align el.Align) *PopoverView {
	v.side, v.align = side, align
	return v
}

// Width sets the panel width in dp; by default it fits its content.
func (v *PopoverView) Width(dp float32) *PopoverView { v.width = dp; return v }

// OnChange is called when the user opens or closes the popover, not by SetOpen.
func (v *PopoverView) OnChange(fn func(bool)) *PopoverView { v.onChange = fn; return v }
func (v *PopoverView) Value() bool                         { return v.open }
func (v *PopoverView) SetValue(open bool)                  { v.open = open }

// Toggle opens or closes the popover as a user action; pass it as the
// trigger's click handler.
func (v *PopoverView) Toggle() { v.change(!v.open) }

func (v *PopoverView) change(open bool) {
	if v.open == open {
		return
	}
	v.open = open
	if v.onChange != nil {
		v.onChange(open)
	}
}

func (v *PopoverView) Render(cx *el.Context) el.Element {
	id := autoID("popover", v)
	if v.open {
		panel := surface().Role("dialog").P(12)
		if v.width > 0 {
			panel.W(el.Dp(v.width))
		}
		if v.content != nil {
			panel.Child(v.content.Render(cx))
		}
		cx.Overlay(id, el.Anchored(id, panel).Placement(v.side, v.align).OnDismiss(func() { v.change(false) }))
	}
	return anchor(id, cx, v.trigger)
}
