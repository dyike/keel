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

func TestAttachmentStatusBorderPixelsAndRestoration(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	black := color.NRGBA{A: 255}
	a := kit.Attachment("file", 0).PartStyle(kit.AttachmentPartRoot, func(e *el.DivEl) { e.Bg(black).Border(2, red) })
	w := openTest(t, Options{Width: 400, Height: 240, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(a.Render(cx)) }))})
	sample := func() (image.Image, Element) {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		return im, element(t, w, "file")
	}
	gaps := func() int {
		im, e := sample()
		n := 0
		for x := e.X + 20; x < e.X+e.Width-20; x++ {
			c := color.NRGBAModel.Convert(im.At(x, e.Y+1)).(color.NRGBA)
			if c.R < 40 {
				n++
			}
		}
		return n
	}
	if gaps() != 0 {
		t.Fatal("solid border has gaps")
	}
	a.SetStatus(kit.AttachmentStatusPending)
	if g := gaps(); g < 15 || g > 100 {
		t.Fatal("pending dash distribution", g)
	}

	a.PartStyle(kit.AttachmentPartRoot, func(e *el.DivEl) { e.Bg(black).Border(2, red).BorderDashed(false) })
	if gaps() != 0 {
		t.Fatal("style could not restore solid border")
	}
	a.PartStyle(kit.AttachmentPartRoot, nil)
	a.SetStatus(kit.AttachmentStatusFailed)
	im, e := sample()
	got := color.NRGBAModel.Convert(im.At(e.X, e.Y+e.Height/2)).(color.NRGBA)
	if int(got.R)-int(got.G) < 20 {
		t.Fatal("failure border", got)
	}
	a.SetStatus(kit.AttachmentStatusComplete)
	im, e = sample()
	got = color.NRGBAModel.Convert(im.At(e.X, e.Y+e.Height/2)).(color.NRGBA)
	if int(got.R)-int(got.G) > 20 {
		t.Fatal("failure styling retained")
	}
}
