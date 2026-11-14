package kit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/locale"
)

func TestDiffSplitContextFoldingPreservesTextAndSelection(t *testing.T) {
	var before, after strings.Builder
	for i := range 120 {
		fmt.Fprintf(&before, "line %d\n", i)
		if i == 20 || i == 90 {
			fmt.Fprintf(&after, "changed %d\n", i)
		} else {
			fmt.Fprintf(&after, "line %d\n", i)
		}
	}
	v := Diff(before.String(), after.String()).Mode(DiffModeSplit).Height(250)
	h := renderView(v, 900, 1)
	h.Frame()
	if v.old.rowCount() >= 50 || !v.old.hidden(60) || v.old.hidden(20) || v.new.hidden(90) {
		t.Fatal("context not folded around changes", v.old.rowCount())
	}
	v.old.Focus()
	h.Frame()
	h.Key("A", key.ModShortcut)
	h.Key("C", key.ModShortcut)
	if _, data, ok := h.Router.WriteClipboard(); !ok || string(data) != before.String() {
		t.Fatal("folded source copy lost lines")
	}
	click(t, h, locale.Current().DiffExpand)
	if v.old.rowCount() != v.old.Lines() || v.new.rowCount() != v.new.Lines() {
		t.Fatal("expand did not restore full rows")
	}
	click(t, h, locale.Current().DiffCollapse)
	if !v.old.hidden(60) {
		t.Fatal("collapse not restored")
	}
	v.SetCurrent(1)
	h.Frame()
	h.Frame()
	if v.old.hidden(90) || v.new.hidden(90) {
		t.Fatal("selected change hidden")
	}
	if v.Value().Before != before.String() || v.Value().After != after.String() {
		t.Fatal("folding changed source")
	}
	x, y := center(bounds(h, locale.Current().DiffBefore))
	h.Scroll(x, y, -60)
	h.Frame()
	h.Frame()
	oldLine := v.old.lineOf(int(v.old.scrollY / float32(v.old.metrics.lh)))
	wantRow := v.new.rowOf(int(diffMappedLine(float32(oldLine), v.changes, false)))
	gotRow := int(v.new.scrollY / float32(v.new.metrics.lh))
	if gotRow != wantRow {
		t.Fatal("folded scroll mismatch", oldLine, wantRow, gotRow)
	}
	settled := v.lastScroll
	for range 5 {
		h.Frame()
	}
	if v.lastScroll != settled {
		t.Fatal("folded synchronization feedback")
	}
}
