package kit

import (
	"strings"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
)

func TestCodeEditorSoftWrap(t *testing.T) {
	long := strings.Repeat("word ", 60) // far wider than 800px
	ed := CodeEditor("short\n" + long + "\nend").SoftWrap(true)
	h := editorHarness(ed)
	ed.Focus()
	h.Frame()
	rows := ed.vrowCount()
	if rows < 5 {
		t.Fatalf("long line should wrap over several rows, got %d rows", rows)
	}
	width := ed.wrapWidth()
	for r := 0; r < rows; r++ {
		vr := ed.vrow(r)
		xs := ed.colX(vr.line)
		if xs[vr.end]-xs[vr.start] > width {
			t.Fatalf("row %d is %dpx, wider than %d", r, xs[vr.end]-xs[vr.start], width)
		}
		if vr.line == 1 && vr.end < len(xs)-1 && ed.buf.line(1)[vr.end-1] != ' ' {
			t.Fatalf("row %d breaks mid-word", r)
		}
	}
	// Down walks the rows of the wrapped line before reaching the next line.
	ed.SetCursor(1, 2)
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	// The same x on the next row lands on the same offset in a uniform line.
	if line, col := ed.Cursor(); line != 1 || col != ed.vrow(2).start+2 {
		t.Fatalf("down should keep x within the wrapped line: %d:%d, row starts at %d", line, col, ed.vrow(2).start)
	}
	for i := 0; i < rows; i++ {
		h.Key(key.NameDownArrow, 0)
	}
	if line, _ := ed.Cursor(); line != 2 {
		t.Fatal("down reaches the last line", line)
	}
	// A click on the second visual row maps into the middle of line 1.
	p := ed.posAt(f32.Pt(float32(ed.metrics.textOrigin+ed.metrics.space), float32(2*ed.metrics.lh+ed.metrics.lh/2)))
	if p.line != 1 || p.col != ed.vrow(2).start+1 {
		t.Fatalf("click on a wrapped row: %+v, row starts %d", p, ed.vrow(2).start)
	}
	// A caret at a break shows at the start of the next row.
	if pt := ed.caretPoint(codePos{1, ed.vrow(2).start}); pt.X != ed.metrics.textOrigin || pt.Y != 2*ed.metrics.lh {
		t.Fatalf("caret at a break: %v", pt)
	}
	// Folding and editing rebuild the rows; turning wrap off restores one row per line.
	typeAt(h, strings.Repeat("x", 300))
	if ed.vrowCount() <= rows {
		t.Fatal("typing a long run should add rows", ed.vrowCount(), rows)
	}
	ed.SoftWrap(false)
	h.Frame()
	if ed.vrowCount() != 3 {
		t.Fatal("without wrap there is a row per line", ed.vrowCount())
	}
}

func TestCodeDecorationFrameSpansRows(t *testing.T) {
	ed := CodeEditor("alpha\n\nbeta gamma\n")
	h := editorHarness(ed)
	h.Frame()
	from, to := codePos{0, 2}, codePos{2, 4}
	if _, _, ok := ed.decorationSpan(from, to, CodeDecorationFrame, 1); !ok {
		t.Fatal("an empty line inside a frame joins it")
	}
	if _, _, ok := ed.decorationSpan(from, to, CodeDecorationUnderline, 1); ok {
		t.Fatal("an underline skips empty lines")
	}
	x0, x1, ok := ed.decorationSpan(from, to, CodeDecorationFrame, 0)
	xs := ed.colX(0)
	if !ok || x0 != xs[2] || x1 != xs[5]+ed.metrics.space {
		t.Fatal("first row runs to the line end plus a space", x0, x1)
	}
	ed.Decorations(CodeDecoration{Range: CodeRange{Line: 0, Col: 2, EndLine: 2, EndCol: 4}, Style: CodeDecorationFrame})
	h.Frame() // paints without panicking across rows
}

func TestCodeDecorationStyleOnlyKeepsSyntaxColor(t *testing.T) {
	ed := CodeEditor("func main() {}")
	ed.Decorations(CodeDecoration{Range: CodeRange{EndCol: 4}, Style: CodeDecorationText, Italic: true})
	ds := ed.lineDecorations(0)
	if len(ds) != 1 || !ds[0].keep || !ds[0].italic {
		t.Fatalf("italic-only decoration should keep the syntax color: %+v", ds)
	}
}
