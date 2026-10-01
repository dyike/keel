package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// AlertView is a dismissible inline status message.
type AlertView struct {
	title, description string
	tone               Tone
	hidden             bool
	disabled           bool
	onClose            func()
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
	text := el.Div().Grow().Gap(6).Child(el.Text(v.title).Bold().TextColor(v.tone.color()))
	if v.description != "" {
		text.Child(el.Text(v.description).TextSize(13).TextColor(theme.Muted))
	}
	body := el.Div().Row().Grow().P(12).Gap(8).Items(el.Start).Child(Icon(name).Color(v.tone.color()).Render(cx), text)
	if v.onClose != nil {
		body.Child(el.Div().ID("close").Name("关闭 " + v.title).Focusable(true).P(4).OnClick(v.close).Child(Icon(IconClose).Render(cx)))
	}
	return el.Div().Disabled(v.disabled).W(el.Full).Role("alert").Name(v.title).Value(v.tone.name()).Row().Items(el.Stretch).Rounded(6).Border(1, theme.Border).Bg(theme.Surface).Child(el.Div().W(el.Dp(4)).NoShrink().Bg(v.tone.color()), body)
}
