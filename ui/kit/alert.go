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
	onClose            func()
}

// Alert accepts an optional description for compatibility with the first kit draft.
func Alert(title string, description ...string) *AlertView {
	v := &AlertView{title: title, tone: Info}
	if len(description) > 0 {
		v.description = description[0]
	}
	return v
}
func (v *AlertView) Description(s string) *AlertView { v.description = s; return v }
func (v *AlertView) Tone(t Tone) *AlertView          { v.SetKind(t); return v }
func (v *AlertView) Success() *AlertView             { return v.Tone(Success) }
func (v *AlertView) Warning() *AlertView             { return v.Tone(Warning) }
func (v *AlertView) Danger() *AlertView              { return v.Tone(Danger) }
func (v *AlertView) SetKind(t Tone) {
	if t < Info || t > Danger {
		t = Info
	}
	v.tone = t
}
func (v *AlertView) SetTitle(s string)            { v.title = s }
func (v *AlertView) SetDescription(s string)      { v.description = s }
func (v *AlertView) OnClose(fn func()) *AlertView { v.onClose = fn; return v }
func (v *AlertView) Visible() bool                { return !v.hidden }
func (v *AlertView) SetVisible(b bool)            { v.hidden = !b }
func (v *AlertView) close() {
	if v.hidden {
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
	case Success:
		name = IconCheck
	case Warning:
		name = IconWarning
	case Danger:
		name = IconError
	}
	text := el.Div().Grow().Gap(6).Child(el.Text(v.title).Bold().TextColor(v.tone.color()))
	if v.description != "" {
		text.Child(el.Text(v.description).TextSize(13).TextColor(theme.Muted))
	}
	body := el.Div().Row().Grow().P(12).Gap(8).Items(el.Start).Child(Icon(name).Color(v.tone.color()).Render(cx), text)
	if v.onClose != nil {
		body.Child(el.Div().ID("close").Name("关闭 " + v.title).Focusable().P(4).OnClick(v.close).Child(Icon(IconClose).Render(cx)))
	}
	return el.Div().W(el.Full).Role("alert").Name(v.title).Value(v.tone.name()).Row().Items(el.Stretch).Rounded(6).Border(1, theme.Border).Bg(theme.Surface).Child(el.Div().W(el.Dp(4)).NoShrink().Bg(v.tone.color()), body)
}
