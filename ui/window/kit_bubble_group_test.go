package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestBubbleGroupSemanticsAndActions(t *testing.T) {
	calls := 0
	first := kit.Bubble(kit.Button("Accept", func() { calls++ }))
	second := kit.Bubble(el.ViewFunc(func(*el.Context) el.Element { return el.Text("Second message") })).Mine()
	group := kit.BubbleGroup(first, second).Name("Messages")
	w := openTest(t, Options{Width: 300, Height: 240, Content: el.Root(group)})
	if e := element(t, w, "Messages"); e.Role != "group" {
		t.Fatal("group semantics", e)
	}
	w.click(element(t, w, "Accept").center())
	if calls != 1 {
		t.Fatal("nested action unavailable")
	}
	group.SetItems(second, first)
	if err := w.press("Space"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatal("reorder lost focus")
	}
	group.SetDisabled(true)
	w.click(element(t, w, "Accept").center())
	if calls != 2 {
		t.Fatal("disabled group activated")
	}
	group.SetItems(second)
	for _, e := range w.snapshot() {
		if e.Name == "Accept" {
			t.Fatal("removed item remains in agent")
		}
	}
}
