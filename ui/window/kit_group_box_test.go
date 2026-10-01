package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitGroupBoxSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.GroupBox("账户").Child(kit.Tag("已验证")))})
	if e := element(t, w, "账户"); e.Role != "group" {
		t.Fatalf("invalid group: %+v", e)
	}
	if e := element(t, w, "已验证"); e.Role != "tag" {
		t.Fatal("group absorbed child")
	}
}
