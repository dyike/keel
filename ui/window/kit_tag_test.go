package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitTagSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Tag("已完成").Tone(kit.Success))})
	if e := element(t, w, "已完成"); e.Role != "tag" || e.Value != "success" {
		t.Fatalf("invalid tag: %+v", e)
	}
}
