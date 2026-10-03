package window

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
)

func TestAttachmentMediaSourceAgentAndPixels(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			pixels.SetNRGBA(x, y, color.NRGBA{G: 255, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, pixels); err != nil {
		t.Fatal(err)
	}
	a := kit.Attachment("source", 1024).Vertical(true).ShowContent(false).MediaSource("data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes()))
	defer a.MediaSource("")
	w := openTest(t, Options{Width: 300, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(a.Render(cx)) }))})
	wait := func() {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for a.MediaLoading() && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
			w.render()
		}
		if a.MediaLoading() {
			t.Fatal("load timed out")
		}
	}
	wait()
	e := element(t, w, "source")
	if e.Role != "image" || e.Value != "loaded" {
		t.Fatal("loaded image semantics", e)
	}
	data, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	c := color.NRGBAModel.Convert(im.At(e.X+e.Width/2, e.Y+e.Height/2)).(color.NRGBA)
	if c.G < 250 || c.R > 5 || c.B > 5 {
		t.Fatal("decoded pixels not painted", c)
	}
	a.MediaSource("data:image/png;base64,invalid")
	wait()
	e = element(t, w, "source")
	if e.Role != "group" || e.Value != "image-error" || a.Status() != kit.AttachmentStatusComplete {
		t.Fatal("image error semantics", e)
	}
	retryName := locale.Current().Name(locale.Current().Retry, "source")
	if element(t, w, retryName).Role != "button" {
		t.Fatal("missing image retry")
	}
	a.SetStatus(kit.AttachmentStatusProcessing)
	for _, e := range w.snapshot() {
		if e.Name == retryName {
			t.Fatal("image retry overlaps lifecycle progress")
		}
	}
	a.SetStatus(kit.AttachmentStatusFailed)
	for _, e := range w.snapshot() {
		if e.Name == retryName {
			t.Fatal("image retry overlaps rejection")
		}
	}
}
