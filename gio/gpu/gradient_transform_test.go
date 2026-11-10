// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"math"
	"testing"

	"github.com/dyike/keel/gio/f32"
)

func TestGradientSpaceTransformFollowsPaintTransform(t *testing.T) {
	bounds := image.Rect(11, 19, 291, 237)
	start, end := f32.Pt(8, 12), f32.Pt(140, 90)
	transforms := []f32.Affine2D{
		f32.AffineId(),
		f32.AffineId().Offset(f32.Pt(40, 60)),
		f32.AffineId().Scale(f32.Point{}, f32.Pt(.5, 2)).Offset(f32.Pt(40, 60)),
		f32.AffineId().Rotate(f32.Point{}, .7).Offset(f32.Pt(150, 60)),
		f32.NewAffine2D(1, .4, 30, -.2, 1.5, 70),
	}
	for _, transform := range transforms {
		mapping := gradientSpaceTransform(bounds, transform, start, end)
		for _, p := range []float32{0, .25, .5, 1} {
			point := transform.Transform(start.Add(end.Sub(start).Mul(p)))
			uv := f32.Pt((point.X-float32(bounds.Min.X))/float32(bounds.Dx()), (point.Y-float32(bounds.Min.Y))/float32(bounds.Dy()))
			got := mapping.Transform(uv)
			if math.Abs(float64(got.X-p)) > 1e-5 || math.Abs(float64(got.Y)) > 1e-5 {
				t.Fatalf("transform %v at %g: %v", transform, p, got)
			}
		}
	}
	if got := gradientSpaceTransform(bounds, f32.AffineId(), start, start).Transform(f32.Pt(.5, .5)); got != (f32.Point{}) {
		t.Fatalf("coincident stops: %v", got)
	}
}
