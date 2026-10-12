package markdown

import (
	"fmt"
	"image"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/theme"
)

func TestCodeTooltipReadableWithDarkButtonForeground(t *testing.T) {
	old := theme.Current()
	defer theme.Apply(old)
	for _, scale := range []int{1, 2} {
		for _, label := range []string{"自动换行", "复制"} {
			t.Run(fmt.Sprintf("scale=%d/label=%s", scale, label), func(t *testing.T) {
				palette := theme.Dark()
				// Apps with light primary buttons use a dark OnColor in dark mode.
				palette.OnColor = theme.RGB(0x222222)
				theme.Apply(palette)
				gpu, err := headless.NewWindow(400*scale, 200*scale)
				if err != nil {
					t.Fatal(err)
				}
				defer gpu.Release()
				var ops op.Ops
				paint.Fill(&ops, theme.RGB(0x222222))
				offset := op.Offset(image.Pt(240*scale, 120*scale)).Push(&ops)
				theme.BeginTextFrame()
				codeTooltip(layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(32*scale, 32*scale)), Metric: unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}}, label)
				offset.Pop()
				if err := gpu.Frame(&ops); err != nil {
					t.Fatal(err)
				}
				pixels := image.NewRGBA(image.Rect(0, 0, 400*scale, 200*scale))
				if err := gpu.Screenshot(pixels); err != nil {
					t.Fatal(err)
				}
				bright := 0
				for y := 0; y < 200*scale; y++ {
					for x := 0; x < 400*scale; x++ {
						c := pixels.RGBAAt(x, y)
						if c.R > 180 && c.G > 180 && c.B > 180 {
							bright++
						}
					}
				}
				if bright < 10*scale*scale {
					t.Fatal("dark tooltip has no readable light text")
				}
			})
		}
	}
}
