package kit

import (
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func bigSource(lines int) string {
	var sb strings.Builder
	for i := range lines {
		fmt.Fprintf(&sb, "x%d := %d // 第 %d 行\n", i, i*7, i+1)
	}
	return sb.String()
}

func editorHarness(ed *CodeEditorView) *uitest.Harness {
	root := el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().Child(ed.Fill().Render(cx)) }))
	return uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(800, 600); root.Layout(gtx) })
}

// typeAt types like a platform input method or the automation server: at the
// selection the editor last reported.
func typeAt(h *uitest.Harness, s string) {
	h.Router.Queue(key.EditEvent{Range: h.Router.EditorState().Selection.Range, Text: s})
	h.Frame()
}

func TestCodeEditorEditingKeys(t *testing.T) {
	var changes int
	ed := CodeEditor("func f() {\n\treturn\n}").OnChange(func(string) { changes++ })
	h := editorHarness(ed)
	ed.Focus()
	h.Frame()
	ed.SetCursor(0, 10) // after {
	h.Frame()
	h.Key(key.NameReturn, 0)
	typeAt(h, "x")
	if got := ed.Value(); got != "func f() {\n\tx\n\treturn\n}" {
		t.Fatalf("enter keeps indent and adds one after {: %q", got)
	}
	h.Key(key.NameHome, 0)
	h.Key(key.NameHome, key.ModShift) // to column 0, selecting the indent
	if ed.Selection() != "\t" {
		t.Fatalf("home toggles to column 0: %q", ed.Selection())
	}
	h.Key("Z", key.ModShortcut)
	h.Key("Z", key.ModShortcut)
	if ed.Value() != "func f() {\n\treturn\n}" || changes == 0 {
		t.Fatalf("undo: %q", ed.Value())
	}
	h.Key("Z", key.ModShortcut|key.ModShift)
	if !strings.Contains(ed.Value(), "{\n\t\n") {
		t.Fatalf("redo: %q", ed.Value())
	}
	ed.SetCursor(1, 0)
	h.Frame()
	h.Key(key.NameDownArrow, key.ModShift)
	h.Key(key.NameTab, 0)
	if lines := strings.Split(ed.Value(), "\n"); lines[1] != "\t\t" || lines[2] != "\treturn" {
		t.Fatalf("tab indents selected lines: %q", ed.Value())
	}
	h.Key(key.NameEscape, 0)
	h.Key(key.NameTab, 0) // after Esc, Tab leaves instead of indenting
	if strings.Count(ed.Value(), "\t") != 3 {
		t.Fatalf("tab after esc edited: %q", ed.Value())
	}
}

func TestCodeEditorLargeFileFrames(t *testing.T) {
	ed := CodeEditor(bigSource(200000)).Language("go")
	h := editorHarness(ed)
	ed.Focus()
	h.Frame()
	start := time.Now()
	ed.SetCursor(150000, 3)
	h.Frame()
	typeAt(h, "abc")
	h.Key(key.NamePageDown, 0)
	h.Frame()
	per := time.Since(start) / 4
	if l, _ := ed.Cursor(); l < 150000 || !strings.Contains(strings.Split(ed.Value(), "\n")[150000], "abc") {
		t.Fatalf("edit at line 150001: cursor line %d", l)
	}
	if per > 100*time.Millisecond {
		t.Fatalf("a frame in a 200,000-line file took %v", per)
	}
}

func BenchmarkCodeEditorLargeFileFrame(b *testing.B) {
	ed := CodeEditor(bigSource(200000)).Language("go")
	h := editorHarness(ed)
	h.Frame()
	for i := 0; b.Loop(); i++ {
		ed.SetCursor(i*997%200000, 0)
		h.Frame()
	}
}
