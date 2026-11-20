package el

import (
	"fmt"
	"image"
	"testing"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestSignalSetDeduplicatesAndTracksComputed(t *testing.T) {
	count := NewSignal(1)
	genBefore := loop.Generation()
	if count.Set(1) {
		t.Fatal("Set with equal value should return false")
	}
	if loop.Generation() != genBefore || count.Version() != 0 {
		t.Fatalf("unchanged Set bumped generation=%d version=%d", loop.Generation()-genBefore, count.Version())
	}

	evals := 0
	doubled := NewComputed(func() int {
		evals++
		return count.Get() * 2
	})
	if got := doubled.Get(); got != 2 || evals != 1 {
		t.Fatalf("got=%d evals=%d", got, evals)
	}
	if got := doubled.Get(); got != 2 || evals != 1 {
		t.Fatalf("cached Computed re-evaluated: got=%d evals=%d", got, evals)
	}
	if !count.Set(3) {
		t.Fatal("Set with new value should return true")
	}
	if loop.Generation() == genBefore || count.Version() != 1 {
		t.Fatal("changed Set did not invalidate or bump version")
	}
	if got := doubled.Get(); got != 6 || evals != 2 {
		t.Fatalf("after update: got=%d evals=%d", got, evals)
	}
	if count.Peek() != 3 {
		t.Fatalf("Peek = %d", count.Peek())
	}
}

func TestMemoRebuildsOnlyAffectedSubtreesAndReusesLayout(t *testing.T) {
	left := NewSignal("L0")
	right := NewSignal("R0")
	leftBuilds, rightBuilds := 0, 0
	var leftEl, rightEl Element

	root := Root(ViewFunc(func(cx *Context) Element {
		leftEl = cx.Memo("left", func() Element {
			leftBuilds++
			return Div().P(4).Child(Text(left.Get()))
		})
		rightEl = cx.Memo("right", func() Element {
			rightBuilds++
			return Div().P(4).Child(Text(right.Get()))
		})
		return Div().Row().Gap(8).Child(leftEl, rightEl)
	}))

	h := uitest.New(root)
	if leftBuilds != 1 || rightBuilds != 1 {
		t.Fatalf("frame 1 builds: left=%d right=%d", leftBuilds, rightBuilds)
	}
	// Second frame with no changes: neither rebuilds, both reuse layout.
	h.Frame()
	if leftBuilds != 1 || rightBuilds != 1 {
		t.Fatalf("frame 2 builds: left=%d right=%d", leftBuilds, rightBuilds)
	}
	if !leftEl.node().reused || !rightEl.node().reused {
		t.Fatalf("expected layout reuse: left=%v right=%v", leftEl.node().reused, rightEl.node().reused)
	}

	// Changing only left rebuilds only left; right skips build AND reuses layout.
	left.Set("L1")
	h.Frame()
	if leftBuilds != 2 || rightBuilds != 1 {
		t.Fatalf("after left.Set: left=%d right=%d", leftBuilds, rightBuilds)
	}
	if leftEl.node().reused || !rightEl.node().reused {
		t.Fatalf("layout reuse after left.Set: left=%v right=%v", leftEl.node().reused, rightEl.node().reused)
	}

	// Theme switch invalidates memos just like Context.Cache.
	theme.Apply(theme.Dark())
	t.Cleanup(func() { theme.Apply(theme.Light()) })
	h.Frame()
	if leftBuilds != 3 || rightBuilds != 2 {
		t.Fatalf("after theme switch: left=%d right=%d", leftBuilds, rightBuilds)
	}
}

func TestNestedMemoUpdatesInnerInPlaceWithoutRebuildingOuter(t *testing.T) {
	outerSig := NewSignal("outer0")
	innerSig := NewSignal("inner0")
	sibSig := NewSignal("sib0")
	outerBuilds, innerBuilds, sibBuilds := 0, 0, 0
	var innerNode *Node

	root := Root(ViewFunc(func(cx *Context) Element {
		return cx.Memo("outer", func() Element {
			outerBuilds++
			inner := cx.Memo("inner", func() Element {
				innerBuilds++
				el := Text(innerSig.Get())
				innerNode = el.node()
				return el
			})
			sib := cx.Memo("sib", func() Element {
				sibBuilds++
				return Text(sibSig.Get())
			})
			return Div().Col().Gap(4).Child(Text(outerSig.Get()), inner, sib)
		})
	}))

	h := uitest.New(root)
	if outerBuilds != 1 || innerBuilds != 1 || sibBuilds != 1 {
		t.Fatalf("frame 1: outer=%d inner=%d sib=%d", outerBuilds, innerBuilds, sibBuilds)
	}

	// Update only innerSig: outer and sib must NOT rebuild!
	innerSig.Set("inner1")
	h.Frame()
	if outerBuilds != 1 || innerBuilds != 2 || sibBuilds != 1 {
		t.Fatalf("after innerSig.Set: outer=%d inner=%d sib=%d", outerBuilds, innerBuilds, sibBuilds)
	}
	if innerNode.text != "inner1" {
		t.Fatalf("inner text = %q, want inner1", innerNode.text)
	}

	// Verify subsequent clean frame does not sweep nested memos or rebuild.
	h.Frame()
	if outerBuilds != 1 || innerBuilds != 2 || sibBuilds != 1 {
		t.Fatalf("after clean frame: outer=%d inner=%d sib=%d", outerBuilds, innerBuilds, sibBuilds)
	}

	// Update outerSig: outer rebuilds, inner and sib remain cached!
	outerSig.Set("outer1")
	h.Frame()
	if outerBuilds != 2 || innerBuilds != 2 || sibBuilds != 1 {
		t.Fatalf("after outerSig.Set: outer=%d inner=%d sib=%d", outerBuilds, innerBuilds, sibBuilds)
	}
}

func TestMemoRejectsFrameArenaEscape(t *testing.T) {
	root := Root(ViewFunc(func(cx *Context) Element {
		frame := Text("frame-local")
		return cx.Memo("bad", func() Element { return Div().Child(frame) })
	}))
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when Memo returns frame-arena element")
		}
	}()
	var ops op.Ops
	gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	gtx.Constraints.Max = image.Pt(200, 100)
	root.Layout(gtx)
}

