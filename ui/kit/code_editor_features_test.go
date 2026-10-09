package kit

import (
	"io"
	"strings"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"github.com/dyike/keel/ui/internal/uitest"
)

// at is the screen point inside a text position's cell.
func at(ed *CodeEditorView, line, col int) f32.Point {
	p := ed.caretPoint(codePos{line, col})
	return f32.Pt(float32(p.X)+1, float32(p.Y+ed.metrics.lh/2))
}

func press(h *uitest.Harness, p f32.Point, mods key.Modifiers) {
	h.Router.Queue(
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p, Modifiers: mods},
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p, Modifiers: mods},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p, Modifiers: mods},
	)
	h.Frame()
}

// dragFrom presses at a, moves with the button held to b and releases.
func dragFrom(h *uitest.Harness, a, b f32.Point, mods key.Modifiers) {
	h.Router.Queue(
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: a, Modifiers: mods},
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: a, Modifiers: mods},
	)
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: b, Modifiers: mods})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: b, Modifiers: mods})
	h.Frame()
}

func focusedEditor(text string) (*CodeEditorView, *uitest.Harness) {
	ed := CodeEditor(text)
	h := editorHarness(ed)
	ed.Focus()
	h.Frame()
	h.Frame()
	return ed, h
}

func TestCodeEditorMultipleCarets(t *testing.T) {
	ed, h := focusedEditor("alpha\nbeta\ngamma")
	ed.SetCursor(0, 0)
	h.Frame()
	h.Key(key.NameDownArrow, key.ModShortcut|key.ModAlt)
	h.Key(key.NameDownArrow, key.ModShortcut|key.ModAlt)
	if ed.Cursors() != 3 {
		t.Fatalf("carets %d", ed.Cursors())
	}
	typeAt(h, "> ")
	if ed.Value() != "> alpha\n> beta\n> gamma" {
		t.Fatalf("typing at every caret: %q", ed.Value())
	}
	h.Key(key.NameDeleteBackward, 0)
	if ed.Value() != ">alpha\n>beta\n>gamma" {
		t.Fatalf("delete at every caret: %q", ed.Value())
	}
	h.Key("Z", key.ModShortcut)
	if ed.Value() != "> alpha\n> beta\n> gamma" {
		t.Fatalf("one undo for a multi-caret edit: %q", ed.Value())
	}
	h.Key(key.NameEscape, 0)
	if ed.Cursors() != 1 {
		t.Fatal("Esc keeps one caret")
	}
	// Alt+click adds a caret; a plain click goes back to one.
	press(h, at(ed, 0, 4), key.ModAlt)
	press(h, at(ed, 1, 4), key.ModAlt)
	if ed.Cursors() != 3 {
		t.Fatalf("alt+click carets %d", ed.Cursors())
	}
	press(h, at(ed, 0, 0), 0)
	if ed.Cursors() != 1 {
		t.Fatal("click keeps one caret")
	}
}

func TestCodeEditorColumnSelectionAndNextOccurrence(t *testing.T) {
	ed, h := focusedEditor("let a = 1\nlet b = 2\nlet c = 3")
	dragFrom(h, at(ed, 0, 4), at(ed, 2, 5), key.ModAlt|key.ModShift)
	if ed.Cursors() != 3 {
		t.Fatalf("block rows %d", ed.Cursors())
	}
	typeAt(h, "x")
	if ed.Value() != "let x = 1\nlet x = 2\nlet x = 3" {
		t.Fatalf("typing over a block: %q", ed.Value())
	}
	ed.SetCursor(0, 1)
	h.Frame()
	h.Key("D", key.ModShortcut) // select the word
	h.Key("D", key.ModShortcut) // and its next occurrence
	h.Key("D", key.ModShortcut)
	if ed.Cursors() != 3 || ed.Selection() != "let" {
		t.Fatalf("next occurrence: %d %q", ed.Cursors(), ed.Selection())
	}
	typeAt(h, "var")
	if ed.Value() != "var x = 1\nvar x = 2\nvar x = 3" {
		t.Fatalf("editing occurrences: %q", ed.Value())
	}
	// Pasting as many lines as carets gives one line each.
	h.Key(key.NameHome, key.ModShift)
	h.Key("V", key.ModShortcut)
	h.Router.Queue(transfer.DataEvent{Type: "application/text", Open: func() io.ReadCloser { return io.NopCloser(strings.NewReader("a\nb\nc")) }})
	h.Frame()
	if !strings.HasPrefix(ed.Value(), "a x = 1\nb x") {
		t.Fatalf("paste distributes lines: %q", ed.Value())
	}
}

