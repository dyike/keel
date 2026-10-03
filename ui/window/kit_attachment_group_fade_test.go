package window

import (
	"bytes"
	"gioui.org/f32"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestAttachmentGroupEdgeFadePixelsAndHit(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	calls := 0
	tile := func() el.View {
		return el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().W(el.Dp(240)).H(el.Dp(60)).Bg(red).OnClick(func() { calls++ })
		})
	}
	first := tile()
	group := kit.AttachmentGroup(first, tile(), tile()).Name("Files").Gap(0)
	w := openTest(t, Options{Width: 400, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(300)).Child(group.Render(cx)) }))})
	sample := func() (color.NRGBA, color.NRGBA) {
		t.Helper()
		w.render()
		w.render()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		e := element(t, w, "Files")
		return color.NRGBAModel.Convert(im.At(e.X+2, e.Y+25)).(color.NRGBA), color.NRGBAModel.Convert(im.At(e.X+e.Width-3, e.Y+25)).(color.NRGBA)
	}
	left, right := sample()
	if left != red || right != red {
		t.Fatal("baseline", left, right)
	}
	group.EdgeFade(blue)
	left, right = sample()
	if left != red || right.B < 200 || right.R > 128 {
		t.Fatal("start edge", left, right)
	}
	group.ScrollTo(100)
	left, right = sample()
	if left.B < 200 || right.B < 200 {
		t.Fatal("middle edges", left, right)
	}
	e := element(t, w, "Files")
	w.click(f32.Pt(float32(e.X+2), float32(e.Y+25)))
	if calls != 1 {
		t.Fatal("fade intercepted click")
	}
	group.ScrollTo(10000)
	left, right = sample()
	if left.B < 200 || right != red {
		t.Fatal("end edge", left, right)
	}
	group.ClearEdgeFade()
	left, right = sample()
	if left != red || right != red {
		t.Fatal("clear fade")
	}
	group.EdgeFade(blue).SetItems(first)
	left, _ = sample()
	if left != red {
		t.Fatal("non-overflowing group faded")
	}
}
