package kit

import (
	"strings"
	"testing"
)

func TestCodeBufferReplaceUndoRedo(t *testing.T) {
	b := newCodeBuffer("func main() {\n\tprintln(\"hi\")\n}")
	end := b.edit(codePos{1, 1}, codePos{1, 8}, "fmt.Println", [2]codePos{}, false)
	if got := b.text(); got != "func main() {\n\tfmt.Println(\"hi\")\n}" || end != (codePos{1, 12}) {
		t.Fatalf("replace in line: %q %v", got, end)
	}
	end = b.edit(codePos{2, 0}, codePos{2, 0}, "\t// done\n", [2]codePos{}, false)
	if got := b.text(); got != "func main() {\n\tfmt.Println(\"hi\")\n\t// done\n}" || end != (codePos{3, 0}) {
		t.Fatalf("insert lines: %q %v", got, end)
	}
	if b.slice(codePos{1, 5}, codePos{2, 3}) != "Println(\"hi\")\n\t//" {
		t.Fatalf("slice %q", b.slice(codePos{1, 5}, codePos{2, 3}))
	}
	end = b.edit(codePos{0, 12}, codePos{3, 0}, "", [2]codePos{}, false)
	if got := b.text(); got != "func main() }" || end != (codePos{0, 12}) {
		t.Fatalf("join lines: %q", got)
	}
	for range 3 {
		b.undoOne()
	}
	if b.text() != "func main() {\n\tprintln(\"hi\")\n}" {
		t.Fatalf("undo: %q", b.text())
	}
	b.redoOne()
	b.redoOne()
	if b.text() != "func main() {\n\tfmt.Println(\"hi\")\n\t// done\n}" {
		t.Fatalf("redo: %q", b.text())
	}
}

func TestCodeBufferTypingMergesUntilSpace(t *testing.T) {
	b := newCodeBuffer("")
	p := codePos{}
	for _, r := range "go test" {
		p = b.edit(p, p, string(r), [2]codePos{p, p}, true)
	}
	if b.text() != "go test" || len(b.undo) != 2 { // "go", " test": a space starts a new step
		t.Fatalf("%q undo steps %d", b.text(), len(b.undo))
	}
	b.undoOne()
	if b.text() != "go" {
		t.Fatalf("undo a word: %q", b.text())
	}
}

func TestCodeBufferLargeFile(t *testing.T) {
	var sb strings.Builder
	for i := range 200000 {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("x := 1 // line")
	}
	b := newCodeBuffer(sb.String())
	if len(b.lines) != 200000 {
		t.Fatal(len(b.lines))
	}
	b.edit(codePos{100000, 0}, codePos{100000, 0}, "a\nb\n", [2]codePos{}, false)
	if len(b.lines) != 200002 || string(b.lines[100001]) != "b" || string(b.lines[100002]) != "x := 1 // line" {
		t.Fatal("insert in the middle of a large file")
	}
	if s, e := b.wordAt(codePos{100002, 1}); s.col != 0 || e.col != 1 {
		t.Fatal("word at")
	}
}
