//go:build darwin && !ios

package el

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget/material"
	"github.com/dyike/keel/ui/theme"
)

func TestMixedLabelNaturalWidth(t *testing.T) {
	gpu, err := headless.NewWindow(1000, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	th := *theme.Material
	th.Face = "Helvetica Neue, PingFang SC, Hiragino Sans GB, Apple Color Emoji"
	for _, scale := range []float32{1, 1.5, 2} {
		for _, value := range []string{"运行了 1 个工具", "执行命令 · 确认 PR 信息", "PR 已创建并确认。"} {
			var ops op.Ops
			gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale}, Constraints: layout.Constraints{Max: image.Pt(1000, 1000)}}
			lb := material.Label(&th, 13, value)
			lb.Color = color.NRGBA{A: 255}
			e := engine{}
			dims := e.measureLabel(gtx, lb)
			gtx.Constraints.Max = dims.Size
			params := labelParameters(gtx, lb)
			lb.Shaper.LayoutString(params, value)
			count := 0
			for g, ok := lb.Shaper.NextGlyph(); ok; g, ok = lb.Shaper.NextGlyph() {
				if g.Flags&text.FlagLineBreak != 0 {
					count++
				}
				if g.Advance == 0 {
					t.Errorf("%q scale=%v: visible character has zero advance", value, scale)
				}
			}
			var imgs [2]*image.RGBA
			for i, width := range []int{dims.Size.X, 1000} {
				theme.BeginTextFrame()
				ops.Reset()
				paint.Fill(&ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
				gtx.Constraints.Max = image.Pt(width, 100)
				lb.Layout(gtx)
				if err := gpu.Frame(&ops); err != nil {
					t.Fatal(err)
				}
				imgs[i] = image.NewRGBA(image.Rect(0, 0, 1000, 100))
				if err := gpu.Screenshot(imgs[i]); err != nil {
					t.Fatal(err)
				}
			}
			ink := 0
			for i := 0; i < len(imgs[0].Pix); i += 4 {
				if imgs[0].Pix[i] < 240 {
					ink++
				}
			}
			if ink < 20 {
				t.Fatalf("%q scale=%v: no visible text", value, scale)
			}
			if !bytes.Equal(imgs[0].Pix, imgs[1].Pix) {
				t.Errorf("%q scale=%v: measured label clips pixels", value, scale)
			}
			if path := os.Getenv("KEEL_LABEL_PREVIEW"); path != "" && scale == 2 && value == "运行了 1 个工具" {
				file, err := os.Create(path)
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(file, imgs[0])
				file.Close()
				if err != nil {
					t.Fatal(err)
				}
			}

			if count != 1 {
				t.Errorf("%q scale=%v natural width wraps: %v", value, scale, dims)
			}
		}
	}
}