func TestCodeEditorBracketPairs(t *testing.T) {
	ed, h := focusedEditor("")
	typeAt(h, "f")
	typeAt(h, "(")
	if ed.Value() != "f()" {
		t.Fatalf("pair: %q", ed.Value())
	}
	typeAt(h, "x")
	typeAt(h, ")")
	if ed.Value() != "f(x)" || func() int { _, c := ed.Cursor(); return c }() != 4 {
		t.Fatalf("step over the closer: %q", ed.Value())
	}
	typeAt(h, " ")
	typeAt(h, "{")
	h.Key(key.NameReturn, 0)
	if ed.Value() != "f(x) {\n    \n}" {
		t.Fatalf("enter between a pair: %q", ed.Value())
	}
	h.Key(key.NameDeleteBackward, key.ModShortcut) // clear the indent
	typeAt(h, "[")
	h.Key(key.NameDeleteBackward, 0)
	if ed.Value() != "f(x) {\n\n}" {
		t.Fatalf("backspace deletes an empty pair: %q", ed.Value())
	}
	typeAt(h, "don't")
	if !strings.Contains(ed.Value(), "don't\n") {
		t.Fatalf("apostrophe is not a quote pair: %q", ed.Value())
	}
	// Wrap a selection.
	h.Key(key.NameLeftArrow, key.ModShift|key.ModShortcut)
	typeAt(h, "\"")
	if !strings.Contains(ed.Value(), "\"don't\"") || ed.Selection() != "don't" {
		t.Fatalf("wrap selection: %q %q", ed.Value(), ed.Selection())
	}
	// A closer typed on a blank line dedents to its opener.
	ed2, h2 := focusedEditor("if x {\n        ")
	ed2.SetCursor(1, 8)
	h2.Frame()
	typeAt(h2, "}")
	if ed2.Value() != "if x {\n}" {
		t.Fatalf("dedent closer: %q", ed2.Value())
	}
	// No pairing with AutoClose off.
	ed3, h3 := focusedEditor("")
	ed3.AutoClose(false)
	typeAt(h3, "(")
	if ed3.Value() != "(" {
		t.Fatalf("auto close off: %q", ed3.Value())
	}
}

func TestCodeEditorPairsSkipStringsAndComments(t *testing.T) {
	ed, h := focusedEditor("x := \"\" // note")
	ed.Language("go")
	ed.buf.highlightNear("go", codeStyleName(), 0, 0)
	ed.SetCursor(0, 6) // inside the string
	h.Frame()
	typeAt(h, "(")
	if ed.Value() != "x := \"(\" // note" {
		t.Fatalf("no pair in a string: %q", ed.Value())
	}
}

func TestCodeEditorFindReplace(t *testing.T) {
	ed, h := focusedEditor("foo Foo food\nfoo(bar)")
	h.Key("F", key.ModShortcut)
	h.Frame() // the panel takes focus into its field
	if !ed.search.open {
		t.Fatal("Cmd/Ctrl+F opens find")
	}
	ed.search.query = "foo"
	if ed.SearchMatches() != 4 { // case-insensitive, inside words too
		t.Fatalf("matches %d", ed.SearchMatches())
	}
	ed.search.wholeWord = true
	if ed.SearchMatches() != 3 {
		t.Fatalf("whole word matches %d", ed.SearchMatches())
	}
	ed.search.matchCase = true
	if ed.SearchMatches() != 2 {
		t.Fatalf("case matches %d", ed.SearchMatches())
	}
	ed.findNext(1)
	if ed.Selection() != "foo" {
		t.Fatalf("find next selects: %q", ed.Selection())
	}
	ed.search.replacement = "baz"
	ed.replaceAll()
	if ed.Value() != "baz Foo food\nbaz(bar)" {
		t.Fatalf("replace all: %q", ed.Value())
	}
	ed.Focus() // the find field had focus; Cmd+Z there undoes the query
	h.Frame()
	h.Key("Z", key.ModShortcut)
	if ed.Value() != "foo Foo food\nfoo(bar)" {
		t.Fatalf("replace all is one undo step: %q", ed.Value())
	}
	ed.search.regex, ed.search.wholeWord = true, false
	ed.search.query, ed.search.replacement = `(\w+)\((\w+)\)`, "$2.$1()"
	ed.replaceAll()
	if ed.Value() != "foo Foo food\nbar.foo()" {
		t.Fatalf("regex replace: %q", ed.Value())
	}
	ed.search.query = "("
	if ed.SearchMatches() != 0 || !ed.search.badRegex {
		t.Fatal("invalid regex reported")
	}
	ed.CloseSearch()
	if ed.search.open {
		t.Fatal("close")
	}
	ed.SetReadOnly(true)
	ed.OpenSearch(true)
	if ed.search.replace {
		t.Fatal("read-only editors only find")
	}
}

