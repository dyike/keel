package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// StatusBarView lays out a primary status and trailing detail. The parent
// decides where the bar is placed; this view does not attach to a window.
type StatusBarView struct{ status, detail string }

func StatusBar(status, detail string) *StatusBarView { return &StatusBarView{status, detail} }
func (v *StatusBarView) SetStatus(s string)          { v.status = s }
func (v *StatusBarView) SetDetail(s string)          { v.detail = s }
func (v *StatusBarView) Render(*el.Context) el.Element {
	return el.Div().W(el.Full).Role("status").Name(v.status).Value(v.detail).Row().Gap(12).Px(12).Py(6).Bg(theme.Subtle).Child(el.Text(v.status).Grow().TextSize(float32(theme.SmallSize)).TextColor(theme.Text), el.Text(v.detail).Grow().TextSize(float32(theme.SmallSize)).TextColor(theme.Muted))
}
