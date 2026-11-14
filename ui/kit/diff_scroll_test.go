package kit

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestDiffLineMapping(t *testing.T) {
	c := []diffRange{{2, 2, 2, 5}, {8, 10, 11, 12}}
	for _, tc := range []struct {
		line, want float32
		reverse    bool
	}{{1, 1, false}, {2, 5, false}, {7, 10, false}, {9, 11.5, false}, {11, 13, false}, {3, 2, true}, {6, 3, true}, {11.5, 9, true}} {
		if got := diffMappedLine(tc.line, c, tc.reverse); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}

func TestDiffSplitScrollSynchronizesWithoutFeedback(t *testing.T) {
	var before, after strings.Builder
	for i := range 150 {
		fmt.Fprintf(&before, "line %d\n", i)
		if i == 10 {
			after.WriteString("inserted one\ninserted two\ninserted three\n")
		}
		fmt.Fprintf(&after, "line %d\n", i)
	}
	v := Diff(before.String(), after.String()).Mode(DiffModeSplit).Expanded(true).Height(200)
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return v.Render(cx) }), 900, 1)
	h.Frame()
	// Real wheel events over each editor, not directly assigned offsets.
	x, y := center(bounds(h, "Before"))
	h.Scroll(x, y, 450)
	h.Frame()
	h.Frame()
	if v.old.scrollY <= 0 || v.new.scrollY <= 0 {
		t.Fatal("wheel did not scroll both panes")
	}
	want := diffMappedLine(v.old.scrollY/float32(v.old.metrics.lh), v.changes, false) * float32(v.new.metrics.lh)
	if math.Abs(float64(v.new.scrollY-want)) > 0.01 {
		t.Fatal("forward mapping", v.old.scrollY, v.new.scrollY, want)
	}
	x, y = center(bounds(h, "After"))
	h.Scroll(x, y, 300)
	h.Frame()
	h.Frame()
	want = diffMappedLine(v.new.scrollY/float32(v.new.metrics.lh), v.changes, true) * float32(v.old.metrics.lh)
	if math.Abs(float64(v.old.scrollY-want)) > 0.01 {
		t.Fatal("reverse mapping", v.old.scrollY, want)
	}
	settled := v.lastScroll
	for range 5 {
		h.Frame()
	}
	if v.lastScroll != settled {
		t.Fatal("idle scroll feedback", settled, v.lastScroll)
	}
	v.SetCurrent(0)
	h.Frame()
	h.Frame()
	for _, ed := range []*CodeEditorView{v.old, v.new} {
		line, _ := ed.Cursor()
		top := float32(line*ed.metrics.lh) - ed.scrollY
		if top < 0 || top+float32(ed.metrics.lh) > float32(ed.metrics.size.Y) {
			t.Fatal("change cursor outside viewport", line, top, ed.metrics.size.Y)
		}
	}
}
