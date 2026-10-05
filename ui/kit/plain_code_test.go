package kit

import (
	"testing"

	"github.com/dyike/keel/ui/core"
)

// Without a highlighter the editor shows plain text and still works.
func TestCodeEditorWithoutHighlighter(t *testing.T) {
	h := core.CurrentHighlighter()
	core.SetHighlighter(nil)
	defer core.SetHighlighter(h)
	if spans := highlightCode("go", "package main // hi", codeStyleName()); spans != nil {
		t.Fatal("highlighted without a highlighter")
	}
	if knownLanguage("go") {
		t.Fatal("no language is known without a highlighter")
	}
	ed := CodeEditor("package main\n\nfunc main() {}\n").Language("go")
	hr := editorHarness(ed)
	hr.Frame()
	if ed.Value() == "" {
		t.Fatal("editor lost its text")
	}
}
