package kit

import (
	"gioui.org/f32"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image"
	"image/color"
	"math"
	"time"
)

// SkeletonView is decorative loading geometry. Shimmer is an option, not a
// separate component, and both animations honor ReducedMotion.
type SkeletonView struct {
	w, h            el.Length
	circle, shimmer bool
	secondary       bool
	radius          *float32
}

func Skeleton() *SkeletonView                       { return &SkeletonView{w: el.Full, h: el.Dp(16)} }
func (v *SkeletonView) W(w el.Length) *SkeletonView { v.w = w; return v }
func (v *SkeletonView) H(h el.Length) *SkeletonView { v.h = h; return v }
func (v *SkeletonView) Circle() *SkeletonView       { v.circle = true; return v }
func (v *SkeletonView) Shimmer() *SkeletonView      { v.shimmer = true; return v }

// Secondary halves the opacity of the whole placeholder, including its animation.
func (v *SkeletonView) Secondary(on bool) *SkeletonView { v.secondary = on; return v }

// Rounded sets a rectangle's corner radius in dp; zero makes square corners.
// It replaces Circle. Negative and non-finite values are ignored.
func (v *SkeletonView) Rounded(dp float32) *SkeletonView {
	if dp >= 0 && !math.IsInf(float64(dp), 0) {
		v.radius = &dp
		v.circle = false
	}
	return v
}
func (v *SkeletonView) Render(cx *el.Context) el.Element {
	reduced := el.ReducedMotion()
	phase := float32(0)
	if !reduced {
		phase = float32(cx.Now().UnixNano()%int64(1500*time.Millisecond)) / float32(1500*time.Millisecond)
		cx.Animating()
	}
	base, shine := theme.Subtle, theme.SubtleHover
	radiusDp := float32(theme.RadiusSm)
	if v.radius != nil {
		radiusDp = *v.radius
	}
	graphic := el.Widget(core.Func(func(gtx core.C) core.D {
		size := gtx.Constraints.Min
		r := image.Rectangle{Max: size}
		scale := gtx.Metric.PxPerDp
		if scale == 0 {
			scale = 1
		}
		radius := int(min(float64(min(size.X, size.Y)/2), math.Round(float64(radiusDp)*float64(scale))))
		if v.circle {
			diameter := min(size.X, size.Y)
			r = image.Rect((size.X-diameter)/2, (size.Y-diameter)/2, (size.X+diameter)/2, (size.Y+diameter)/2)
			radius = diameter / 2
		}
		defer clip.UniformRRect(r, radius).Push(gtx.Ops).Pop()
		c := base
		if !reduced && !v.shimmer {
			a := float32(.5 - .5*math.Cos(float64(phase)*2*math.Pi))
			mix := func(x, y uint8) uint8 { return uint8(float32(x)*(1-a) + float32(y)*a) }
			c = color.NRGBA{R: mix(base.R, shine.R), G: mix(base.G, shine.G), B: mix(base.B, shine.B), A: 255}
		}
		paint.Fill(gtx.Ops, c)
		if !reduced && v.shimmer {
			w := float32(size.X)
			x := phase*w*2 - w/2
			band := w / 3
			paint.LinearGradientOp{Stop1: f32.Pt(x-band, 0), Stop2: f32.Pt(x, 0), Color1: base, Color2: shine}.Add(gtx.Ops)
			left := clip.Rect(image.Rect(int(x-band), 0, int(x), size.Y)).Push(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			left.Pop()
			paint.LinearGradientOp{Stop1: f32.Pt(x, 0), Stop2: f32.Pt(x+band, 0), Color1: shine, Color2: base}.Add(gtx.Ops)
			right := clip.Rect(image.Rect(int(x), 0, int(x+band), size.Y)).Push(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			right.Pop()
		}
		return core.D{Size: size}
	})).W(v.w).H(v.h)
	if v.secondary {
		graphic.Opacity(.5)
	}
	return graphic
}
