package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// LinkView is clickable text in the primary color, e.g. "查看详情". Tab
// reaches it; Enter or Space follows it.
type LinkView struct {
	text     string
	onClick  func()
	disabled bool
}

func Link(text string, onClick func()) *LinkView { return &LinkView{text: text, onClick: onClick} }
func (v *LinkView) SetText(s string)             { v.text = s }
func (v *LinkView) SetDisabled(on bool)          { v.disabled = on }

func (v *LinkView) Render(cx *el.Context) el.Element {
	l := el.Div().Role("link").Name(v.text).Disabled(v.disabled).Rounded(2).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		TextColor(theme.PrimaryText).OnClick(func() {
		if !v.disabled && v.onClick != nil {
			v.onClick()
		}
	}).Child(el.Text(v.text))
	if v.disabled {
		l.TextColor(theme.Muted)
	} else {
		l.CursorPointer().Hover(func(s *el.Style) { s.TextColor(theme.PrimaryHover) })
	}
	return el.Div().Items(el.Start).Child(l)
}
