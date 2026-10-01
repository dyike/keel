package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitMarkerHasNoSemantics(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Marker(kit.MarkerDiamond))})
	for _, e := range w.snapshot() {
		if e.Role != "" {
			t.Fatalf("decorative marker exposed: %+v", e)
		}
	}
}