func BenchmarkMemoPartialUpdate(b *testing.B) {
	const rows = 200
	sigs := make([]*Signal[int], rows)
	for i := range sigs {
		sigs[i] = NewSignal(i)
	}

	runBench := func(b *testing.B, useMemo bool) {
		root := Root(ViewFunc(func(cx *Context) Element {
			list := Div().Col().Gap(2)
			for i := range sigs {
				if useMemo {
					s := sigs[i]
					list.Child(cx.Memo(i, func() Element {
						return Div().Row().Gap(8).P(4).Child(
							Text(fmt.Sprintf("Row %d", i)),
							Text(fmt.Sprintf("val=%d", s.Get())),
						)
					}))
				} else {
					list.Child(Div().Row().Gap(8).P(4).Child(
						Text(fmt.Sprintf("Row %d", i)),
						Text(fmt.Sprintf("val=%d", sigs[i].Peek())),
					))
				}
			}
			return list
		}))
		var ops op.Ops
		gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
		gtx.Constraints.Max = image.Pt(600, 800)
		root.Layout(gtx)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sigs[0].Set(i + 1)
			ops.Reset()
			root.Layout(gtx)
		}
	}

	b.Run("full-rebuild", func(b *testing.B) { runBench(b, false) })
	b.Run("memo-single-row", func(b *testing.B) { runBench(b, true) })
}

func BenchmarkMemoScrollViewportPartialUpdate(b *testing.B) {
	// Realistic desktop scenario: 500-row list inside a 400px viewport (only ~12 rows
	// visible/painted), where one row's Signal updates per frame.
	const rows = 500
	sigs := make([]*Signal[int], rows)
	for i := range sigs {
		sigs[i] = NewSignal(i)
	}

	runBench := func(b *testing.B, useMemo bool) {
		root := Root(ViewFunc(func(cx *Context) Element {
			list := Div().Col().ScrollY().W(Dp(600)).H(Dp(400)).Gap(4)
			for i := range sigs {
				if useMemo {
					s := sigs[i]
					list.Child(cx.Memo(i, func() Element {
						return Div().Row().Gap(8).P(6).Bg(theme.Surface).Rounded(6).Child(
							Text(fmt.Sprintf("Item #%03d — detailed status description", i)),
							Text(fmt.Sprintf("count=%d", s.Get())),
						)
					}))
				} else {
					list.Child(Div().Row().Gap(8).P(6).Bg(theme.Surface).Rounded(6).Child(
						Text(fmt.Sprintf("Item #%03d — detailed status description", i)),
						Text(fmt.Sprintf("count=%d", sigs[i].Peek())),
					))
				}
			}
			return list
		}))
		var ops op.Ops
		gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
		gtx.Constraints.Max = image.Pt(600, 400)
		root.Layout(gtx)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sigs[0].Set(i + 1)
			ops.Reset()
			root.Layout(gtx)
		}
	}

	b.Run("full-rebuild", func(b *testing.B) { runBench(b, false) })
	b.Run("memo-single-row", func(b *testing.B) { runBench(b, true) })
}
