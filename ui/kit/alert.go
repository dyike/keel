package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

type AlertSize uint8

const (
	AlertSizeMedium AlertSize = iota
	AlertSizeXSmall
	AlertSizeSmall
	AlertSizeLarge
)

// AlertView is a dismissible inline status message.
type AlertView struct {
	title, description string
	tone               Tone
	hidden             bool
	disabled           bool
	onClose            func()
	banner             bool
	size               AlertSize
	icon               *IconName
	content            el.View
}

// Alert creates an inline message with an ToneInfo tone.
func Alert(title string) *AlertView                  { return &AlertView{title: title, tone: ToneInfo} }
func (v *AlertView) Description(s string) *AlertView { v.description = s; return v }
func (v *AlertView) Tone(t Tone) *AlertView          { v.SetTone(t); return v }
func (v *AlertView) SetTone(t Tone) {
	if t > ToneDanger {
		t = ToneInfo
	}
	v.tone = t
}
func (v *AlertView) SetTitle(s string)            { v.title = s }
func (v *AlertView) SetDescription(s string)      { v.description = s }
func (v *AlertView) OnClose(fn func()) *AlertView { v.onClose = fn; return v }
func (v *AlertView) SetDisabled(b bool)           { v.disabled = b }
func (v *AlertView) Visible() bool                { return !v.hidden }
func (v *AlertView) SetVisible(b bool)            { v.hidden = !b }

// Banner renders an edge-to-edge strip without a separate title row.
// If no body is supplied, the title becomes the banner message.
func (v *AlertView) Banner(on bool) *AlertView { v.banner = on; return v }

// Size controls text, icon and padding scales; Medium preserves the default.
func (v *AlertView) Size(size AlertSize) *AlertView {
	if size <= AlertSizeLarge {
		v.size = size
	}
	return v
}

// Icon overrides the tone's icon. IconNone hides the icon and its spacing.
func (v *AlertView) Icon(name IconName) *AlertView { v.icon = &name; return v }

// Content replaces the description, allowing rich text and application actions.
// Nil restores Description. The title remains the accessible name.
func (v *AlertView) Content(content el.View) *AlertView { v.content = content; return v }
func (v *AlertView) close() {
	if v.hidden || v.disabled {
		return
	}
	v.hidden = true
	if v.onClose != nil {
		v.onClose()
	}
}
func (v *AlertView) Render(cx *el.Context) el.Element {
	if v.hidden {
		return el.Div().Hidden(true)
	}
	name := IconInfo
	switch v.tone {
	case ToneSuccess:
		name = IconCheck
	case ToneWarning:
		name = IconWarning
	case ToneDanger:
		name = IconError
	}

	if v.icon != nil {
		name = *v.icon
	}
	font, descriptionFont, padding, gap, iconSize := float32(theme.TextBody), float32(theme.TextMd), float32(theme.SpaceLg), float32(theme.SpaceMd), float32(18)
	switch v.size {
	case AlertSizeXSmall:
		font, descriptionFont, padding, gap, iconSize = theme.TextSm, theme.TextXs, theme.SpaceSm, theme.SpaceXs, 14
	case AlertSizeSmall:
		font, descriptionFont, padding, gap, iconSize = theme.TextControl, theme.TextSm, theme.SpaceMd, theme.SpaceSm, 16
	case AlertSizeLarge:
		font, descriptionFont, padding, gap, iconSize = theme.TextLg, theme.TextBody, theme.SpaceXl, theme.SpaceLg, 22
	}
	text := el.Div().Grow().MinW(el.Dp(0)).Gap(theme.SpaceSm).TextSize(descriptionFont).TextColor(theme.Text)
	if !v.banner && v.title != "" {
		text.Child(el.Text(v.title).Bold().TextSize(font).TextColor(v.tone.color()))
	}
	if v.content != nil {
		text.Child(v.content.Render(cx))
	} else if v.description != "" {
		text.Child(el.Text(v.description).TextColor(theme.Muted))
	} else if v.banner {
		text.Child(el.Text(v.title).TextSize(font).TextColor(v.tone.color()))
	}
	body := el.Div().Row().Grow().MinW(el.Dp(0)).P(padding).Gap(gap).Items(el.Start)
	if name != IconNone {
		body.Child(Icon(name).Size(iconSize).Color(v.tone.color()).Render(cx))
	}
	body.Child(text)
	if v.onClose != nil {
		body.Child(el.Div().ID("close").Name(locale.Current().Name(locale.Current().Close, v.title)).Focusable(true).NoShrink().P(theme.SpaceXs).OnClick(v.close).Child(Icon(IconClose).Size(iconSize).Render(cx)))
	}
	box := el.Div().Disabled(v.disabled).W(el.Full).Role("alert").Name(v.title).Value(v.tone.name()).Row().Items(el.Stretch)
	if v.banner {
		box.Bg(tint(v.tone.color(), 24))
	} else {
		box.Rounded(theme.RadiusMd).Border(1, theme.Border).Bg(theme.Surface).Child(el.Div().W(el.Dp(4)).NoShrink().Bg(v.tone.color()))
	}
	return box.Child(body)
}
