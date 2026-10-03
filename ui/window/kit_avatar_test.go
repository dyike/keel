package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitAvatarSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Avatar("Ada Lovelace").Status(kit.AvatarOnline))})
	if e := element(t, w, "Ada Lovelace"); e.Role != "avatar" || e.Value != "online" {
		t.Fatalf("invalid avatar: %+v", e)
	}
}

func TestKitAvatarGroupSnapshot(t *testing.T) {
	group := kit.AvatarGroup(kit.Avatar("Ada").Status(kit.AvatarOnline), kit.Avatar("Bob"), kit.Avatar("Carol")).Limit(2)
	w := openTest(t, Options{Width: 320, Height: 200, Content: el.Embed(group)})
	if e := element(t, w, "Ada"); e.Role != "avatar" || e.Value != "online" {
		t.Fatalf("member: %+v", e)
	}
	if e := element(t, w, "更多 1"); e.Role != "avatar" || e.Value != "1" {
		t.Fatalf("overflow: %+v", e)
	}
	for _, e := range w.snapshot() {
		if e.Name == "Carol" {
			t.Fatal("hidden member exposed")
		}
	}
	group.Ellipsis(true)
	if e := element(t, w, "更多 1"); e.Value != "1" {
		t.Fatal("ellipsis lost count")
	}
}
