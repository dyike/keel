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
	text     string
	onClick  func()
	click    widget.Clickable
	disabled bool
}

// Link creates clickable text.
func Link(text string, onClick func()) *LinkText { return &LinkText{text: text, onClick: onClick} }

func (l *LinkText) SetText(s string) { l.text = s }

// SetDisabled prevents pointer and keyboard activation.
func (l *LinkText) SetDisabled(v bool) { l.disabled = v }

func (l *LinkText) Layout(gtx C) D {
	if l.disabled {
		gtx = gtx.Disabled()
	}
	gtx.Constraints.Min = image.Point{} // don't stretch inside Column
	for l.click.Clicked(gtx) {
		core.Call(gtx, l.onClick)
	}
	return core.Semantic(gtx, func(gtx C) D {
		return l.click.Layout(gtx, func(gtx C) D {
			if gtx.Enabled() {
				pointer.CursorPointer.Add(gtx.Ops)
			}
			lb := material.Label(theme.Material, theme.BodySize, l.text)
			lb.Color = theme.Primary
			if !gtx.Enabled() {
				lb.Color = theme.Muted
			}
			return lb.Layout(gtx)
		})
	}, semantic.Button, core.Role("link"), semantic.LabelOp(l.text), semantic.EnabledOp(gtx.Enabled()))
}
