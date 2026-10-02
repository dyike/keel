package kit

import (
	"strings"
	"testing"
)

func edit1(b *codeBuffer, from, to codePos, text string) codePos {
	b.begin(nil)
	end := b.edit(from, to, text)
	b.commit(nil, false)
	return end
}

func TestCodeBufferReplaceUndoRedo(t *testing.T) {
	b := newCodeBuffer("func main() {\n\tprintln(\"hi\")\n}")
	end := edit1(b, codePos{1, 1}, codePos{1, 8}, "fmt.Println")
	if got := b.text(); got != "func main() {\n\tfmt.Println(\"hi\")\n}" || end != (codePos{1, 12}) {
		t.Fatalf("replace in line: %q %v", got, end)
	}
	end = edit1(b, codePos{2, 0}, codePos{2, 0}, "\t// done\n")
	if got := b.text(); got != "func main() {\n\tfmt.Println(\"hi\")\n\t// done\n}" || end != (codePos{3, 0}) {
		t.Fatalf("insert lines: %q %v", got, end)
	}
	if b.slice(codePos{1, 5}, codePos{2, 3}) != "Println(\"hi\")\n\t//" {
		t.Fatalf("slice %q", b.slice(codePos{1, 5}, codePos{2, 3}))
	}
	end = edit1(b, codePos{0, 12}, codePos{3, 0}, "")
	if got := b.text(); got != "func main() }" || end != (codePos{0, 12}) {
		t.Fatalf("join lines: %q", got)
	}
	for range 3 {
		b.undoStep()
	}
	if b.text() != "func main() {\n\tprintln(\"hi\")\n}" {
		t.Fatalf("undo: %q", b.text())
	}
	b.redoStep()
	b.redoStep()
	if b.text() != "func main() {\n\tfmt.Println(\"hi\")\n\t// done\n}" {
		t.Fatalf("redo: %q", b.text())
	}
}

func TestCodeBufferStepUndoesEveryEdit(t *testing.T) {
	b := newCodeBuffer("a\nb\nc")
	b.begin(nil)
	b.edit(codePos{0, 1}, codePos{0, 1}, "1")
	b.edit(codePos{2, 1}, codePos{2, 1}, "3\nx")
	b.commit(nil, false)
	if b.text() != "a1\nb\nc3\nx" {
		t.Fatal(b.text())
	}
	b.undoStep()
	if b.text() != "a\nb\nc" {
		t.Fatalf("one undo for the step: %q", b.text())
	}
	b.redoStep()
	if b.text() != "a1\nb\nc3\nx" {
		t.Fatalf("redo the step: %q", b.text())
	}
}

func TestCodeBufferTypingMergesUntilSpace(t *testing.T) {
	b := newCodeBuffer("")
	p := codePos{}
	for _, r := range "go test" {
		b.begin(nil)
		p = b.edit(p, p, string(r))
		b.commit(nil, true)
	}
	if b.text() != "go test" || len(b.undo) != 2 { // "go", " test": a space starts a new step
		t.Fatalf("%q undo steps %d", b.text(), len(b.undo))
	}
	b.undoStep()
	if b.text() != "go" {
		t.Fatalf("undo a word: %q", b.text())
	}
}

func TestCodeBufferLargeFileChunks(t *testing.T) {
	var sb strings.Builder
	for i := range 200000 {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("x := 1 // line")
	}
	b := newCodeBuffer(sb.String())
	if b.count() != 200000 || len(b.lines.chunks) != 200000/codeChunk+1 {
		t.Fatal(b.count(), len(b.lines.chunks))
	}
	edit1(b, codePos{100000, 0}, codePos{100000, 0}, "a\nb\n")
	if b.count() != 200002 || string(b.line(100001)) != "b" || string(b.line(100002)) != "x := 1 // line" {
		t.Fatal("insert in the middle of a large file")
	}
	// Delete across many chunks, then everything.
	edit1(b, codePos{10, 0}, codePos{150000, 0}, "")
	if b.count() != 200002-149990 || string(b.line(10)) != "x := 1 // line" {
		t.Fatal("delete across chunks", b.count())
	}
	edit1(b, codePos{}, codePos{b.count() - 1, 14}, "")
	if b.count() != 1 || b.text() != "" {
		t.Fatal("delete all", b.count())
	}
	b.undoStep()
	if b.count() != 200002-149990 {
		t.Fatal("undo delete all")
	}
	if s, e := b.wordAt(codePos{11, 1}); s.col != 0 || e.col != 1 {
		t.Fatal("word at")
	}
}
