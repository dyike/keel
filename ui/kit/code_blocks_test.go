package kit

import (
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/key"
	"strings"
	"testing"
)

func TestCodeEditorBlockGeometryAndKeyboard(t *testing.T) {
	ed := CodeEditor("first\nsecond\nthird")
	ed.setBlocks([]codeBlock{{line: 1}, {line: 1}, {line: 2}})
	h := editorHarness(ed)
	h.Frame()
	lh, bh := ed.metrics.lh, ed.blocks.height
	if ed.rowTop(0) != 0 || ed.rowTop(1) != lh+2*bh || ed.rowTop(2) != 2*lh+3*bh || ed.contentHeight() != 3*lh+3*bh {
		t.Fatal("block geometry", ed.blocks.rows)
	}
	for row := range 3 {
		top := ed.rowTop(row)
		if got := ed.rowAt(float32(top + lh/2)); got != row {
			t.Fatal("inverse row", row, got)
		}
		p := ed.posAt(f32.Pt(float32(ed.metrics.textOrigin), float32(top+lh/2)))
		if p.line != row || p.col != 0 {
			t.Fatal("pointer source mapping", row, p)
		}
	}
	ed.SetCursor(0, 0)
	ed.Focus()
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	if line, col := ed.Cursor(); line != 1 || col != 0 {
		t.Fatal("keyboard entered control row", line, col)
	}
	h.Key(key.NameDownArrow, 0)
	if line, _ := ed.Cursor(); line != 2 {
		t.Fatal("keyboard skipped source")
	}
	ed.SetValue("replaced")
	h.Frame()
	if len(ed.blocks.items) != 0 || ed.rowTop(0) != 0 {
		t.Fatal("obsolete blocks survived source replacement")
	}
}

func TestCodeEditorBlocksFollowFoldingAndWrap(t *testing.T) {
	ed := CodeEditor("header\n  inside\nlast")
	ed.setBlocks([]codeBlock{{line: 1}, {line: 2}})
	h := editorHarness(ed)
	h.Frame()
	ed.SetCursor(0, 0)
	ed.Fold(0)
	h.Frame()
	if len(ed.blocks.rows) != 1 || ed.blocks.rows[0].index != 1 || ed.blocks.rows[0].row != 1 {
		t.Fatal("hidden source block shown", ed.blocks.rows)
	}
	ed.UnfoldAll()
	h.Frame()
	if len(ed.blocks.rows) != 2 || ed.blocks.rows[1].row != 2 {
		t.Fatal("fold cache not rebuilt", ed.blocks.rows)
	}
	ed.SoftWrap(true)
	h.Frame()
	if ed.rowAt(float32(ed.rowTop(2))) != 2 {
		t.Fatal("wrap map")
	}
}

func TestCodeEditorPageNavigationAccountsForBlocks(t *testing.T) {
	ed := CodeEditor(strings.Repeat("source\n", 100))
	var blocks []codeBlock
	for i := 0; i < 100; i += 2 {
		blocks = append(blocks, codeBlock{line: i})
	}
	ed.setBlocks(blocks)
	h := editorHarness(ed)
	ed.SetCursor(0, 0)
	ed.Focus()
	h.Frame()
	want := ed.vrow(ed.rowAt(float32(ed.rowTop(0) + ed.metrics.size.Y))).line
	h.Key(key.NamePageDown, 0)
	if got, _ := ed.Cursor(); got != want {
		t.Fatal("page jumped past controls", got, want)
	}
}
