package kit

import (
	"github.com/dyike/keel/gio/io/input"
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestCodeEditorLineNumberPress(t *testing.T) {
	ed := CodeEditor("func f() {\n  return\n}\nlast\n")
	var got []int
	ed.OnLineNumberPress(func(line int) { got = append(got, line) })
	h := editorHarness(ed)
	h.Frame()
	x := float32(ed.metrics.foldW+ed.metrics.gutter) / 2
	y := float32(ed.metrics.lh) / 2
	h.Click(x, y)
	if len(got) != 1 || got[0] != 0 || ed.Selection() != "func f() {\n" {
		t.Fatal(got, ed.Selection())
	}
	ed.SetCursor(0, 0)
	h.Frame()
	h.Click(float32(ed.metrics.foldW)/2, y)
	if len(got) != 1 || !ed.folds[0] {
		t.Fatal("fold arrow notified or failed", got)
	}
	h.Click(x, float32(ed.metrics.lh)*2.5)
	if len(got) != 2 || got[1] != 3 || ed.Selection() != "last\n" {
		t.Fatal("folded row lost source identity", got, ed.Selection())
	}
	ed.SetDisabled(true)
	h.Frame()
	h.Click(x, y)
	if len(got) != 2 {
		t.Fatal("disabled editor notified")
	}
}

func TestDiffLineNumberSourceMappings(t *testing.T) {
	v := Diff("same\nold", "same\nnew")
	renderView(v, 600, 1)
	want := map[int]DiffLine{3: {DiffSideAfter, 0}, 4: {DiffSideBefore, 1}, 6: {DiffSideAfter, 1}}
	for row := 0; row < v.unified.Lines(); row++ {
		line, ok := v.sourceLine(2, row)
		expected, exists := want[row]
		if ok != exists || ok && line != expected {
			t.Fatalf("row %d: %v %v, want %v %v", row, line, ok, expected, exists)
		}
	}
	raw := "--- a/file\n+++ b/file\n@@ -900000000,1 +800000000,1 @@\n-old\n+new\n"
	if err := v.SetPatch(raw); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		pane, row int
		want      DiffLine
	}{
		{0, 1, DiffLine{DiffSideBefore, 899999999}},
		{1, 1, DiffLine{DiffSideAfter, 799999999}},
		{2, 3, DiffLine{DiffSideBefore, 899999999}},
		{2, 4, DiffLine{DiffSideAfter, 799999999}},
	} {
		got, ok := v.sourceLine(tc.pane, tc.row)
		if !ok || got != tc.want {
			t.Fatal(tc, got, ok)
		}
	}
	for pane := range 3 {
		if _, ok := v.sourceLine(pane, 0); ok {
			t.Fatal("header has source")
		}
	}
	v.SetValue(DiffValue{"equal\n", "equal\n"})
	renderView(v, 600, 1)
	if got, ok := v.sourceLine(2, 0); !ok || got != (DiffLine{DiffSideAfter, 0}) {
		t.Fatal(got, ok)
	}
}

func TestDiffGutterClickAndAncestorDisabled(t *testing.T) {
	v := Diff("old\n", "new\n").Mode(DiffModeSplit)
	var got []DiffLine
	v.OnLineNumberPress(func(line DiffLine) { got = append(got, line) })
	disabled := false
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 900, 1)
	var b image.Rectangle
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Label == locale.Current().DiffBefore && n.Desc.Class.String() == "Editor" {
			b = n.Desc.Bounds
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, n := range h.Router.AppendSemantics(nil) {
		walk(n)
	}
	if b.Empty() {
		t.Fatal("missing editor")
	}
	x := float32(b.Min.X) + float32(v.old.metrics.foldW+v.old.metrics.gutter)/2
	y := float32(b.Min.Y) + float32(v.old.metrics.lh)/2
	h.Click(x, y)
	if len(got) != 1 || got[0] != (DiffLine{DiffSideBefore, 0}) {
		t.Fatal(got)
	}
	disabled = true
	h.Frame()
	h.Click(x, y)
	if len(got) != 1 {
		t.Fatal("disabled ancestor notified", got)
	}
}

func TestDiffLineMappingRebuiltForExpandedContext(t *testing.T) {
	before := "first\n" + strings.Repeat("context\n", 20) + "old\n"
	after := strings.TrimSuffix(before, "old\n") + "new\n"
	v := Diff(before, after)
	renderView(v, 600, 1)
	if got, _ := v.sourceLine(2, 3); got.Line == 0 {
		t.Fatal("unexpected distant context")
	}
	v.Expanded(true)
	if got, ok := v.sourceLine(2, 3); !ok || got != (DiffLine{DiffSideAfter, 0}) {
		t.Fatal(got, ok)
	}
}
