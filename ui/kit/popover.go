package kit

import (
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

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
	disabled         bool
	side             el.Side
	align            el.Align
	width            float32
	offset           float32
	onChange         func(bool)
	plain            bool
	arrow            bool
	mouseButton      pointer.Buttons
	panelStyle       func(*el.DivEl)
}

func Popover(content el.View) *PopoverView { return &PopoverView{content: content, offset: 4} }

// Trigger sets the view the popover is anchored to.
func (v *PopoverView) Trigger(t el.View) *PopoverView { v.trigger = t; return v }
func (v *PopoverView) Placement(side el.Side, align el.Align) *PopoverView {
	v.side, v.align = side, align
	return v
}

// RightClick toggles the popover on secondary presses over its trigger.
// Primary and keyboard actions remain owned by the trigger. The default is false.
// Do not also wire Toggle to the trigger's context-menu handler.
func (v *PopoverView) RightClick(on bool) *PopoverView {
	if on {
		return v.MouseButton(pointer.ButtonSecondary)
	}
	return v.MouseButton(0)
}

// MouseButton selects an automatic mouse-press trigger. Zero (the default)
// leaves activation to the trigger view; invalid values and chords are ignored.
// The trigger's own handlers still run: do not also bind Toggle to that button.
func (v *PopoverView) MouseButton(button pointer.Buttons) *PopoverView {
	switch button {
	case 0, pointer.ButtonPrimary, pointer.ButtonSecondary, pointer.ButtonTertiary:
		v.mouseButton = button
	}
	return v
}

// Offset sets the gap from the trigger in dp, 4 by default. Zero makes the
// panel touch the trigger; negative values overlap it. Non-finite values
// are ignored. Viewport avoidance remains active.
func (v *PopoverView) Offset(dp float32) *PopoverView {
	if finiteNumber(float64(dp)) {
		v.offset = dp
	}
	return v
}

// Arrow shows a pointer toward the trigger, following placement and flips.
// Offset measures to its tip; the panel is placed another 6dp away.
func (v *PopoverView) Arrow(on bool) *PopoverView { v.arrow = on; return v }

// Appearance controls the default background, border, radius, shadow and padding.
// It is enabled by default; disabling it leaves positioning and behavior intact.
func (v *PopoverView) Appearance(on bool) *PopoverView { v.plain = !on; return v }

// PanelStyle refines the panel after appearance defaults and Width, each frame.
// Nil removes the refinement. Do not retain the element. Identity, dialog role,
// viewport limits and scrolling are maintained by the popover.
func (v *PopoverView) PanelStyle(fn func(*el.DivEl)) *PopoverView {
	v.panelStyle = fn
	return v
}

// Width sets the panel width in dp; by default it fits its content.
func (v *PopoverView) Width(dp float32) *PopoverView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.width = dp
	}
	return v
}

// OnChange is called when the user opens or closes the popover, not by SetOpen.
func (v *PopoverView) OnChange(fn func(bool)) *PopoverView { v.onChange = fn; return v }
func (v *PopoverView) Value() bool                         { return v.open }
func (v *PopoverView) SetValue(open bool)                  { v.open = open && !v.disabled }

// Toggle opens or closes the popover as a user action; pass it as the
// trigger's click handler.
func (v *PopoverView) Toggle() { v.change(!v.open) }

func (v *PopoverView) change(open bool) {
	if v.disabled || v.open == open {
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
		w, h := cx.ViewportSize()
		panel := el.Div()
		if !v.plain {
			panel = floating(theme.ElevationMd).P(theme.SpaceLg)
		}
		if v.width > 0 {
			panel.W(el.Dp(v.width))
		}
		if v.panelStyle != nil {
			v.panelStyle(panel)
		}
		panel.ID(id + "/panel").Role("dialog").MaxW(el.Dp(max(0, w-16))).MaxH(el.Dp(max(0, h-16))).ScrollY().ScrollX()
		cx.Overlay(id, el.Anchored(id, panel).Placement(v.side, v.align).Offset(v.offset).Arrow(v.arrow).OnDismiss(func() { v.change(false) }))
		if v.content != nil {
			panel.Child(v.content.Render(cx))
		}
	}
	target := anchor(id, cx, v.trigger).Focusable(false)
	if v.mouseButton != 0 {
		target.OnMousePress(v.mouseButton, v.Toggle)
	}
	return el.Div().Disabled(v.disabled).Child(target)
}

func (v *PopoverView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.open = false
	}
}
