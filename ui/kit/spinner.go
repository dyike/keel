package kit

import (
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/op"
	giowidget "gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

type SpinnerView struct {
	size     float32
	label    string
	labelSet bool
	icon     *giowidget.Icon
	color    *color.NRGBA
}

// Spinner shows an indeterminate progress ring labelled with the locale's
// Loading text until Label sets another one ("" hides the text).
func Spinner() *SpinnerView { return &SpinnerView{size: 20} }
func (v *SpinnerView) Size(dp float32) *SpinnerView {
	if dp > 0 && !math.IsInf(float64(dp), 0) {
		v.size = dp
	}
	return v
}
func (v *SpinnerView) Label(s string) *SpinnerView { v.label, v.labelSet = s, true; return v }

// Icon replaces the ring with a rotating built-in icon. IconNone restores the ring.
func (v *SpinnerView) Icon(name IconName) *SpinnerView { v.icon = Icon(name).icon; return v }

// VectorIcon uses a custom decoded Gio icon; nil restores the ring.
func (v *SpinnerView) VectorIcon(icon *giowidget.Icon) *SpinnerView { v.icon = icon; return v }

// Color changes the graphic color, leaving the label in the theme's muted color.
func (v *SpinnerView) Color(c color.NRGBA) *SpinnerView { v.color = &c; return v }
func (v *SpinnerView) Render(cx *el.Context) el.Element {
	label := v.label
	if !v.labelSet {
		label = locale.Current().Loading
	}
	c := theme.PrimaryText
	if v.color != nil {
		c = *v.color
	}
	ring := spinnerGraphic(cx, v.size, c, v.icon)
	box := el.Div().Role("progressbar").Name(label).Value("indeterminate").Row().Items(el.Center).Gap(theme.SpaceMd).Child(ring)
	if label != "" {
		box.Child(el.Text(label).TextColor(theme.Muted))
	}
	return box
}

// spinnerRing is shared by Spinner and loading controls without adding semantics.
func spinnerRing(cx *el.Context, size float32, c color.NRGBA) el.Element {
	return spinnerGraphic(cx, size, c, nil)
}

func spinnerGraphic(cx *el.Context, size float32, c color.NRGBA, icon *giowidget.Icon) el.Element {
	phase := float32(0)
	if !el.ReducedMotion() {
		phase = float32(cx.Now().UnixNano()%int64(time.Second)) / float32(time.Second)
		cx.Animating()
	}
	return el.Widget(core.Func(func(gtx core.C) core.D {
		size := gtx.Constraints.Min
		defer op.Affine(f32.Affine2D{}.Rotate(f32.Pt(float32(size.X)/2, float32(size.Y)/2), phase*2*math.Pi)).Push(gtx.Ops).Pop()
		if icon != nil {
			return icon.Layout(gtx, c)
		}
		p := material.ProgressCircle(theme.Material, .7)
		p.Color = c
		return p.Layout(gtx)
	})).Size(el.Dp(size)).NoShrink()
}
