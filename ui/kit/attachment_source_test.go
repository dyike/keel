package kit

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestAttachmentSourceCancellationRetryAndReplacement(t *testing.T) {
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewNRGBA(image.Rect(0, 0, 8, 4))); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	var fail atomic.Bool
	started, canceled := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/slow" {
			close(started)
			<-r.Context().Done()
			close(canceled)
			return
		}
		if fail.Load() {
			http.Error(w, "failed", 500)
			return
		}
		w.Write(data.Bytes())
	}))
	defer server.Close()
	uploadRetries, opened := 0, 0
	a := Attachment("photo", 1024).OnOpen(func() { opened++ }).OnRetry(func() { uploadRetries++ }).MediaSource(server.URL+"/slow").PartStyle(AttachmentPartMedia, func(e *el.DivEl) { e.Name("Source box") })
	defer a.MediaSource("")
	h := renderView(a, 240, 1)
	before := bounds(h, "Source box")
	if !a.MediaLoading() {
		t.Fatal("source did not start")
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("request not started")
	}
	a.MediaSource(server.URL)
	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("old request not canceled")
	}
	wait := func() {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for a.MediaLoading() && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
			h.Frame()
		}
		if a.MediaLoading() {
			t.Fatal("load timed out")
		}
		h.Frame()
	}
	wait()
	if a.MediaError() != nil || a.imageSource.pixels.Size() != image.Pt(8, 4) || bounds(h, "Source box") != before {
		t.Fatal("load or stable bounds")
	}
	click(t, h, "Source box")
	if opened != 1 {
		t.Fatal("loaded source preview did not open")
	}
	opened = 0
	n := requests.Load()
	a.MediaSource(server.URL)
	if a.MediaLoading() || requests.Load() != n {
		t.Fatal("same source reloaded")
	}
	fail.Store(true)
	a.RetryMedia()
	wait()
	if a.MediaError() == nil || a.Status() != AttachmentStatusComplete || bounds(h, "Source box") != before {
		t.Fatal("image failure changed lifecycle or size")
	}
	fail.Store(false)
	a.SetDisabled(true)
	h.Frame()
	click(t, h, locale.Current().Name(locale.Current().Retry, "photo"))
	if a.MediaLoading() || requests.Load() != n+1 {
		t.Fatal("disabled image retry ran")
	}
	a.SetDisabled(false)
	h.Frame()
	click(t, h, locale.Current().Name(locale.Current().Retry, "photo"))
	wait()
	if a.MediaError() != nil || uploadRetries != 0 || opened != 0 {
		t.Fatal("preview retry changed upload")
	}
	// A queued result must not replace an explicit media view.
	a.RetryMedia()
	explicit := Image(image.NewNRGBA(image.Rect(0, 0, 3, 3)), "Explicit")
	a.Media(explicit)
	for range 10 {
		time.Sleep(time.Millisecond)
		h.Frame()
	}
	if a.imageSource != nil || a.media != explicit || a.MediaLoading() || a.MediaError() != nil {
		t.Fatal("stale response replaced media")
	}
	a.MediaSource("")
	h.Frame()
	if a.media != nil || a.imageSource != nil || a.MediaLoading() {
		t.Fatal("source clear did not restore default")
	}
}
