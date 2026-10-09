package el

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/headless"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

// A disabled root's DisabledStyle text color reaches its text children; they
// must not reset it to the default muted color.
func TestDisabledStyleTextColorReachesChildren(t *testing.T) {
	gpu, err := headless.NewWindow(200, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	red := color.NRGBA{R: 0xff, A: 0xff}
	root := Root(ViewFunc(func(cx *Context) Element {
		return Div().Bg(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}).Child(
			Div().Disabled(true).DisabledStyle(func(s *Style) { s.TextColor(red) }).
				Child(Text("国国").TextSize(40)))
	}))
	var reddish bool
	h := uitest.NewFunc(func(gtx core.C) {
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rect(0, 0, 200, 100))
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		reddish = false
		for y := 0; y < 100; y++ {
			for x := 0; x < 200; x++ {
				if p := img.RGBAAt(x, y); p.R > 200 && p.G < 80 && p.B < 80 {
					reddish = true
				}
			}
		}
	})
	h.Frame()
	if !reddish {
		t.Fatal("text inside a disabled root did not take its DisabledStyle color")
	}
}
