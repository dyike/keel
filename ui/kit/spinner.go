package kit

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"math"
	"time"
)

type SpinnerView struct {
	size     float32
	label    string
	labelSet bool
}

// Spinner shows an indeterminate progress ring labelled with the locale's
// Loading text until Label sets another one ("" hides the text).
func Spinner() *SpinnerView { return &SpinnerView{size: 20} }
func (v *SpinnerView) Size(dp float32) *SpinnerView {
	if dp > 0 {
		v.size = dp
	}
	return v
}
func (v *SpinnerView) Label(s string) *SpinnerView { v.label, v.labelSet = s, true; return v }
func (v *SpinnerView) Render(cx *el.Context) el.Element {
	label := v.label
	if !v.labelSet {
		label = locale.Current().Loading
	}
	ring := spinnerRing(cx, v.size, theme.PrimaryText)
	box := el.Div().Role("progressbar").Name(label).Value("indeterminate").Row().Items(el.Center).Gap(8).Child(ring)
	if label != "" {
		box.Child(el.Text(label).TextColor(theme.Muted))
	}
	return box
}

// spinnerRing is shared by Spinner and loading controls without adding semantics.
func spinnerRing(cx *el.Context, size float32, c color.NRGBA) el.Element {
	phase := float32(0)
	if !el.ReducedMotion() {
		phase = float32(cx.Now().UnixNano()%int64(time.Second)) / float32(time.Second)
		cx.Animating()
	}
	return el.Widget(core.Func(func(gtx core.C) core.D {
		size := gtx.Constraints.Min
		defer op.Affine(f32.Affine2D{}.Rotate(f32.Pt(float32(size.X)/2, float32(size.Y)/2), phase*2*math.Pi)).Push(gtx.Ops).Pop()
		p := material.ProgressCircle(theme.Material, .7)
		p.Color = c
		return p.Layout(gtx)
	})).Size(el.Dp(size)).NoShrink()
}
