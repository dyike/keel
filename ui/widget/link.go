package widget

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// LinkText is clickable text in the primary color.
type LinkText struct {
	text    string
	onClick func()
	click   widget.Clickable
}

// Link creates clickable text.
func Link(text string, onClick func()) *LinkText { return &LinkText{text: text, onClick: onClick} }

func (l *LinkText) SetText(s string) { l.text = s }

func (l *LinkText) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{} // don't stretch inside Column
	for l.click.Clicked(gtx) {
		core.Call(gtx, l.onClick)
	}
	return l.click.Layout(gtx, func(gtx C) D {
		pointer.CursorPointer.Add(gtx.Ops)
		semantic.Button.Add(gtx.Ops)
		core.Role("link").Add(gtx.Ops)
		lb := material.Label(theme.Material, theme.BodySize, l.text)
		lb.Color = theme.Primary
		return lb.Layout(gtx)
	})
}
