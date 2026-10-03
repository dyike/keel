package kit

import (
	"image/color"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ToggleView is a button that stays pressed, e.g. bold in a toolbar.
type ToggleView struct {
	text            string
	icon            *IconView
	value, disabled bool
	onChange        func(bool)
	appearance      toggleAppearance
}

func Toggle(text string, on bool) *ToggleView            { return &ToggleView{text: text, value: on} }
func (v *ToggleView) Icon(name IconName) *ToggleView     { v.icon = Icon(name); return v }
func (v *ToggleView) OnChange(fn func(bool)) *ToggleView { v.onChange = fn; return v }
func (v *ToggleView) Value() bool                        { return v.value }
func (v *ToggleView) SetValue(on bool)                   { v.value = on }
func (v *ToggleView) SetDisabled(on bool)                { v.disabled = on }

func (v *ToggleView) Render(cx *el.Context) el.Element {
	return styledToggleButton(cx, v.appearance, autoID("toggle", v), v.text, v.icon, v.value, v.disabled, func() {
		v.value = !v.value
		if v.onChange != nil {
			v.onChange(v.value)
		}
	})
}

func toggleButton(cx *el.Context, id, text string, icon *IconView, on, disabled bool, fn func()) *el.DivEl {
	return styledToggleButton(cx, toggleAppearance{}, id, text, icon, on, disabled, fn)
}

func styledToggleButton(cx *el.Context, appearance toggleAppearance, id, text string, icon *IconView, on, disabled bool, fn func()) *el.DivEl {
	height, font, iconSize, padding := appearance.metrics()
	bg, fg, border := theme.Surface, theme.Text, theme.Border
	if appearance.variant == ToggleGhost {
		bg, border = color.NRGBA{}, color.NRGBA{}
	}
	if appearance.variant == ToggleOutline {
		bg = color.NRGBA{}
	}
	if on {
		bg, fg = theme.Highlight, theme.PrimaryText
		if appearance.variant != ToggleGhost {
			border = theme.Primary
		}
	}
	if disabled {
		fg = theme.Muted
	}
	b := el.Div().ID(id).Role("toggle").Name(text).Selected(on).Disabled(disabled).
		Row().Items(el.Center).Gap(theme.SpaceSm).H(el.Dp(height)).Px(padding).Rounded(theme.RadiusMd).Bg(bg).Border(1, border).TextColor(fg).TextSize(font).
		Focusable(true).OnClick(fn).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) })
	if !disabled {
		b.CursorPointer().Hover(func(s *el.Style) {
			if !on {
				s.Bg(theme.SubtleHover)
			}
		})
	}
	if icon != nil {
		ic := *icon
		b.Child(ic.Size(iconSize).Color(fg).Render(cx))
	}
	if text != "" {
		b.Child(el.Text(text))
	}
	return b
}
