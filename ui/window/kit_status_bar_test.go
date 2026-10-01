package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitStatusBarSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.StatusBar().Left(el.Text("连接正常")).Right(el.Text("3 个任务")))})
	element(t, w, "连接正常")
	element(t, w, "3 个任务")
	found := false
	for _, e := range w.snapshot() {
		if e.Role == "status" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing status container")
	}
}
