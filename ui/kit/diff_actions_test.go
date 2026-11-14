package kit

import (
	"fmt"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestDiffEmbeddedActionsTargetAndPreserveSource(t *testing.T) {
	for _, mode := range []DiffMode{DiffModeUnified, DiffModeSplit} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			v := Diff("a\nb\nc\nd\n", "A\nb\nc\nD\n").Mode(mode).Height(450)
			var got []DiffHunk
			factories := 0
			v.HunkActions(func(h DiffHunk) []DiffAction {
				factories++
				return []DiffAction{{Label: fmt.Sprintf("Inspect %d", h.Index), OnSelect: func(current DiffHunk) { got = append(got, current) }},
					{Label: fmt.Sprintf("Restore %d", h.Index), OnSelect: func(current DiffHunk) { v.RevertAt(current.Index) }}}
			})
			h := renderView(v, 900, 1)
			click(t, h, "Inspect 1")
			if len(got) != 1 || got[0].Index != 1 || got[0].Before != "d\n" || v.Current() != 0 {
				t.Fatal(got, v.Current())
			}
			if factories != 2 {
				t.Fatal("factory called during frames", factories)
			}
			ed := v.unified
			if mode == DiffModeSplit {
				ed = v.new
			}
			ed.Focus()
			h.Frame()
			h.Key("A", key.ModShortcut)
			h.Key("C", key.ModShortcut)
			expected := v.Patch("before", "after")
			if mode == DiffModeSplit {
				expected = v.Value().After
			}
			if _, data, ok := h.Router.WriteClipboard(); !ok || string(data) != expected {
				t.Fatal("controls entered source copy", string(data))
			}
			click(t, h, "Restore 1")
			if v.Changes() != 1 || v.Value().After != "A\nb\nc\nd\n" {
				t.Fatal("wrong region restored", v.Value())
			}
			v.HunkActions(nil)
			h.Frame()
			if len(v.old.blocks.items)+len(v.new.blocks.items)+len(v.unified.blocks.items) != 0 {
				t.Fatal("rows retained after removal")
			}
		})
	}
}

func TestDiffActionsDisabledAndPatchSnapshot(t *testing.T) {
	v := Diff("", "")
	if err := v.SetPatch("--- a/f\n+++ b/f\n@@ -90 +80 @@\n-old\n+new\n"); err != nil {
		t.Fatal(err)
	}
	calls := 0
	v.HunkActions(func(DiffHunk) []DiffAction {
		return []DiffAction{
			{Label: "Inspect patch region", OnSelect: func(h DiffHunk) {
				calls++
				if !h.PatchOnly || h.BeforeStart != 89 || h.AfterStart != 79 {
					t.Fatal(h)
				}
			}},
			{Label: "Unavailable action", Disabled: true, OnSelect: func(DiffHunk) { t.Fatal("disabled action") }},
		}
	})
	disabled := false
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 390, 1)
	click(t, h, "Inspect patch region")
	if calls != 1 {
		t.Fatal(calls)
	}
	disabled = true
	h.Frame()
	click(t, h, "Inspect patch region")
	if calls != 1 {
		t.Fatal("ancestor disabled ignored")
	}
	disabled = false
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "Inspect patch region")
	if calls != 1 {
		t.Fatal("disabled view ignored")
	}
}

func TestDiffActionsFoldedScrollAndIdle(t *testing.T) {
	before := strings.Repeat("context\n", 30) + "old\n" + strings.Repeat("tail\n", 30)
	after := strings.Replace(before, "old\n", "new\n", 1)
	v := Diff(before, after).Mode(DiffModeSplit).Height(250)
	v.HunkActions(func(DiffHunk) []DiffAction {
		return []DiffAction{{Label: "Inspect change", OnSelect: func(DiffHunk) {}}}
	})
	h := renderView(v, 900, 1)
	click(t, h, "Inspect change")
	x, y := center(bounds(h, "After"))
	h.Scroll(x, y, 100)
	h.Frame()
	h.Frame()
	settled := v.lastScroll
	for range 6 {
		h.Frame()
	}
	if v.lastScroll != settled {
		t.Fatal("scroll synchronization did not settle")
	}
	v.Expanded(true)
	h.Frame()
	v.SetCurrent(0)
	h.Frame()
	click(t, h, "Inspect change")
}

func TestDiffActionsStaleGeneration(t *testing.T) {
	v := Diff("old", "new")
	calls := 0
	v.HunkActions(func(DiffHunk) []DiffAction {
		return []DiffAction{{Label: "Old action", OnSelect: func(DiffHunk) { calls++ }}}
	})
	old := v.unified.blocks.items[0].widget
	h := uitest.New(old)
	h.Frame()
	v.SetValue(DiffValue{"different", "comparison"})
	click(t, h, "Old action")
	if calls != 0 {
		t.Fatal("stale button acted on new comparison")
	}
}

func TestDiffActionsNarrowHorizontalScroll(t *testing.T) {
	v := Diff("old", "new").Mode(DiffModeUnified)
	calls := 0
	v.HunkActions(func(DiffHunk) []DiffAction {
		return []DiffAction{
			{Label: "First custom action", OnSelect: func(DiffHunk) { calls++ }},
			{Label: "Last custom action", OnSelect: func(DiffHunk) { calls++ }},
		}
	})
	root := el.Embed(v)
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: 1.6}
		gtx.Constraints.Max = image.Pt(320, 1000)
		root.Layout(gtx)
	})
	first := bounds(h, "First custom action")
	if first.Empty() {
		t.Fatal("first action missing")
	}
	x, y := center(first)
	for _, event := range []pointer.Event{
		{Kind: pointer.Press, Position: f32.Pt(x, y)},
		{Kind: pointer.Move, Position: f32.Pt(x-60, y)},
		{Kind: pointer.Move, Position: f32.Pt(5, y)},
		{Kind: pointer.Release, Position: f32.Pt(5, y)},
	} {
		event.Source, event.PointerID = pointer.Touch, 1
		h.Router.Queue(event)
		h.Frame()
	}
	h.Frame()
	if calls != 0 {
		t.Fatal("drag activated action")
	}
	// Reach the end deterministically; the preceding sequence verifies touch
	// scrolling doesn't accidentally activate the button.
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(100, y), Scroll: f32.Pt(1000, 0)})
	h.Frame()
	h.Frame()
	last := bounds(h, "Last custom action")
	if last.Empty() || last.Max.X > 320 {
		t.Fatal("last action unreachable", last)
	}
	click(t, h, "Last custom action")
	if calls != 1 {
		t.Fatal(calls)
	}
}

func TestDiffActionScrollMapsInsertionAndDeletionHeaders(t *testing.T) {
	for _, value := range []DiffValue{
		{"same\ntail\n", "same\ninserted\ntail\n"},
		{"same\ndeleted\ntail\n", "same\ntail\n"},
	} {
		v := Diff(value.Before, value.After).Mode(DiffModeSplit)
		v.HunkActions(func(DiffHunk) []DiffAction { return []DiffAction{{Label: "Action", OnSelect: func(DiffHunk) {}}} })
		renderView(v, 900, 1)
		for _, reverse := range []bool{false, true} {
			source, target := v.old, v.new
			if reverse {
				source, target = target, source
			}
			source.scrollY = float32(source.rowStart(source.blocks.rows[0].row) + source.blocks.height/2)
			want := float32(target.rowStart(target.blocks.rows[0].row) + target.blocks.height/2)
			if got := diffScrollPosition(source, target, v.changes, reverse); got != want {
				t.Fatal("header mapping", reverse, got, want)
			}
		}
	}
}
