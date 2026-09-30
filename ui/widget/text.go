package widget

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// Label is read-only text that wraps to the available width.
type Label struct {
	text  string
	size  unit.Sp
	bold  bool
	color *color.NRGBA
}

func Text(s string) *Label    { return &Label{text: s, size: theme.BodySize} }
func Heading(s string) *Label { return &Label{text: s, size: theme.HeadingSize, bold: true} }

// Muted is smaller, secondary text.
func Muted(s string) *Label { return &Label{text: s, size: theme.SmallSize, color: &theme.Muted} }

func (l *Label) Text() string     { return l.text }
func (l *Label) SetText(s string) { l.text = s }

func (l *Label) Layout(gtx C) D {
	lb := material.Label(theme.Material, l.size, l.text)
	lb.Color = theme.Text
	if l.color != nil {
		lb.Color = *l.color
	}
	if l.bold {
		lb.Font.Weight = font.Bold
	}
	// Gio's label node spans the whole available height; this one has the
	// text's real bounds, and the inner node becomes its child.
	return core.Semantic(gtx, lb.Layout, semantic.LabelOp(l.text))
}
