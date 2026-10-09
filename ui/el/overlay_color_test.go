package el

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/headless"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestOverlayScrimFollowsRuntimeTheme(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Release()
	root := Root(ViewFunc(func(cx *Context) Element {
		cx.Overlay("modal", Modal(Div().Size(Dp(50)).Bg(theme.Surface)))
		// Transparent background lets the framebuffer expose the scrim's actual
		// alpha, independently of background color or blending-space assumptions.
		return Div().Bg(color.NRGBA{})
	}))
	var pixel color.RGBA
	h := uitest.NewFunc(func(gtx core.C) {
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rect(0, 0, 400, 300))
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		pixel = img.RGBAAt(5, 5)
	})
	for _, p := range []theme.Palette{theme.Light(), theme.Dark(), theme.Light()} {
		core.Update(func() { theme.Apply(p) })
		h.Frame()
		if pixel.A != p.Scrim.A {
			t.Fatalf("scrim alpha=%d want=%d", pixel.A, p.Scrim.A)
		}
	}
}
