package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestAttachmentVerticalImageCover(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			c := color.NRGBA{G: 255, A: 255}
			if x < 16 {
				c = color.NRGBA{R: 255, A: 255}
			}
			if x >= 48 {
				c = color.NRGBA{B: 255, A: 255}
			}
			pixels.SetNRGBA(x, y, c)
		}
	}
	a := kit.Attachment("photo", 0).Vertical(true).Media(kit.Image(pixels, "Preview")).OnRemove(func() {})
	w := openTest(t, Options{Width: 320, Height: 360, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(a.Render(cx)) }))})
	e := element(t, w, "Preview")
	if e.Width != e.Height || e.Width < 180 {
		t.Fatal("not a square preview", e)
	}
	data, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []int{e.X + 10, e.X + e.Width - 10} {
		c := color.NRGBAModel.Convert(im.At(x, e.Y+e.Height/2)).(color.NRGBA)
		if c.G < 240 || c.R > 20 || c.B > 20 {
			t.Fatal("image not centered/cropped to cover", c)
		}
	}
}

func TestAttachmentTileFocusPaintsAboveImage(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			pixels.SetNRGBA(x, y, color.NRGBA{G: 255, A: 255})
		}
	}
	opened := 0
	a := kit.Attachment("tile", 0).Vertical(true).ShowContent(false).Media(kit.Image(pixels, "Tile preview")).OnOpen(func() { opened++ })
	w := openTest(t, Options{Width: 300, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(a.Render(cx)) }))})
	sample := func() color.NRGBA {
		t.Helper()
		e := element(t, w, "Tile preview")
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(e.X+1, e.Y+e.Height/2)).(color.NRGBA)
	}
	before := sample()
	w.click(element(t, w, "Tile preview").center())
	after := sample()
	if opened != 1 || after == before || after.B < before.B+30 {
		t.Fatal("tile focus hidden by image", opened, before, after)
	}
	a.ShowMedia(false)
	for _, e := range w.snapshot() {
		if e.Name == "Tile preview" {
			t.Fatal("hidden media retained in agent")
		}
	}
}
