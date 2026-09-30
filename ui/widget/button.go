package widget

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

type buttonKind int

const (
	primary buttonKind = iota
	secondary
	danger
)

type Btn struct {
	text     string
	onClick  func()
	kind     buttonKind
	disabled bool
	click    widget.Clickable
}

// Button creates a primary button. Chain Secondary or Danger to restyle it.
func Button(text string, onClick func()) *Btn { return &Btn{text: text, onClick: onClick} }

func (b *Btn) Secondary() *Btn      { b.kind = secondary; return b }
func (b *Btn) Danger() *Btn         { b.kind = danger; return b }
func (b *Btn) SetText(s string)     { b.text = s }
func (b *Btn) SetDisabled(d bool)   { b.disabled = d }
func (b *Btn) SetOnClick(fn func()) { b.onClick = fn }

func (b *Btn) colors() (bg, fg color.NRGBA) {
	switch b.kind {
	case secondary:
		return theme.Subtle, theme.Text
	case danger:
		return theme.Danger, theme.OnColor
	}
	return theme.Primary, theme.OnColor
}

func (b *Btn) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	if b.disabled {
		gtx = gtx.Disabled()
	}
	for b.click.Clicked(gtx) {
		core.Call(gtx, b.onClick)
	}
	st := material.Button(theme.Material, &b.click, b.text)
	st.Background, st.Color = b.colors()
	st.CornerRadius = 6
	st.TextSize = 14
	st.Inset = giolayout.Inset{Top: 8 + theme.CJKNudge, Bottom: 8 - theme.CJKNudge, Left: 16, Right: 16}
	// Gio drops the clickable's node when disabled; this one stays either way.
	return core.Semantic(gtx, st.Layout, semantic.Button, semantic.LabelOp(b.text), semantic.EnabledOp(!b.disabled))
}
