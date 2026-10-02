package window

import (
	"strings"
	"testing"

	"github.com/dyike/keel/ui/kit"
)

func TestKitCodeEditorAgentTypes(t *testing.T) {
	ed := kit.CodeEditor("package main\n").Language("go").Name("main.go")
	w := openTest(t, Options{Content: views(ed)})
	e := element(t, w, "main.go")
	if e.Role != "textbox" || !strings.HasPrefix(e.Value, "package main") {
		t.Fatalf("editor semantics: %+v", e)
	}
	w.click(e.center())
	if err := w.typeText("x"); err != nil {
		t.Fatal(err)
	}
	w.snapshot()
	if !strings.Contains(ed.Value(), "x") {
		t.Fatalf("typed text missing: %q", ed.Value())
	}
}
