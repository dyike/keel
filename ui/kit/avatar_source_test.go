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
)

func TestAvatarSourceLoadingFailureRetryAndReplacement(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, 8, 4))); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	var fail atomic.Bool
	started, cancelled := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/slow" {
			close(started)
			<-r.Context().Done()
			close(cancelled)
			return
		}
		if fail.Load() {
			http.Error(w, "failed", 500)
			return
		}
		w.Write(buf.Bytes())
	}))
	defer server.Close()
	a := Avatar("Ada").Source(server.URL + "/slow")
	defer a.Source("")
	h := renderView(a, 100, 1)
	before := bounds(h, "Ada")
	if !a.Loading() || !shown(h, "A") {
		t.Fatal("loading initials missing")
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not start")
	}
	a.Source(server.URL)
	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("old request was not cancelled")
	}
	wait := func() {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for a.Loading() && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
			h.Frame()
		}
		if a.Loading() {
			t.Fatal("load timed out")
		}
	}
	wait()
	if !a.hasImage || a.ImageError() != nil || a.pixels.Size() != image.Pt(8, 4) || bounds(h, "Ada") != before {
		t.Fatal("replacement source not loaded at stable size")
	}
	n := requests.Load()
	a.Source(server.URL)
	if a.Loading() || requests.Load() != n {
		t.Fatal("same source reloaded")
	}
	fail.Store(true)
	a.Retry()
	wait()
	if a.hasImage || a.ImageError() == nil || !shown(h, "A") {
		t.Fatal("failure did not restore initials")
	}
	fail.Store(false)
	a.Retry()
	wait()
	if !a.hasImage || a.ImageError() != nil {
		t.Fatal("retry did not restore image")
	}
	// Queue a request, then replace it before any queued completion can commit.
	a.Retry()
	a.Image(image.NewNRGBA(image.Rect(0, 0, 3, 3)))
	for range 10 {
		time.Sleep(time.Millisecond)
		h.Frame()
	}
	if a.Loading() || a.ImageError() != nil || a.pixels.Size() != image.Pt(3, 3) {
		t.Fatal("stale response replaced explicit image")
	}
	a.Source("")
	h.Frame()
	if a.hasImage || a.Loading() || !shown(h, "A") {
		t.Fatal("empty source did not reset decoded image")
	}
}
