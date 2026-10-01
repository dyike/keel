package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitEmptySnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Empty("暂无订单", "创建订单后将在这里显示"))})
	if e := element(t, w, "暂无订单"); e.Role != "empty" {
		t.Fatalf("invalid empty role: %+v", e)
	}
	if e := element(t, w, "创建订单后将在这里显示"); e.Role != "text" {
		t.Fatal("missing description")
	}
}
