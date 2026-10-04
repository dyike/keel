package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

// EmptyOutline draws a border on a transparent surface; EmptyMuted fills
// with the subtle color.
func TestKitEmptyVariants(t *testing.T) {
	saved := theme.Current()
	theme.Apply(theme.Light())
	t.Cleanup(func() { theme.Apply(saved) })
	outline := kit.Empty("拖到这里上传").Variant(kit.EmptyOutline)
	muted := kit.Empty("暂无数据").Variant(kit.EmptyMuted)
	w := openTest(t, Options{Width: 300, Height: 360, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(16).Gap(16).Items(el.Stretch).Child(outline.Render(cx), muted.Render(cx))
	}))})
	w.render()
	shot, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	img, _ := png.Decode(bytes.NewReader(shot))
	at := func(x, y int) color.NRGBA { return color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA) }
	same := func(a, b color.NRGBA) bool { return a.R == b.R && a.G == b.G && a.B == b.B }
	o, m := element(t, w, "拖到这里上传"), element(t, w, "暂无数据")
	// Inside the outline, away from the text, the window background shows.
	if c := at(20, o.Y+o.Height/2); !same(c, theme.Bg) {
		t.Errorf("outline surface is not transparent: %v", c)
	}
	if c := at(20, m.Y+m.Height/2); !same(c, theme.Subtle) {
		t.Errorf("muted surface is not Subtle: %v", c)
	}
	// Somewhere on the outline's left edge there is border ink.
	ink := false
	for y := o.Y - 40; y < o.Y+40; y++ {
		if c := at(16, y); !same(c, theme.Bg) {
			ink = true
		}
	}
	if !ink {
		t.Error("no outline border drawn")
	}
}
