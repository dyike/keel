package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitDescriptionListSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.DescriptionList(kit.Description{Label: "客户", Text: "张三"}))})
	found := false
	for _, e := range w.snapshot() {
		if e.Role == "descriptionlist" && e.Value == "1" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing list")
	}
	element(t, w, "客户")
	element(t, w, "张三")
}
