package kit

import (
	"image/color"
	"math"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// ButtonVariant selects the button's visual treatment.
type ButtonVariant uint8

const (
	ButtonPrimary ButtonVariant = iota
	ButtonSecondary
	ButtonGhost
	ButtonDanger
	ButtonLink
	ButtonText
	ButtonSuccess
	ButtonWarning
	ButtonInfo
)

// ButtonAppearance contains colors for a button's normal and interactive states.
// Transparent colors are valid; disabled colors are supplied by the theme.
type ButtonAppearance struct {
	Background, Foreground            color.NRGBA
	Hover, Active                     color.NRGBA
	HoverForeground, ActiveForeground color.NRGBA
	Border, Focus                     color.NRGBA
}

// ButtonView is an action with pointer and keyboard activation.
type ButtonView struct {
	id, name          string
	text              string
	onClick           func()
	variant           ButtonVariant
	height            float32
	icon              *IconView
	disabled, loading bool
	outline, compact  bool
	content           el.View
	appearance        func(ButtonAppearance) ButtonAppearance
}

func Button(text string, onClick func()) *ButtonView {
	return &ButtonView{text: text, onClick: onClick, height: 32}
}
func (v *ButtonView) Variant(variant ButtonVariant) *ButtonView {
	if variant <= ButtonInfo {
		v.variant = variant
	}
	return v
}

// Outline keeps a colored border and text with a transparent background.
func (v *ButtonView) Outline(on bool) *ButtonView { v.outline = on; return v }

// Compact reduces horizontal padding, preserving the configured height.
func (v *ButtonView) Compact(on bool) *ButtonView { v.compact = on; return v }

// Content replaces the icon and visible label. Supply display-only content;
// the constructor label (or Name) remains the accessible name. Nil restores it.
func (v *ButtonView) Content(content el.View) *ButtonView { v.content = content; return v }

// Appearance transforms variant colors on each Render, before disabled styling.
// Nil restores theme colors. It can reference the current theme at render time.
func (v *ButtonView) Appearance(fn func(ButtonAppearance) ButtonAppearance) *ButtonView {
	v.appearance = fn
	return v
}

// Size sets the height in dp. Recommended heights are 28, 32 and 40.
func (v *ButtonView) Size(dp float32) *ButtonView {
	if dp > 0 && !math.IsInf(float64(dp), 0) {
		v.height = dp
	}
	return v
}

// ID names the button for el: cx.Focus, cx.Hovered and anchored layers.
func (v *ButtonView) ID(id string) *ButtonView { v.id = id; return v }

// Name sets the accessible name; icon-only buttons need one.
func (v *ButtonView) Name(s string) *ButtonView      { v.name = s; return v }
func (v *ButtonView) Icon(name IconName) *ButtonView { v.icon = Icon(name); return v }
func (v *ButtonView) Loading(on bool) *ButtonView    { v.SetLoading(on); return v }
func (v *ButtonView) SetText(s string)               { v.text = s }
func (v *ButtonView) SetDisabled(on bool)            { v.disabled = on }
func (v *ButtonView) SetLoading(on bool)             { v.loading = on }
func (v *ButtonView) activate() {
	if !v.disabled && !v.loading && v.onClick != nil {
		v.onClick()
	}
}

func (v *ButtonView) Render(cx *el.Context) el.Element {
	return v.renderWithRadius(cx, theme.RadiusMd)
}

// renderWithRadius lets composite controls refine corners without mutating the button.
func (v *ButtonView) renderWithRadius(cx *el.Context, radius float32) el.Element {
	bg, hover, fg := theme.Primary, theme.PrimaryHover, theme.OnColor
	switch v.variant {
	case ButtonSecondary:
		bg, hover, fg = theme.Subtle, theme.SubtleHover, theme.Text
	case ButtonGhost:
		bg, hover, fg = color.NRGBA{}, theme.Subtle, theme.Text
	case ButtonDanger:
		bg, hover, fg = theme.Danger, theme.DangerHover, theme.OnColor
	case ButtonLink:
		bg, hover, fg = color.NRGBA{}, color.NRGBA{}, theme.PrimaryText
	case ButtonText:
		bg, hover, fg = color.NRGBA{}, color.NRGBA{}, theme.Text
	case ButtonSuccess:
		bg, hover, fg = theme.Success, theme.Success, theme.OnColor
	case ButtonWarning:
		bg, hover, fg = theme.Warning, theme.Warning, theme.OnColor
	case ButtonInfo:
		bg, hover, fg = theme.Info, theme.Info, theme.OnColor
	}
	if v.variant >= ButtonSuccess {
		fg = contrastingText(bg)
	}
	// Active is a stronger shade of the hover background; resolve every Render.
	active := hover
	active.R = uint8(uint16(active.R) * 9 / 10)
	active.G = uint8(uint16(active.G) * 9 / 10)
	active.B = uint8(uint16(active.B) * 9 / 10)
	focus := theme.PrimaryText
	if v.variant == ButtonPrimary || v.variant == ButtonDanger || v.variant >= ButtonSuccess {
		focus = fg
	}
	appearance := ButtonAppearance{Background: bg, Foreground: fg, Hover: hover, Active: active, HoverForeground: fg, ActiveForeground: fg, Focus: focus}
	if v.variant == ButtonLink || v.variant == ButtonText {
		appearance.HoverForeground, appearance.ActiveForeground = theme.PrimaryHover, theme.PrimaryText
	}
	if v.outline {
		border, text := bg, bg
		switch v.variant {
		case ButtonPrimary:
			text = theme.PrimaryText
		case ButtonDanger:
			text = theme.DangerText
		case ButtonSecondary, ButtonGhost, ButtonText:
			border, text = theme.Border, theme.Text
		case ButtonLink:
			border, text = theme.PrimaryText, theme.PrimaryText
		}
		appearance = ButtonAppearance{Foreground: text, Hover: theme.Subtle, Active: theme.SubtleHover, HoverForeground: text, ActiveForeground: text, Border: border, Focus: theme.PrimaryText}
	}
	if v.appearance != nil {
		appearance = v.appearance(appearance)
	}
	bg, fg = appearance.Background, appearance.Foreground
	disabledBg := theme.Subtle
	if v.outline || v.variant == ButtonGhost || v.variant == ButtonLink || v.variant == ButtonText {
		disabledBg = color.NRGBA{}
	}
	if v.disabled {
		bg, fg = disabledBg, theme.Muted
	}
	font, iconSize, padding := float32(theme.TextControl), float32(16), float32(16)
	if v.height <= 28 {
		font, iconSize, padding = theme.TextSm, 14, 10
	} else if v.height >= 40 {
		font, iconSize, padding = theme.TextBody, 20, 22
	}
	if v.compact {
		padding = theme.SpaceSm
	}
	if v.variant == ButtonLink {
		padding = 0
	}
	name := v.name
	if name == "" {
		name = v.text
	}
	box := el.Div().ID(v.id).Role("button").Name(name).H(el.Dp(v.height)).MaxW(el.Full).Px(padding).Row().Gap(theme.SpaceSm).Items(el.Center).Justify(el.Center).
		Rounded(radius).Bg(bg).TextColor(fg).TextSize(font).Focusable(true).OnClick(v.activate).
		Disabled(v.disabled).
		DisabledStyle(func(s *el.Style) { s.Bg(disabledBg).TextColor(theme.Muted).BorderColor(theme.Border) }).
		FocusStyle(func(s *el.Style) { s.BorderColor(appearance.Focus) })
	if v.outline || appearance.Border.A > 0 {
		box.Border(1, appearance.Border)
	}
	if v.content == nil && v.text == "" && v.icon != nil {
		box.W(el.Dp(v.height)).Px(0)
	}

	if g := theme.PrimaryGradient; v.variant == ButtonPrimary && !v.outline && v.appearance == nil && !v.disabled && !g.IsZero() {
		box.BgGradient(g)
		if !v.loading {
			box.CursorPointer().Hover(func(s *el.Style) { s.BgGradient(shadeGradient(g, 9)) }).
				Active(func(s *el.Style) { s.BgGradient(shadeGradient(g, 8)) })
		}
	} else if !v.loading && !v.disabled {
		box.CursorPointer().Hover(func(s *el.Style) { s.Bg(appearance.Hover).TextColor(appearance.HoverForeground) }).Active(func(s *el.Style) { s.Bg(appearance.Active).TextColor(appearance.ActiveForeground) })
	}
	if v.loading {
		box.Value("loading")
	}
	if v.content != nil {
		content := el.Div().MaxW(el.Full).Child(v.content.Render(cx))
		if v.loading {
			content.Opacity(0)
		}
		box.Child(content)
		if v.loading {
			box.Child(el.Div().Absolute().Top(0).Left(0).Right(0).Bottom(0).Center().Child(spinnerRing(cx, iconSize, fg)))
		}
		return box
	}

	if v.icon != nil {
		if v.loading {
			box.Child(spinnerRing(cx, iconSize, fg))
		} else {
			icon := *v.icon
			box.Child(icon.Size(iconSize).Color(fg).Render(cx))
		}
	}
	if v.text != "" {
		text := el.Text(v.text).MaxLines(1).DisabledStyle(func(s *el.Style) { s.TextColor(theme.Muted) })
		if v.loading && v.icon == nil {
			text.TextColor(color.NRGBA{}).DisabledStyle(func(s *el.Style) { s.TextColor(color.NRGBA{}) })
		}
		box.Child(text)
	}
	if v.loading && v.icon == nil {
		// Keep the label's measured width while displaying the spinner above it.
		box.Child(el.Div().Absolute().Top(0).Left(0).Right(0).Bottom(0).Center().Child(spinnerRing(cx, iconSize, fg)))
	}
	return box
}

// shadeGradient darkens both ends of g to tenths of their brightness, for a
// gradient button's hover and pressed states.
func shadeGradient(g theme.Gradient, tenths uint16) theme.Gradient {
	shade := func(c color.NRGBA) color.NRGBA {
		c.R, c.G, c.B = uint8(uint16(c.R)*tenths/10), uint8(uint16(c.G)*tenths/10), uint8(uint16(c.B)*tenths/10)
		return c
	}
	g.From, g.To = shade(g.From), shade(g.To)
	return g
}
