package widget

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/theme"
)

// ComponentSize is shared by controls. The zero value is Medium.
type ComponentSize uint8

const (
	Medium ComponentSize = iota
	Small
	Large
)

func (s ComponentSize) metrics() (text unit.Sp, padding, icon unit.Dp) {
	switch s {
	case Small:
		return 12, 5, 14
	case Large:
		return 16, 11, 20
	default:
		return 14, 8, 16
	}
}

// loadingIndicator uses frame time, never timers or a separate goroutine.
func loadingIndicator(gtx C, size unit.Dp, col color.NRGBA) D {
	d := gtx.Constraints.Constrain(image.Pt(gtx.Dp(size), gtx.Dp(size)))
	gtx.Constraints.Min, gtx.Constraints.Max = d, d
	phase := float32(gtx.Now.UnixNano()%int64(time.Second)) / float32(time.Second)
	defer op.Affine(f32.Affine2D{}.Rotate(f32.Pt(float32(d.X)/2, float32(d.Y)/2), float32(2*math.Pi)*phase)).Push(gtx.Ops).Pop()
	p := material.ProgressCircle(theme.Material, .7)
	p.Color = col
	p.Layout(gtx)
	gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(time.Second / 30)})
	return D{Size: d}
}
