package kit

import "github.com/dyike/keel/ui/el"
import "github.com/dyike/keel/ui/theme"

// AlertView displays an inline status and its explanation. It has no action
// or dismiss button; those require the reviewed el focus and disabled APIs.
type AlertView struct {
	title, description string
	tone               Tone
}

func Alert(title, description string) *AlertView {
	return &AlertView{title: title, description: description, tone: Info}
}
func (a *AlertView) Tone(t Tone) *AlertView  { a.tone = t; return a }
func (a *AlertView) SetTitle(s string)       { a.title = s }
func (a *AlertView) SetDescription(s string) { a.description = s }
func (a *AlertView) Render(*el.Context) el.Element {
	box := el.Div().Role("alert").Name(a.title).Value(a.tone.name()).P(12).Gap(6).Rounded(6).Border(1, a.tone.color()).Bg(theme.Surface)
	if a.title != "" {
		box.Child(el.Text(a.title).TextColor(a.tone.color()))
	}
	if a.description != "" {
		box.Child(el.Text(a.description).TextColor(theme.Text))
	}
	return box
}
