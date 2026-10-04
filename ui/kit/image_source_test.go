package kit

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func sourceTestPNG(t *testing.T) []byte {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, 8, 4))
	im.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestImageSourceSlotsRetryAndReplacement(t *testing.T) {
	data := sourceTestPNG(t)
	var fail atomic.Bool
	var requests atomic.Int32
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
		w.Write(data)
	}))
	defer server.Close()
	v := Image(nil, "photo").Size(160, 80).Cache(NewImageCache(1024)).LoadingContent(text("Please wait")).Fallback(text("Could not load")).Source(server.URL + "/slow")
	defer v.Source("")
	h := renderView(v, 200, 1)
	if !v.Loading() || !shown(h, "Please wait") {
		t.Fatal("loading slot")
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("not started")
	}
	v.Source(server.URL)
	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("not cancelled")
	}
	wait := func() {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for v.Loading() && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
			h.Frame()
		}
		if v.Loading() {
			t.Fatal("timeout")
		}
	}
	wait()
	if v.img == nil || v.ImageError() != nil || v.op.Size() != image.Pt(8, 4) {
		t.Fatal("loaded image")
	}
	n := requests.Load()
	v.Source(server.URL)
	if v.Loading() || requests.Load() != n {
		t.Fatal("same source reloaded")
	}
	fail.Store(true)
	v.Retry()
	wait()
	if v.img != nil || v.ImageError() == nil || !shown(h, "Could not load") {
		t.Fatal("failure slot")
	}
	fail.Store(false)
	v.Retry()
	wait()
	if v.img == nil || v.ImageError() != nil {
		t.Fatal("retry")
	}
	v.Retry()
	v.SetImage(image.NewNRGBA(image.Rect(0, 0, 3, 3)))
	for range 10 {
		time.Sleep(time.Millisecond)
		h.Frame()
	}
	if v.Loading() || v.source != "" || v.op.Size() != image.Pt(3, 3) {
		t.Fatal("stale completion replaced explicit pixels")
	}
	v.Source(server.URL)
	v.SetError("manual")
	for range 10 {
		time.Sleep(time.Millisecond)
		h.Frame()
	}
	if v.err != "manual" || v.source != "" || v.img != nil {
		t.Fatal("stale completion replaced explicit error")
	}
}

func TestImageCacheSharedRequestCancellation(t *testing.T) {
	data := sourceTestPNG(t)
	start, release := make(chan struct{}), make(chan struct{})
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			close(start)
		}
		select {
		case <-release:
			w.Write(data)
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)
	cache := NewImageCache(1024)
	first, cancel := context.WithCancel(context.Background())
	defer cancel()
	result1, result2 := make(chan error, 1), make(chan error, 1)
	go func() { _, err := cache.load(first, server.URL); result1 <- err }()
	select {
	case <-start:
	case <-time.After(2 * time.Second):
		t.Fatal("start")
	}
	second, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	go func() { _, err := cache.load(second, server.URL); result2 <- err }()
	deadline := time.Now().Add(time.Second)
	for {
		cache.mu.Lock()
		n := cache.pending[server.URL].waiters
		cache.mu.Unlock()
		if n == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("second waiter")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-result1; err != context.Canceled {
		t.Fatal("cancel", err)
	}
	release <- struct{}{}
	if err := <-result2; err != nil {
		t.Fatal("shared request cancelled", err)
	}
	if _, err := cache.load(context.Background(), server.URL); err != nil || requests.Load() != 1 {
		t.Fatal("cache reuse", requests.Load(), err)
	}
}

func TestImageCacheEvictionAndInvalidation(t *testing.T) {
	data := sourceTestPNG(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/error" {
			http.Error(w, "failed", 500)
		} else {
			w.Write(data)
		}
	}))
	defer server.Close()
	cache := NewImageCache(8*4*8 + int64(len(server.URL+"/a"))) // one entry including its source key
	load := func(path string) {
		t.Helper()
		if _, err := cache.load(context.Background(), server.URL+path); err != nil {
			t.Fatal(err)
		}
	}
	load("/a")
	load("/a")
	if requests.Load() != 1 {
		t.Fatal("cached")
	}
	load("/b")
	load("/a")
	if requests.Load() != 3 || cache.used > cache.limit || len(cache.items) != 1 {
		t.Fatal("LRU budget")
	}
	cache.Delete(server.URL + "/a")
	load("/a")
	cache.Clear()
	load("/a")
	if requests.Load() != 5 {
		t.Fatal("invalidation")
	}
	for range 2 {
		if _, err := cache.load(context.Background(), server.URL+"/error"); err == nil {
			t.Fatal("error")
		}
	}
	if requests.Load() != 7 {
		t.Fatal("failure cached")
	}
}

func TestImageCacheClearRejectsStalePopulation(t *testing.T) {
	old := sourceTestPNG(t)
	var fresh bytes.Buffer
	png.Encode(&fresh, image.NewNRGBA(image.Rect(0, 0, 3, 3)))
	started, release := make(chan struct{}), make(chan struct{})
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			close(started)
			<-release
			w.Write(old)
		} else {
			w.Write(fresh.Bytes())
		}
	}))
	defer server.Close()
	cache := NewImageCache(1024)
	done := make(chan struct{})
	go func() { cache.load(context.Background(), server.URL); close(done) }()
	<-started
	cache.Clear()
	img, err := cache.load(context.Background(), server.URL)
	close(release)
	<-done
	if err != nil || img.still.Bounds().Dx() != 3 {
		t.Fatal("new generation", err)
	}
	img, err = cache.load(context.Background(), server.URL)
	if err != nil || img.still.Bounds().Dx() != 3 || requests.Load() != 2 {
		t.Fatal("stale population", err, requests.Load())
	}
}
