package widget

import (
	"image"
	"runtime"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// KbdView displays a shortcut. It does not register a keyboard handler.
type KbdView struct {
	shortcut string
	size     ComponentSize
	plain    bool
}

// Kbd uses core.ParseShortcut syntax (e.g. mod+shift+p). Invalid chords are
// displayed literally, allowing arbitrary key labels as well.
func Kbd(shortcut string) *KbdView                  { return &KbdView{shortcut: shortcut} }
func (k *KbdView) Size(size ComponentSize) *KbdView { k.size = size; return k }
func (k *KbdView) Plain() *KbdView                  { k.plain = true; return k }
func (k *KbdView) SetShortcut(shortcut string)      { k.shortcut = shortcut }
func (k *KbdView) Layout(gtx C) D {
	gtx.Constraints.Min = image.Point{}
	return core.Semantic(gtx, func(gtx C) D {
		content := func(gtx C) D {
			return giolayout.Inset{Left: 6, Right: 6, Top: 3, Bottom: 3}.Layout(gtx, func(gtx C) D {
				size, _, _ := k.size.metrics()
				st := material.Label(theme.Material, size-2, core.ShortcutLabel(k.shortcut, runtime.GOOS))
				st.Color = theme.Muted
				return layoutLabel(gtx, st)
			})
		}
		if k.plain {
			return content(gtx)
		}
		return widget.Border{Color: theme.Border, CornerRadius: 4, Width: 1}.Layout(gtx, content)
	}, semantic.LabelOp(k.shortcut))
}
