package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitInputGroupSnapshot(t *testing.T) {
	input := kit.Input("")
	input.SetValue("hello")
	group := kit.InputGroup("Query", input).Suffix(kit.Button("Search", func() {}))
	w := openTest(t, kitPage(group))
	found := false
	for _, e := range w.snapshot() {
		if e.Role == "textbox" && e.Name == "Query" && e.Value == "hello" {
			found = true
		}
	}
	if !found {
		t.Fatal("group label is not associated with editor")
	}
	if element(t, w, "Search").Role != "button" {
		t.Fatal("suffix action missing")
	}
}

func TestKitInputGroupBlockSnapshot(t *testing.T) {
	input := kit.TextArea("").Rows(2)
	input.SetValue("first\nsecond")
	calls := 0
	group := kit.InputGroup("Message", input).
		Addon("heading", kit.InputGroupBlockStart, el.ViewFunc(func(*el.Context) el.Element { return el.Text("Draft") })).
		Addon("send", kit.InputGroupBlockEnd, kit.Button("Send", func() { calls++ }))
	w := openTest(t, kitPage(group))
	var editor Element
	for _, e := range w.snapshot() {
		if e.Role == "textbox" && e.Name == "Message" {
			editor = e
		}
	}
	if editor.Value != "first\nsecond" || element(t, w, "Draft").Y >= editor.Y || element(t, w, "Send").Y <= editor.Y {
		t.Fatalf("composer layout: %+v", editor)
	}
	w.click(element(t, w, "Send").center())
	if calls != 1 {
		t.Fatal("send action missing")
	}
	group.SetDisabled(true)
	if !element(t, w, "Send").Disabled {
		t.Fatal("addon semantics not disabled")
	}
	group.Addon("send", kit.InputGroupBlockEnd, nil)
	for _, e := range w.snapshot() {
		if e.Name == "Send" {
			t.Fatal("removed addon remains")
		}
	}
}
