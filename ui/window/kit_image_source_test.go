package window

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
)

func TestKitImageSourceAgentRetryAndSlots(t *testing.T) {
	var data bytes.Buffer
	png.Encode(&data, image.NewNRGBA(image.Rect(0, 0, 20, 10)))
	var failure atomic.Bool
	failure.Store(true)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failure.Load() {
			http.Error(w, "failed", 500)
			return
		}
		w.Write(data.Bytes())
	}))
	defer server.Close()
	v := kit.Image(nil, "Remote photo").Size(160, 80).LoadingContent(kit.Label("Fetching pixels")).Fallback(kit.Label("No photo available")).Source(server.URL)
	w := openTest(t, Options{Content: views(v)})
	wait := func(state string) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if e := element(t, w, "Remote photo"); e.Value == state {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("state timeout", state)
	}
	wait("error")
	core.Update(func() { v.SetDisabled(true) })
	w.click(element(t, w, locale.Current().Name(locale.Current().Retry, "Remote photo")).center())
	w.snapshot()
	if v.ImageError() == nil {
		t.Fatal("disabled retry")
	}
	failure.Store(false)
	core.Update(func() { v.SetDisabled(false) })
	w.click(element(t, w, locale.Current().Name(locale.Current().Retry, "Remote photo")).center())
	wait("loaded")
	if v.ImageError() != nil {
		t.Fatal(v.ImageError())
	}
	if _, err := w.screenshot(); err != nil {
		t.Fatal(err)
	}
	core.Update(func() { v.Source("") })
	wait("loading")
}
