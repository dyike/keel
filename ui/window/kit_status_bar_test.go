package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitStatusBarSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.StatusBar("连接正常", "3 个任务"))})
	if e := element(t, w, "连接正常"); e.Role != "status" || e.Value != "3 个任务" {
		t.Fatalf("invalid status: %+v", e)
	}
}
