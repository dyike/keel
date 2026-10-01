package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitTagSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Tag("已完成").Tone(kit.ToneSuccess))})
	if e := element(t, w, "已完成"); e.Role != "tag" || e.Value != "success" {
		t.Fatalf("invalid tag: %+v", e)
	}
}

func TestKitTagKeyboardRemovalAndSelection(t *testing.T) {
	removed := 0
	v := kit.Tag("待办").Selectable().OnRemove(func() { removed++ })
	w := openTest(t, Options{Content: el.Embed(v)})
	for _, s := range []string{"tab", "space", "tab", "backspace"} {
		if err := w.press(s); err != nil {
			t.Fatal(err)
		}
	}
	if !v.Value() || removed != 1 {
		t.Fatal("keyboard tag interaction failed")
	}
	e := element(t, w, "待办")
	if e.Selected == nil || !*e.Selected {
		t.Fatal("missing selected semantics")
	}
}
