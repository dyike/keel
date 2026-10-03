package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestAttachmentDescriptionStatusPixels(t *testing.T) {
	a := kit.Attachment("doc", 1024).Description("MMMMMMMMMM")
	a.SetError("failure")
	w := openTest(t, Options{Width: 360, Height: 180, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(a.Render(cx)) }))})
	redPixels := func() int {
		t.Helper()
		e := element(t, w, "MMMMMMMMMM")
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for y := e.Y; y < e.Y+e.Height; y++ {
			for x := e.X; x < e.X+e.Width; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if int(c.R)-int(c.G) > 40 {
					count++
				}
			}
		}
		return count
	}
	failure := redPixels()
	if failure < 20 {
		t.Fatal("failed description not tinted", failure)
	}
	a.PartStatus(kit.AttachmentPartDescription, kit.AttachmentStatusComplete)
	if redPixels() != 0 {
		t.Fatal("description override did not clear failure tint")
	}
	if element(t, w, "doc").Value != "error" {
		t.Fatal("presentation override changed attachment semantics")
	}
	a.ClearPartStatus(kit.AttachmentPartDescription)
	if redPixels() != failure {
		t.Fatal("failure tint not restored")
	}
	a.Content(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Custom content") }))
	if element(t, w, "Custom content").Name != "Custom content" {
		t.Fatal("custom content lost")
	}
	a.Content(nil)
	if redPixels() != failure {
		t.Fatal("content restoration lost description config")
	}
}
