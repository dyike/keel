package window

import (
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
