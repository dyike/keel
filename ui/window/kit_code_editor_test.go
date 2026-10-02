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

func TestKitCodeEditorCompletion(t *testing.T) {
	for _, accept := range []string{"enter", "click"} {
		t.Run(accept, func(t *testing.T) {
			ed := kit.CodeEditor("").Name("source").OnComplete(func(_, _ int, prefix string) []kit.CodeCompletion {
				if prefix == "gr" {
					return []kit.CodeCompletion{{Label: "greet", Insert: "greet()"}}
				}
				return nil
			})
			w := openTest(t, Options{Content: views(ed)})
			w.click(element(t, w, "source").center())
			if err := w.typeText("gr"); err != nil {
				t.Fatal(err)
			}
			item := element(t, w, "greet")
			if item.Role != "option" || item.Selected == nil || !*item.Selected {
				t.Fatalf("completion semantics: %+v", item)
			}
			if accept == "click" {
				w.click(item.center())
			} else {
				if err := w.press("enter"); err != nil {
					t.Fatal(err)
				}
			}
			w.snapshot()
			if ed.Value() != "greet()" {
				t.Fatalf("completion not accepted: %q", ed.Value())
			}
			for _, e := range w.snapshot() {
				if e.Role == "option" {
					t.Fatal("completion stayed open")
				}
			}
		})
	}
}

// The find panel's field and buttons are listed for agents, who search and
// replace through them.
func TestKitCodeEditorFindPanel(t *testing.T) {
	ed := kit.CodeEditor("foo bar foo").Name("source")
	w := openTest(t, Options{Content: views(ed)})
	ed.OpenSearch(true)
	w.snapshot()
	field := element(t, w, "查找") // the field wins over the panel of the same name
	if field.Role != "textbox" {
		t.Fatalf("find field: %+v", field)
	}
	w.click(field.center())
	if err := w.typeText("foo"); err != nil {
		t.Fatal(err)
	}
	w.click(element(t, w, "全部替换").center())
	w.snapshot()
	if ed.Value() != " bar " {
		t.Fatalf("replace all from the panel: %q", ed.Value())
	}
}
