package widget

import (
	"image"
	"image/color"

	"gioui.org/io/key"
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
	loading  bool
	size     ComponentSize
	icon     *IconView
	click    widget.Clickable
}

// Button creates a primary button. Chain Secondary or Danger to restyle it.
func Button(text string, onClick func()) *Btn { return &Btn{text: text, onClick: onClick} }

func (b *Btn) Secondary() *Btn              { b.kind = secondary; return b }
func (b *Btn) Danger() *Btn                 { b.kind = danger; return b }
func (b *Btn) SetText(s string)             { b.text = s }
func (b *Btn) SetDisabled(d bool)           { b.disabled = d }
func (b *Btn) SetOnClick(fn func())         { b.onClick = fn }
func (b *Btn) Size(size ComponentSize) *Btn { b.size = size; return b }
func (b *Btn) Icon(icon *IconView) *Btn     { b.icon = icon; return b }
func (b *Btn) SetLoading(loading bool)      { b.loading = loading }
func (b *Btn) Loading() bool                { return b.loading }

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
	if b.disabled || b.loading {
		gtx = gtx.Disabled()
	}
	for b.click.Clicked(gtx) {
		if b.disabled || b.loading {
			continue
		}
		gtx.Execute(key.FocusCmd{Tag: &b.click})
		core.Call(gtx, b.onClick)
	}
	st := material.ButtonLayout(theme.Material, &b.click)
	bg, fg := b.colors()
	st.Background = bg
	st.CornerRadius = 6
	textSize, padding, iconSize := b.size.metrics()
	// Gio drops the clickable's node when disabled; this one stays either way.
	return core.Semantic(gtx, func(gtx C) D {
		return st.Layout(gtx, func(gtx C) D {
			return giolayout.Inset{Top: padding, Bottom: padding, Left: padding * 2, Right: padding * 2}.Layout(gtx, func(gtx C) D {
				var children []giolayout.FlexChild
				if b.loading || b.icon != nil {
					children = append(children, giolayout.Rigid(func(gtx C) D {
						if b.loading {
							return loadingIndicator(gtx, iconSize, fg)
						}
						ic := *b.icon
						ic.size = iconSize
						ic.color = &fg
						return ic.Layout(gtx)
					}))
					if b.text != "" {
						children = append(children, giolayout.Rigid(giolayout.Spacer{Width: 6}.Layout))
					}
				}
				children = append(children, giolayout.Rigid(func(gtx C) D {
					lb := material.Label(theme.Material, textSize, b.text)
					lb.Color = fg
					return layoutLabel(gtx, lb)
				}))
				return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx, children...)
			})
		})
	}, semantic.Button, semantic.LabelOp(b.text), semantic.EnabledOp(gtx.Enabled()))
}
