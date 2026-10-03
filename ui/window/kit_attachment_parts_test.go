package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"image/color"
	"image/png"
	"testing"
)

func TestAttachmentPartStylePixels(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	a := kit.Attachment("file", 0).Vertical(true).PartStyle(kit.AttachmentPartRoot, func(e *el.DivEl) { e.Bg(red).P(20) })
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(a.Render(cx)) }))})
	sample := func() color.NRGBA {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		e := element(t, w, "file")
		return color.NRGBAModel.Convert(im.At(e.X+10, e.Y+e.Height/2)).(color.NRGBA)
	}
	if got := sample(); got != red {
		t.Fatal("part color", got)
	}
	a.PartStyle(kit.AttachmentPartRoot, nil)
	if sample() == red {
		t.Fatal("style retained after clearing")
	}
}
