package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
)

func TestAttachmentMediaStatusAndRetry(t *testing.T) {
	opened, retried := 0, 0
	preview := el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().Role("image").Name("preview").Size(el.Dp(80)).Bg(color.NRGBA{R: 255, A: 255})
	})
	a := kit.Attachment("photo", 0).Media(preview).OnOpen(func() { opened++ }).OnRetry(func() { retried++ })
	a.SetStatus(kit.AttachmentStatusPending)
	disabled := false
	w := openTest(t, Options{Width: 400, Height: 240, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Disabled(disabled).Items(el.Start).Child(a.Render(cx))
	}))})
	previewBounds := element(t, w, "preview")
	sample := func() uint8 {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		e := previewBounds
		return color.NRGBAModel.Convert(im.At(e.X+12, e.Y+12)).(color.NRGBA).R
	}
	plain := sample()
	if plain < 250 {
		t.Fatal("preview missing", plain)
	}
	a.SetProgress(.4)
	uploading := sample()
	if uploading >= plain-20 {
		t.Fatal("missing upload scrim", uploading)
	}
	progress := element(t, w, locale.Current().Name(locale.Current().Uploading, "photo"))
	if progress.Value != "40%" {
		t.Fatal("progress semantics", progress)
	}
	a.SetStatus(kit.AttachmentStatusProcessing)
	if sample() != uploading {
		t.Fatal("processing scrim differs")
	}
	processing := element(t, w, locale.Current().Name(locale.Current().AttachmentProcessing, "photo"))
	if processing.Value != "indeterminate" {
		t.Fatal("processing semantics", processing)
	}
	a.SetError("failed")
	if sample() >= uploading-10 {
		t.Fatal("failure not darker")
	}
	// The overlay retry is painted before the metadata/action retry.
	var retry Element
	for _, e := range w.snapshot() {
		if e.Name == locale.Current().Name(locale.Current().Retry, "photo") && e.X < previewBounds.X+previewBounds.Width {
			retry = e
			break
		}
	}
	if retry.Width == 0 {
		t.Fatal("missing media retry")
	}
	media := element(t, w, "preview")
	if retry.X < media.X || retry.X+retry.Width > media.X+media.Width {
		t.Fatal("retry not in preview", retry, media)
	}
	disabled = true
	w.click(retry.center())
	if retried != 0 {
		t.Fatal("ancestor disabled retry")
	}
	disabled = false
	w.click(retry.center())
	if retried != 1 || opened != 0 || a.Status() != kit.AttachmentStatusUploading {
		t.Fatal("retry transition", retried, opened, a.Status())
	}
	a.OnRetry(nil).SetError("permanent")
	if sample() >= uploading-10 {
		t.Fatal("non-retry failure scrim")
	}
	a.SetStatus(kit.AttachmentStatusComplete)
	if sample() != plain {
		t.Fatal("completion did not restore preview")
	}
	w.click(previewBounds.center())
	if opened != 1 {
		t.Fatal("completed preview not openable")
	}
}
