package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitMarkerSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Marker("在线").Tone(kit.Success))})
	if e := element(t, w, "在线"); e.Role != "marker" || e.Value != "success" {
		t.Fatalf("invalid marker: %+v", e)
	}
}
