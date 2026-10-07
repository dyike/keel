package window

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestRoundedScrollbarEnds(t *testing.T) {
	bg := color.NRGBA{R: 240, G: 220, B: 230, A: 255}
	for _, horizontal := range []bool{false, true} {
		for _, scale := range []float32{1, 2} {
			t.Run(fmt.Sprintf("horizontal%v/scale%g", horizontal, scale), func(t *testing.T) {
				var cx *el.Context
				root := el.Root(el.ViewFunc(func(c *el.Context) el.Element {
					cx = c
					box := el.Div().ID("strip").Rounded(18).Bg(bg).Scrollbars(el.ScrollbarAlways)
					child := el.Div()
					if horizontal {
						box.W(el.Dp(220)).H(el.Dp(36)).ScrollX()
						child.W(el.Dp(440)).H(el.Dp(36))
					} else {
						box.W(el.Dp(36)).H(el.Dp(220)).ScrollY()
						child.W(el.Dp(36)).H(el.Dp(440))
					}
					return el.Div().Items(el.Start).Child(box.Child(child))
				}))
				path := filepath.Join(t.TempDir(), "rounded.png")
				for _, end := range []bool{false, true} {
					if end {
						if horizontal {
							cx.ScrollToX("strip", 220)
						} else {
							cx.ScrollTo("strip", 220)
						}
					}
					if err := ScreenshotAtScale(root, 240, 240, scale, path); err != nil {
						t.Fatal(err)
					}
					f, err := os.Open(path)
					if err != nil {
						t.Fatal(err)
					}
					im, err := png.Decode(f)
					f.Close()
					if err != nil {
						t.Fatal(err)
					}
					pixel := func(main int) color.NRGBA {
						x, y := main, 31
						if !horizontal {
							x, y = y, x
						}
						return color.NRGBAModel.Convert(im.At(int(float32(x)*scale), int(float32(y)*scale))).(color.NRGBA)
					}
					clear, thumb := 15, 23
					if end {
						clear, thumb = 205, 197
					}
					if pixel(clear) != bg || pixel(thumb) == bg {
						t.Fatalf("end%v: corner=%v thumb=%v, background=%v", end, pixel(clear), pixel(thumb), bg)
					}
				}
			})
		}
	}
}
