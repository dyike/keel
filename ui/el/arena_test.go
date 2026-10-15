package el

import (
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func arenaTestFrame(root *RootWidget) func() {
	var ops op.Ops
	var router input.Router
	return func() {
		ops.Reset()
		root.Layout(core.C{
			Ops: &ops, Source: router.Source(), Now: time.Unix(0, 0),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Constraints{Max: image.Pt(400, 300)},
		})
		router.Frame(&ops)
	}
}

func arenaPanic(t *testing.T, message string, fn func()) {
	t.Helper()
	defer func() {
		p := recover()
		if p == nil || !strings.Contains(fmt.Sprint(p), message) {
			t.Errorf("panic = %v, want %q", p, message)
		}
	}()
	fn()
}

func TestArenaReusesNodesButNotPointerHandles(t *testing.T) {
	var box *DivEl
	var label *TextEl
	var field *InputEl
	var embedded *WidgetEl
	root := Root(ViewFunc(func(*Context) Element {
		box = Div().Wrap()
		label = Text("label").Ranges(TextRange{Start: 0, End: 1})
		field = Input().Placeholder("field")
		embedded = Widget(core.Func(func(core.C) core.D { return core.D{Size: image.Pt(2, 2)} }))
		return box.Child(label, field, embedded)
	}))
	frame := arenaTestFrame(root)
	frame()
	frame() // the first layout does not reserve chunks for a one-shot root
	oldBox, oldLabel, oldField, oldEmbedded := box, label, field, embedded
	nodes := [4]*Node{box.node(), label.node(), field.node(), embedded.node()}
	childStorage := &box.node().children[0]
	rangeStorage := &label.node().textRanges[0]
	frame()
	if box == oldBox || label == oldLabel || field == oldField || embedded == oldEmbedded {
		t.Fatal("a pointer handle was overwritten by the next pass")
	}
	if got := [4]*Node{box.node(), label.node(), field.node(), embedded.node()}; got != nodes {
		t.Fatal("node slots were not reused")
	}
	if childStorage != &box.node().children[0] || rangeStorage != &label.node().textRanges[0] {
		t.Fatal("node slice storage was not retained")
	}
	for _, fn := range []func(){
		func() { oldBox.P(20) },
		func() { oldLabel.Ranges() },
		func() { oldField.Placeholder("stale") },
		func() { oldEmbedded.Hidden(true) },
		func() { oldBox.When(false, nil) },
		func() { Div().Child(oldLabel) },
	} {
		arenaPanic(t, "after its build pass", fn)
	}
}

func TestArenaExpiresBeforeSlotReuseAndKeepsOwnerIdentity(t *testing.T) {
	var saved *DivEl
	pass := 0
	root := Root(ViewFunc(func(*Context) Element {
		if pass != 0 {
			arenaPanic(t, "after its build pass", func() { saved.IsHidden() })
		}
		pass++
		saved = Div()
		return saved
	}))
	frame := arenaTestFrame(root)
	frame()
	frame()
	foreign := Root(ViewFunc(func(*Context) Element { return saved }))
	arenaPanic(t, "another root", arenaTestFrame(foreign))
	if buildingArena != nil {
		t.Fatal("a panicking render left the allocator installed")
	}
	frame()
}

func TestArenaRejectsPreviousPassAfterClickRerender(t *testing.T) {
	var built []*DivEl
	calls := 0
	root := Root(ViewFunc(func(*Context) Element {
		box := Div().Name("button").Size(Dp(30)).OnClick(func() { calls++ })
		built = append(built, box)
		return box
	}))
	h := uitest.New(root)
	h.Click(10, 10)
	if calls != 1 || len(built) != 3 {
		t.Fatalf("calls=%d builds=%d; want a click and a same-frame rerender", calls, len(built))
	}
	arenaPanic(t, "after its build pass", func() { built[1].P(1) })
	built[2].P(1)
}

func TestArenaShrinksAndClearsOmittedReferences(t *testing.T) {
	count := 130
	var omitted *TextEl
	root := Root(ViewFunc(func(*Context) Element {
		box := Div()
		for i := range count {
			text := Text(strings.Repeat("payload", 20)).OnClick(func() {})
			if i == count-1 {
				omitted = text
			}
			box.Child(text)
		}
		return box
	}))
	frame := arenaTestFrame(root)
	frame()
	frame() // populate reusable storage after the one-shot-root fast path
	old, slot := omitted, omitted.node()
	count = 1
	frame()
	if len(root.arena.chunks) != 2 {
		t.Fatalf("kept %d chunks; want the used chunk and one spare", len(root.arena.chunks))
	}
	if slot.text != "" || slot.onClick != nil || slot.children != nil || slot.arena != nil {
		t.Fatal("a saved handle retained references in a dropped chunk")
	}
	arenaPanic(t, "after its build pass", func() { old.Name("stale") })
	count = 33
	frame()
	if len(root.arena.chunks) != 2 {
		t.Fatal("small list growth did not reuse the spare chunk")
	}
}

func TestArenaCacheAndStandaloneElementsKeepTheirStorage(t *testing.T) {
	standalone := Text("retained")
	var cached Element
	builds := 0
	root := Root(ViewFunc(func(cx *Context) Element {
		cached = cx.Cache("cached", func() Element {
			builds++
			return Div().Child(Text("cached").Bold())
		})
		return Div().Child(standalone, cached)
	}))
	frame := arenaTestFrame(root)
	frame()
	old := cached
	for range 5 {
		frame()
		standalone.TextSize(14)
	}
	if builds != 1 || cached != old || cached.node().arena != nil || standalone.node().arena != nil {
		t.Fatal("cached or standalone storage became frame-local")
	}
}

func TestArenaRejectsEscapingFrameNodesThroughCache(t *testing.T) {
	root := Root(ViewFunc(func(cx *Context) Element {
		frame := Text("temporary")
		return cx.Cache("bad", func() Element { return frame })
	}))
	arenaPanic(t, "persistent subtree", arenaTestFrame(root))
	if buildingArena != nil {
		t.Fatal("a cache failure left the allocator installed")
	}
	root = Root(ViewFunc(func(cx *Context) Element {
		frame := Text("temporary")
		return cx.Cache("bad", func() Element { return Div().Child(frame) })
	}))
	arenaPanic(t, "persistent element", arenaTestFrame(root))
}

func TestArenaGenerationCannotWrapToAnOldHandle(t *testing.T) {
	a := nodeArena{owner: &arenaOwner{generation: ^uint64(0)}}
	arenaPanic(t, "generation exhausted", a.begin)
}

func TestArenaOverlayFocusQueriesSurviveDispatchAndRerender(t *testing.T) {
	var duringBuild, duringClick bool
	root := Root(ViewFunc(func(cx *Context) Element {
		duringBuild = cx.FocusWithin("panel")
		cx.Focus("inside")
		cx.Overlay("overlay", Modal(Div().ID("panel").Size(Dp(100)).Child(
			Div().ID("inside").Name("inside").Size(Dp(50)).OnClick(func() {
				duringClick = cx.FocusWithin("panel")
			}))))
		return Div()
	}))
	h := uitest.New(root)
	h.Frame()
	if !duringBuild {
		t.Fatal("building lost the previous overlay's focus ancestry")
	}
	h.Click(170, 120)
	if !duringBuild || !duringClick {
		t.Fatalf("overlay focus: build=%v click=%v", duringBuild, duringClick)
	}
}

func TestArenaNestedRootRestoresTheEnclosingAllocator(t *testing.T) {
	var innerBox, outerBox *DivEl
	inner := Root(ViewFunc(func(cx *Context) Element {
		cx.Focus("inner")
		innerBox = Div().ID("inner").Focusable(true).Child(Text("inner"))
		return innerBox
	}))
	innerFrame := arenaTestFrame(inner)
	outer := Root(ViewFunc(func(*Context) Element {
		innerFrame()
		outerBox = Div().Child(Text("outer"))
		return outerBox
	}))
	frame := arenaTestFrame(outer)
	for range 3 {
		frame()
		if innerBox.node().arena != inner.arena.owner || outerBox.node().arena != outer.arena.owner {
			t.Fatal("nested construction used the wrong root's arena")
		}
		if buildingArena != nil {
			t.Fatal("nested rendering left an allocator installed")
		}
	}
}

func TestArenaSteadyBuildAllocationCeiling(t *testing.T) {
	root := Root(ViewFunc(func(*Context) Element {
		box := Div()
		for range 100 {
			box.Child(Text("same text").Bg(theme.Primary).TextColor(theme.Text).Bold().Mono().Italic().TextAlign(Start).Top(0).Left(0).RoundedCorners(1, 2, 3, 4).Selected(false))
		}
		return box
	}))
	cx := &Context{root: root}
	for range 3 {
		root.renderTree(cx)
	}
	if got := testing.AllocsPerRun(20, func() { root.renderTree(cx) }); got > 6 {
		t.Fatalf("warmed 101-node build allocated %.0f times; want at most 6 handle batches", got)
	}
}

func TestArenaOneShotRootDoesNotReserveNodeChunks(t *testing.T) {
	var saved *TextEl
	root := Root(ViewFunc(func(*Context) Element {
		saved = Text("one shot")
		return Div().Child(saved)
	}))
	frame := arenaTestFrame(root)
	frame()
	if len(root.arena.chunks) != 0 || root.arena.divs.batch != nil || root.arena.texts.batch != nil {
		t.Fatal("a one-shot root reserved reusable chunks or handle batches")
	}
	old := saved
	frame()
	if len(root.arena.chunks) == 0 {
		t.Fatal("the second build did not start pooling")
	}
	arenaPanic(t, "after its build pass", func() { old.TextSize(1) })
}

func TestArenaMountedViewsDoNotMutateCachedRoot(t *testing.T) {
	var cached *DivEl
	root := Root(ViewFunc(func(cx *Context) Element {
		cx.Mount("mounted", ViewFunc(func(*Context) Element { return Text("mounted").Absolute() }))
		cached = cx.Cache("root", func() Element { return Div().P(8).Child(Text("persistent")) }).(*DivEl)
		return cached
	}))
	frame := arenaTestFrame(root)
	for range 4 {
		frame()
		if len(cached.node().children) != 1 || len(root.mainTree.children) != 2 {
			t.Fatal("mounts leaked into a cached root or accumulated across frames")
		}
	}
}

func TestArenaContentBottomCannotRetainAFrameTarget(t *testing.T) {
	retained := Div()
	root := Root(ViewFunc(func(cx *Context) Element {
		temporary := Text("temporary")
		cx.Cache("bad", func() Element { return retained.ContentBottom(temporary) })
		return Div()
	}))
	arenaPanic(t, "persistent element", arenaTestFrame(root))
}
