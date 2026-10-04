package kit

import "testing"

func TestCodeSearchRegexSpansLines(t *testing.T) {
	ed := CodeEditor("func a() {\n\treturn 1\n}\nfunc b() {\n\treturn 2\n}\n你好\nworld")
	h := editorHarness(ed)
	h.Frame()
	ed.SetSearchQuery(`\{\n\treturn (\d)`, CodeSearchOptions{Regex: true})
	ms := ed.search.matches
	if len(ms) != 2 || ms[0] != (codeRange{codePos{0, 9}, codePos{1, 9}}) || ms[1] != (codeRange{codePos{3, 9}, codePos{4, 9}}) {
		t.Fatalf("multi-line matches: %+v", ms)
	}
	h.Frame() // paints highlights across rows
	ed.SetSearchQuery(`^func`, CodeSearchOptions{Regex: true})
	if n := ed.SearchMatches(); n != 2 {
		t.Fatal("^ anchors each line", n)
	}
	ed.SetSearchQuery(`.+\n.+`, CodeSearchOptions{Regex: true})
	if ms := ed.search.matches; len(ms) == 0 || ms[0].from.line != 0 || ms[0].to.line != 1 {
		t.Fatalf(". stops at a line break: %+v", ms)
	}
	ed.SetSearchQuery(`好\nw`, CodeSearchOptions{Regex: true})
	if ms := ed.search.matches; len(ms) != 1 || ms[0] != (codeRange{codePos{6, 1}, codePos{7, 1}}) {
		t.Fatalf("rune columns: %+v", ms)
	}
	ed.SetSearchQuery(`b`, CodeSearchOptions{Regex: true, WholeWord: true})
	if n := ed.SearchMatches(); n != 1 {
		t.Fatal("whole word over the document", n)
	}

	// Replacing expands groups in context; replace-all is one undo step.
	ed.SetSearchQuery(`\{\n\treturn (\d)\n\}`, CodeSearchOptions{Regex: true})
	if n := ed.ReplaceAllSearchMatches("{ return $1 }"); n != 2 {
		t.Fatal("replaced", n)
	}
	if got := ed.Value(); got != "func a() { return 1 }\nfunc b() { return 2 }\n你好\nworld" {
		t.Fatalf("after replace all: %q", got)
	}
	ed.SetValue("x {\n\treturn 7\n}")
	h.Frame()
	ed.SetSearchQuery(`\{\n\treturn (\d)\n\}`, CodeSearchOptions{Regex: true})
	if !ed.SelectSearchMatch(0) || !ed.ReplaceCurrentSearchMatch("=> $1") {
		t.Fatal("select and replace the match")
	}
	if got := ed.Value(); got != "x => 7" {
		t.Fatalf("replace one: %q", got)
	}
}