func TestCodeEditorFolding(t *testing.T) {
	src := "func a() {\n\tx := 1\n\tif x {\n\t\ty()\n\t}\n}\nfunc b() {}"
	ed, h := focusedEditor(src)
	if ed.foldEnd(0) != 4 || ed.foldEnd(2) != 3 || ed.foldEnd(6) != 6 {
		t.Fatalf("regions %d %d %d", ed.foldEnd(0), ed.foldEnd(2), ed.foldEnd(6))
	}
	ed.Fold(0)
	h.Frame()
	if ed.rowCount() != 3 || ed.lineOf(1) != 5 || ed.rowOf(3) != 0 {
		t.Fatalf("rows %d %d %d", ed.rowCount(), ed.lineOf(1), ed.rowOf(3))
	}
	// Moving down from the header skips the folded lines.
	ed.SetCursor(0, 0)
	ed.Fold(0)
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	if l, _ := ed.Cursor(); l != 5 {
		t.Fatalf("down over a fold to line %d", l)
	}
	// A caret moved inside opens the fold.
	ed.SetCursor(3, 0)
	if ed.Folded(0) {
		t.Fatal("a caret inside a fold opens it")
	}
	// Editing above shifts folds; the gutter arrow toggles them.
	ed.SetCursor(0, 0)
	ed.Fold(2)
	h.Frame()
	h.Key(key.NameReturn, 0)
	if !ed.Folded(3) || ed.Folded(2) {
		t.Fatal("fold moved with the edit")
	}
	press(h, f32.Pt(4, float32(ed.rowOf(3)*ed.metrics.lh+ed.metrics.lh/2)), 0)
	if ed.Folded(3) {
		t.Fatal("arrow click opens the fold")
	}
	ed.FoldAll()
	if !ed.Folded(1) || ed.rowCount() != 4 {
		t.Fatalf("fold all: rows %d", ed.rowCount())
	}
	ed.UnfoldAll()
	if ed.rowCount() != ed.Lines() {
		t.Fatal("unfold all")
	}
}

func TestCodeEditorDefinitionWhitespaceAndTabs(t *testing.T) {
	var got [2]int
	ed, h := focusedEditor("call greet()")
	ed.OnDefinition(func(line, col int) { got = [2]int{line, col} })
	press(h, at(ed, 0, 8), key.ModShortcut)
	if got != [2]int{0, 5} {
		t.Fatalf("cmd+click looks up the word start: %v", got)
	}
	ed.SetCursor(0, 2)
	h.Frame()
	h.Key(key.NameF12, 0)
	if got != [2]int{0, 0} {
		t.Fatalf("F12: %v", got)
	}
	ed.ShowWhitespace(true).TabSize(2, false)
	ed.SetValue("\tx  y")
	ed.SetCursor(0, 5)
	h.Frame()
	h.Key(key.NameTab, 0)
	if ed.Value() != "\tx  y  " {
		t.Fatalf("tab size 2 with spaces: %q", ed.Value())
	}
	if x := ed.colX(0); x[1] != 2*ed.metrics.space {
		t.Fatalf("tab stop is 2 spaces wide: %v", x)
	}
}
