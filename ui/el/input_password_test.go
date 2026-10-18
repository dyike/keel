package el

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestPasswordMaskVerticallyCentered(t *testing.T) {
	previous := theme.Material.Face
	defer func() { theme.Material.Face = previous }()
	for _, face := range []font.Typeface{theme.Face, theme.MonoFace, "Go"} {
		theme.Material.Face = face
		for _, scale := range []int{1, 2} {
			for _, size := range []float32{12, 16, 20} {
				value := "mask-test"
				root := Root(ViewFunc(func(cx *Context) Element {
					return Div().Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).Items(Start).Child(
						Input().Password().Bind(&value).TextSize(size).TextColor(color.NRGBA{R: 255, A: 255}).W(Dp(240)).H(Dp(44)).Border(0, color.NRGBA{}),
					)
				}))
				pixels := image.NewRGBA(image.Rect(0, 0, 300*scale, 80*scale))
				gpu, err := headless.NewWindow(pixels.Bounds().Dx(), pixels.Bounds().Dy())
				if err != nil {
					t.Fatal(err)
				}
				harness := uitest.NewFunc(func(gtx core.C) {
					gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					gtx.Constraints = layout.Exact(pixels.Bounds().Size())
					root.Layout(gtx)
					if err := gpu.Frame(gtx.Ops); err != nil {
						t.Fatal(err)
					}
					if err := gpu.Screenshot(pixels); err != nil {
						t.Fatal(err)
					}
				})
				harness.Frame()
				gpu.Release()
				ink := image.Rectangle{}
				for vertical := 0; vertical < 44*scale; vertical++ {
					for horizontal := 0; horizontal < 240*scale; horizontal++ {
						pixel := pixels.RGBAAt(horizontal, vertical)
						if int(pixel.R)-int(pixel.G) > 40 {
							ink = ink.Union(image.Rect(horizontal, vertical, horizontal+1, vertical+1))
						}
					}
				}
				if offset := ink.Min.Y + ink.Max.Y - 44*scale; ink.Empty() || offset < -2 || offset > 2 {
					t.Errorf("face=%s scale=%d size=%g mask not centered: %v", face, scale, size, ink)
				}
			}
		}
	}
}
