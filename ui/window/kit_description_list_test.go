package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitDescriptionListSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.DescriptionList().Item("客户", "张三"))})
	e := element(t, w, "客户：张三")
	if e.Role != "text" {
		t.Fatal(e)
	}
	for _, e := range w.snapshot() {
		if e.Name == "客户" || e.Name == "张三" {
			t.Fatal("field duplicated")
		}
	}
}
