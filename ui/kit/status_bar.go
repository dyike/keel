package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// StatusBarView is a single-line 24dp status strip; its parent places it.
type StatusBarView struct {
	status, detail string
	left, right    []el.Element
}

func StatusBar(text ...string) *StatusBarView {
	v := &StatusBarView{}
	if len(text) > 0 {
		v.status = text[0]
	}
	if len(text) > 1 {
		v.detail = text[1]
	}
	return v
}
func (v *StatusBarView) Left(els ...el.Element) *StatusBarView {
	v.left = append([]el.Element(nil), els...)
	return v
}
func (v *StatusBarView) Right(els ...el.Element) *StatusBarView {
	v.right = append([]el.Element(nil), els...)
	return v
}
func (v *StatusBarView) SetStatus(s string) { v.status = s }
func (v *StatusBarView) SetDetail(s string) { v.detail = s }
func (v *StatusBarView) Render(*el.Context) el.Element {
	left := el.Div().Row().Grow().Gap(8).MaxLines(1).Items(el.Center)
	right := el.Div().Row().Gap(8).MaxLines(1).Items(el.Center)
	if v.status != "" {
		left.Child(el.Text(v.status))
	}
	if v.detail != "" {
		right.Child(el.Text(v.detail))
	}
	for _, e := range v.left {
		if e != nil {
			left.Child(e)
		}
	}
	for _, e := range v.right {
		if e != nil {
			right.Child(e)
		}
	}
	return el.Div().W(el.Full).H(el.Dp(24)).Role("status").Name(v.status).Value(v.detail).TextSize(12).TextColor(theme.Muted).Bg(theme.Surface).Child(
		el.Div().Absolute().Top(0).Left(0).Right(0).H(el.Dp(1)).Bg(theme.Border),
		el.Div().W(el.Full).H(el.Full).Row().Px(8).Gap(8).Items(el.Center).Child(left, right),
	)
}
