package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// StatusBarView is a single-line 24dp status strip; its parent places it.
type StatusBarView struct {
	left, right []el.View
}

func StatusBar() *StatusBarView { return &StatusBarView{} }
func (v *StatusBarView) Left(views ...el.View) *StatusBarView {
	v.left = append([]el.View(nil), views...)
	return v
}
func (v *StatusBarView) Right(views ...el.View) *StatusBarView {
	v.right = append([]el.View(nil), views...)
	return v
}
func (v *StatusBarView) Render(cx *el.Context) el.Element {
	left := el.Div().Row().Grow().Gap(8).MaxLines(1).Items(el.Center)
	right := el.Div().Row().Gap(8).MaxLines(1).Items(el.Center)
	for _, e := range v.left {
		if e != nil {
			left.Child(e.Render(cx))
		}
	}
	for _, e := range v.right {
		if e != nil {
			right.Child(e.Render(cx))
		}
	}
	return el.Div().W(el.Full).H(el.Dp(24)).Role("status").TextSize(theme.TextSm).TextColor(theme.Muted).Bg(theme.Surface).Child(
		el.Div().Absolute().Top(0).Left(0).Right(0).H(el.Dp(1)).Bg(theme.Border),
		el.Div().W(el.Full).H(el.Full).Row().Px(8).Gap(8).Items(el.Center).Child(left, right),
	)
}
