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
)

// ButtonView is an action with pointer and keyboard activation.
type ButtonView struct {
	text              string
	onClick           func()
	variant           ButtonVariant
	height            float32
	icon              *IconView
	disabled, loading bool
}

func Button(text string, onClick func()) *ButtonView {
	return &ButtonView{text: text, onClick: onClick, height: 32}
}
func (v *ButtonView) Variant(variant ButtonVariant) *ButtonView {
	if variant <= ButtonDanger {
		v.variant = variant
	}
	return v
}

// Size sets the height in dp. Recommended heights are 28, 32 and 40.
func (v *ButtonView) Size(dp float32) *ButtonView {
	if dp > 0 && !math.IsInf(float64(dp), 0) {
		v.height = dp
	}
	return v
}
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
	bg, hover, fg := theme.Primary, theme.PrimaryHover, theme.OnColor
	switch v.variant {
	case ButtonSecondary:
		bg, hover, fg = theme.Subtle, theme.SubtleHover, theme.Text
	case ButtonGhost:
		bg, hover, fg = color.NRGBA{}, theme.Subtle, theme.Text
	case ButtonDanger:
		bg, hover, fg = theme.Danger, theme.DangerHover, theme.OnColor
	}
	// Active is a stronger shade of the hover background; resolve every Render.
	active := hover
	active.R = uint8(uint16(active.R) * 9 / 10)
	active.G = uint8(uint16(active.G) * 9 / 10)
	active.B = uint8(uint16(active.B) * 9 / 10)
	if v.disabled {
		fg = theme.Muted
		if v.variant != ButtonGhost {
			bg = theme.Subtle
		}
	}
	font, iconSize, padding := float32(14), float32(16), float32(16)
	if v.height <= 28 {
		font, iconSize, padding = 12, 14, 10
	} else if v.height >= 40 {
		font, iconSize, padding = 16, 20, 22
	}
	box := el.Div().Role("button").Name(v.text).H(el.Dp(v.height)).MaxW(el.Full).Px(padding).Row().Gap(6).Items(el.Center).Justify(el.Center).
		Rounded(6).Bg(bg).TextColor(fg).TextSize(font).Focusable(true).OnClick(v.activate).
		Disabled(v.disabled).
		DisabledStyle(func(s *el.Style) { s.Bg(bg).TextColor(fg) }).
		FocusStyle(func(s *el.Style) {
			if v.variant == ButtonPrimary || v.variant == ButtonDanger {
				s.BorderColor(theme.OnColor)
			} else {
				s.BorderColor(theme.PrimaryText)
			}
		})
	if !v.loading && !v.disabled {
		box.CursorPointer().Hover(func(s *el.Style) { s.Bg(hover) }).Active(func(s *el.Style) { s.Bg(active) })
	}
	if v.loading {
		box.Value("loading")
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
		text := el.Text(v.text).MaxLines(1).DisabledStyle(func(s *el.Style) { s.TextColor(fg) })
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
