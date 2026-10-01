package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitAvatarSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Avatar("Ada Lovelace"))})
	if e := element(t, w, "Ada Lovelace"); e.Role != "image" || e.Value != "initials" {
		t.Fatalf("invalid avatar: %+v", e)
	}
}
