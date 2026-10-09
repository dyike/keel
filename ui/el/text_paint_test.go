package el

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/font/gofont"
	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/text"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/third_party/gio/widget/material"
	"github.com/dyike/keel/ui/theme"
)

func TestFragmentedLabelMatchesGioPixels(t *testing.T) {
	w, err := headless.NewWindow(400, 150)
	if err != nil {
		t.Skipf("GPU unavailable: %v", err)
	}
	defer w.Release()
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	var ops op.Ops
	for _, th := range []*material.Theme{th, theme.Material} {
		for _, value := range []string{"R14·12345", "0123456789", "Hello world", "", "one1\ntwo2", "a long label 123 that wraps", "مرحبا 12", "中文123", "AV fi123", "12345678901234567890"} {
			for _, scale := range []float32{1, 1.5, 2} {
				for _, affineScale := range []float32{1, 1.25} {
					for _, width := range []int{35, 300} {
						for _, align := range []text.Alignment{text.Start, text.Middle, text.End} {
							lb := material.Label(th, 12, value)
							lb.Color = color.NRGBA{R: 230, G: 190, B: 160, A: 255}
							lb.Alignment = align
							imgs := [2]*image.RGBA{}
							for version := range imgs {
								ops.Reset()
								paint.Fill(&ops, color.NRGBA{R: 20, G: 30, B: 40, A: 255})
								tr := op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(affineScale, affineScale)).Offset(f32.Pt(10.25, 10.5))).Push(&ops)
								gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale}, Constraints: layout.Constraints{Min: image.Pt(width, 0), Max: image.Pt(width, 100)}}
								if version == 0 {
									lb.Layout(gtx)
								} else {
									e := engine{}
									e.paintLabel(gtx, lb)
								}
								tr.Pop()
								if err := w.Frame(&ops); err != nil {
									t.Fatal(err)
								}
								imgs[version] = image.NewRGBA(image.Rect(0, 0, 400, 150))
								if err := w.Screenshot(imgs[version]); err != nil {
									t.Fatal(err)
								}
							}
							maxDelta, sumDelta, ink := 0, 0, 0
							for i, want := range imgs[0].Pix {
								background := []byte{20, 30, 40, 255}[i%4]
								if want != background {
									ink++
								}
								delta := int(imgs[1].Pix[i]) - int(want)
								if delta < 0 {
									delta = -delta
								}
								maxDelta = max(maxDelta, delta)
								sumDelta += delta
							}
							// Separate masks may differ by a few antialiasing levels at an
							// edge. Reject displaced ink and any meaningful overall change.
							if maxDelta > 24 || sumDelta > max(ink, 1) {
								t.Fatalf("%q font=%s scale=%v width=%d align=%v max=%d mean=%v", value, th.Face, scale, width, align, maxDelta, float64(sumDelta)/float64(max(ink, 1)))
							}
						}
					}
				}
			}
		}
	}
}
