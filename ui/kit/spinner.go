package kit

import (
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"math"
	"time"
)

type SpinnerView struct {
	size  float32
	label string
}

func Spinner() *SpinnerView { return &SpinnerView{size: 20, label: "加载中"} }
func (v *SpinnerView) Size(dp float32) *SpinnerView {
	if dp > 0 {
		v.size = dp
	}
	return v
}
func (v *SpinnerView) Label(s string) *SpinnerView { v.label = s; return v }
func (v *SpinnerView) Render(cx *el.Context) el.Element {
	phase := float32(0)
	if !el.ReducedMotion() {
		phase = float32(cx.Now().UnixNano()%int64(time.Second)) / float32(time.Second)
		cx.Animating()
	}
	c := theme.PrimaryText
	ring := el.Widget(core.Func(func(gtx core.C) core.D {
		size := gtx.Constraints.Min
		defer op.Affine(f32.Affine2D{}.Rotate(f32.Pt(float32(size.X)/2, float32(size.Y)/2), phase*2*math.Pi)).Push(gtx.Ops).Pop()
		p := material.ProgressCircle(theme.Material, .7)
		p.Color = c
		return p.Layout(gtx)
	})).Size(el.Dp(v.size)).NoShrink()
	box := el.Div().Role("progressbar").Name(v.label).Value("indeterminate").Row().Items(el.Center).Gap(8).Child(ring)
	if v.label != "" {
		box.Child(el.Text(v.label).TextColor(theme.Muted))
	}
	return box
}
