package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"image"
	"math"
	"testing"
)

func TestImageNarrowAspectFitRetryAndCache(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	for _, scale := range []int{1, 2} {
		pic := Image(img, "photo").Width(400)
		h := renderView(pic, 100, scale)
		b := bounds(h, "photo")
		if b.Dx() != 100*scale || b.Dy() != 50*scale {
			t.Fatalf("narrow aspect %v", b)
		}
		pic.Size(100, 80).Fit(ImageCover)
		h.Frame()
		if b := bounds(h, "photo"); b.Dx() != 100*scale || b.Dy() != 80*scale {
			t.Fatalf("fixed size %v", b)
		}
		pic.Width(float32(math.NaN())).Rounded(float32(math.Inf(1)))
		if pic.width != 100 || pic.rounded != 0 {
			t.Fatal("invalid dimensions accepted")
		}
	}
	retried := 0
	pic := Image(img, "photo").OnRetry(func() { retried++ })
	pic.SetError("offline")
	if pic.img != nil || pic.op.Size() != (image.Point{}) {
		t.Fatal("error retained old pixels")
	}
	h := renderView(pic, 300, 1)
	click(t, h, "重试 photo")
	if retried != 1 || shown(h, "重试 photo") || pic.err != "" {
		t.Fatal("retry transition")
	}
	pic.SetImage(img)
	h.Frame()
	if !shown(h, "photo") {
		t.Fatal("retry image absent")
	}
	pic.SetImage(nil)
	if pic.op.Size() != (image.Point{}) {
		t.Fatal("nil retained cache")
	}
}

func TestImagePreviewEscapeAndOwnerDisable(t *testing.T) {
	pic := Image(image.NewNRGBA(image.Rect(0, 0, 120, 60)), "photo").Preview()
	disabled := false
	h := page(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(pic.Render(cx)) }))
	click(t, h, "photo")
	if !pic.dialog.Value() {
		t.Fatal("preview not open")
	}
	h.Key(key.NameEscape, 0)
	if pic.dialog.Value() {
		t.Fatal("escape did not close preview")
	}
	click(t, h, "photo")
	disabled = true
	h.Frame()
	if pic.dialog.Value() {
		t.Fatal("owner disabled but preview survived")
	}
	disabled = false
	h.Frame()
	click(t, h, "photo")
	pic.SetImage(nil)
	h.Frame()
	if pic.dialog.Value() {
		t.Fatal("removed pixels retained preview")
	}
}
