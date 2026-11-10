package el

import (
	"image"
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func dragPointer(h *uitest.Harness, kind pointer.Kind, x, y float32) {
	buttons := pointer.ButtonPrimary
	if kind == pointer.Release || kind == pointer.Cancel {
		buttons = 0
	}
	h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Mouse, Position: f32.Pt(x, y), Buttons: buttons})
	h.Frame()
}

func TestDraggableFollowsPointerAcrossFramesAndSettlesWithoutRender(t *testing.T) {
	now := time.Unix(1000, 0)
	renders, clicks, releases := 0, 0, 0
	var moved, released DragPoint
	root := Root(ViewFunc(func(cx *Context) Element {
		renders++
		cx.Overlay("tip", Anchored("card", Div().Name("tip").Size(Dp(10))).Placement(Bottom, Start))
		return Div().Items(Start).P(10).Child(Div().ID("card").Name("card").Size(Dp(60)).OnClick(func() { clicks++ }).Draggable(DragOptions{
			Axis: DragHorizontal, Bounds: &DragBounds{MaxX: 200}, Snap: []DragPoint{{}, {X: 200}}, Animation: Tween(time.Second, Linear),
			OnMove: func(p DragPoint) { moved = p }, OnRelease: func(p DragPoint) { released = p; releases++ },
		}))
	}))
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	dragPointer(h, pointer.Press, 30, 30)
	for _, x := range []float32{80, 130, 180} {
		now = now.Add(16 * time.Millisecond)
		dragPointer(h, pointer.Move, x, 40)
		if moved != (DragPoint{X: x - 30}) {
			t.Fatalf("pointer x=%v offset=%v", x, moved)
		}
		if got := nodeBounds(h, "card"); got.Min != image.Pt(int(x)-20, 10) {
			t.Fatalf("pointer x=%v box=%v", x, got)
		}
		if got := nodeBounds(h, "tip"); got.Min != image.Pt(int(x)-20, 74) {
			t.Fatalf("anchor=%v", got)
		}
	}
	dragPointer(h, pointer.Release, 180, 40)
	if released != (DragPoint{X: 200}) || releases != 1 || clicks != 0 {
		t.Fatalf("release=%v count=%d clicks=%d", released, releases, clicks)
	}
	// The first quiet frame consumes the input-driven rebuild.
	h.Frame()
	before := renders
	now = now.Add(500 * time.Millisecond)
	h.Frame()
	if renders != before || root.repaints == 0 {
		t.Fatalf("settling rebuilt view: %d -> %d", before, renders)
	}
	if got := nodeBounds(h, "card"); got.Min.X != 185 {
		t.Fatalf("halfway box=%v", got)
	}
	now = now.Add(time.Second)
	h.Frame()
	h.Frame()
	if got := nodeBounds(h, "card"); got.Min.X != 210 {
		t.Fatalf("settled box=%v", got)
	}
	if root.motionDue {
		t.Fatal("settled drag requests frames")
	}
	h.Click(230, 30)
	if clicks != 1 || releases != 1 {
		t.Fatalf("tap clicks=%d drag releases=%d", clicks, releases)
	}
}

func TestDraggableComposesParentAndOwnTransforms(t *testing.T) {
	var moved DragPoint
	h := uitest.New(Root(ViewFunc(func(*Context) Element {
		return Div().P(60).Items(Start).Child(Div().Size(Dp(100)).Items(Start).AnimatePose(Pose{Scale: 2, Rotate: 90}, Instant).Child(
			Div().ID("card").Name("card").Size(Dp(40)).AnimatePose(Pose{Scale: 1.2, Rotate: 20}, Instant).Draggable(DragOptions{OnMove: func(p DragPoint) { moved = p }}),
		))
	})))
	b := nodeBounds(h, "card")
	x, y := center(b)
	dragPointer(h, pointer.Press, x, y)
	for i := 1; i <= 3; i++ {
		dragPointer(h, pointer.Move, x, y+float32(i)*20)
		if math.Abs(float64(moved.X-float32(i)*10)) > .01 || math.Abs(float64(moved.Y)) > .01 {
			t.Fatalf("parent-axis move %d: %v", i, moved)
		}
	}
	dragPointer(h, pointer.Release, x, y+60)
}

func TestDraggableLeavesTransformedChildrenInteractive(t *testing.T) {
	moves, clicks := 0, 0
	text := ""
	h := uitest.New(Root(ViewFunc(func(*Context) Element {
		return Div().Items(Start).Child(Div().ID("drag").Size(Dp(250)).Draggable(DragOptions{OnMove: func(DragPoint) { moves++ }}).Items(Start).Child(
			Div().ID("button").Name("button").Size(Dp(40)).AnimatePose(Pose{X: 120, Rotate: 45}, Instant).OnClick(func() { clicks++ }),
			Input().ID("input").Name("input").W(Dp(150)).Bind(&text),
		))
	})))
	x, y := center(nodeBounds(h, "button"))
	h.Click(x, y)
	h.Drag(x, y, x+30, y)
	if clicks != 1 || moves != 0 {
		t.Fatalf("child click=%d parent moves=%d", clicks, moves)
	}
	h.Click(center(nodeBounds(h, "input")))
	h.Type("中文")
	if text != "中文" || moves != 0 {
		t.Fatalf("input=%q moves=%d", text, moves)
	}
	// The old, untransformed child position belongs to the draggable parent.
	h.Drag(20, 20, 50, 20)
	if moves == 0 {
		t.Fatal("old child position incorrectly blocked the parent")
	}
}

func TestDraggableElasticMomentumCancellationAndReducedMotion(t *testing.T) {
	now := time.Unix(1100, 0)
	st := elemState{}
	st.prepareDrag(DragOptions{Bounds: &DragBounds{MaxX: 100}, Elastic: .5, Momentum: true, Animation: defaultAnimation(Smooth)})
	d := st.dragMotion
	d.handle(pointer.Event{Kind: pointer.Press}, now, 1, false)
	d.handle(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(160, 0)}, now.Add(50*time.Millisecond), 1, true)
	if d.target.X <= 100 || d.target.X >= 160 {
		t.Fatalf("rubberband=%v", d.target)
	}
	d.handle(pointer.Event{Kind: pointer.Release, Position: f32.Pt(160, 0)}, now.Add(60*time.Millisecond), 1, true)
	if d.target.X != 100 || d.values[0].Velocity() <= 0 {
		t.Fatalf("target=%v velocity=%v", d.target, d.values[0].Velocity())
	}
	d.sample(now.Add(60*time.Millisecond), false, 1)
	d.sample(now.Add(80*time.Millisecond), false, 1)
	if d.values[0].Get() <= 100 {
		t.Fatal("spring lost release continuity")
	}
	d.sample(now.Add(10*time.Second), false, 1)
	if d.values[0].Get() != 100 || d.values[0].Animating() {
		t.Fatal("spring did not settle")
	}
	d.handle(pointer.Event{Kind: pointer.Press}, now.Add(11*time.Second), 1, false)
	d.handle(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(-50, 0)}, now.Add(11*time.Second+50*time.Millisecond), 1, true)
	d.handle(pointer.Event{Kind: pointer.Cancel}, now.Add(11*time.Second+60*time.Millisecond), 1, true)
	if d.target.X != 100 || d.values[0].Velocity() != 0 {
		t.Fatal("cancel did not return to drag start")
	}
	d.sample(now.Add(12*time.Second), true, 2)
	if d.values[0].Get() != 100 || d.values[0].Animating() {
		t.Fatal("reduced motion did not settle")
	}
}

func TestDraggableResetDisabledAndOmission(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	reset := uint64(0)
	disabled, enabled := false, true
	var moved DragPoint
	root := Root(ViewFunc(func(*Context) Element {
		card := Div().ID("card").Name("card").Size(Dp(60)).Disabled(disabled)
		if enabled {
			card.Draggable(DragOptions{Reset: reset, Initial: DragPoint{X: 10}, OnMove: func(p DragPoint) { moved = p }})
		}
		return Div().Items(Start).Child(card)
	}))
	h := uitest.New(root)
	h.Drag(30, 20, 80, 20)
	if moved.X != 60 {
		t.Fatalf("direct manipulation with reduced motion: %v", moved)
	}
	dragPointer(h, pointer.Press, 80, 20)
	dragPointer(h, pointer.Move, 90, 20)
	disabled = true
	h.Frame()
	for _, st := range root.store.states {
		if st.dragMotion != nil && st.dragMotion.dragging {
			t.Fatal("disabled drag still active")
		}
	}
	dragPointer(h, pointer.Release, 90, 20)
	disabled = false
	reset++
	h.Frame()
	if got := nodeBounds(h, "card"); got.Min.X != 10 {
		t.Fatalf("reset=%v", got)
	}
	enabled = false
	h.Frame()
	for _, st := range root.store.states {
		if st.dragMotion != nil {
			t.Fatal("omitted declaration retained state")
		}
	}
}

func TestDraggableNativeEventTimesAndFinalReleasePosition(t *testing.T) {
	now := time.Unix(1200, 0)
	st := elemState{}
	st.prepareDrag(DragOptions{Momentum: true, Animation: defaultAnimation(Smooth)})
	d := st.dragMotion
	// All events are dispatched in one frame. Their native times still differ.
	d.handle(pointer.Event{Kind: pointer.Press, Time: time.Second}, now, 1, false)
	d.handle(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(20, 0), Time: time.Second + 20*time.Millisecond}, now, 1, true)
	d.handle(pointer.Event{Kind: pointer.Release, Position: f32.Pt(40, 0), Time: time.Second + 40*time.Millisecond}, now, 1, true)
	if d.values[0].Get() != 40 || math.Abs(float64(d.values[0].Velocity()-1000)) > .1 || d.target.X != 240 {
		t.Fatalf("release x=%v velocity=%v target=%v", d.values[0].Get(), d.values[0].Velocity(), d.target)
	}
	// Holding still before releasing must discard the old fling velocity.
	st.prepareDrag(DragOptions{Reset: 1, Momentum: true, Animation: defaultAnimation(Smooth)})
	d = st.dragMotion
	d.handle(pointer.Event{Kind: pointer.Press, Time: 2 * time.Second}, now, 1, false)
	d.handle(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(20, 0), Time: 2*time.Second + 20*time.Millisecond}, now, 1, true)
	d.handle(pointer.Event{Kind: pointer.Release, Position: f32.Pt(20, 0), Time: 3 * time.Second}, now, 1, true)
	if d.target.X != 20 || d.values[0].Velocity() != 0 {
		t.Fatalf("stale fling: %v velocity=%v", d.target, d.values[0].Velocity())
	}
}

func TestDraggableSnapIgnoresLockedAxis(t *testing.T) {
	st := elemState{}
	st.prepareDrag(DragOptions{Axis: DragHorizontal, Snap: []DragPoint{{X: 90, Y: 500}, {}}, Animation: Instant})
	d := st.dragMotion
	now := time.Unix(1300, 0)
	d.handle(pointer.Event{Kind: pointer.Press}, now, 1, false)
	d.handle(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(80, 30)}, now, 1, true)
	d.handle(pointer.Event{Kind: pointer.Release, Position: f32.Pt(80, 30)}, now, 1, true)
	if d.target != (DragPoint{X: 90}) {
		t.Fatalf("locked axis affected snap: %v", d.target)
	}
}

func BenchmarkDraggableSettlingSample(b *testing.B) {
	d := dragMotionState{options: DragOptions{Animation: defaultAnimation(Smooth)}, target: DragPoint{X: 200, Y: 100}}
	now := time.Unix(1400, 0)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%100 == 0 {
			d.values[0].SetVelocity(0, 200)
			d.values[1].Set(0)
		}
		d.sample(now.Add(time.Duration(i)*time.Millisecond), false, 2)
	}
}

func TestDraggableWithoutCallbacksDoesNotWakeOtherWindows(t *testing.T) {
	root := Root(ViewFunc(func(*Context) Element {
		return Div().Items(Start).Child(Div().ID("card").Name("card").Size(Dp(80)).Draggable(DragOptions{}))
	}))
	h := uitest.New(root)
	generation := loop.Generation()
	h.Drag(20, 20, 100, 20)
	if got := nodeBounds(h, "card"); got.Min.X != 80 {
		t.Fatalf("drag=%v", got)
	}
	if loop.Generation() != generation {
		t.Fatal("internal drag invalidated unrelated windows")
	}
}

func TestDraggableFastReleaseRespectsElasticLimit(t *testing.T) {
	for _, elastic := range []float32{0, .4} {
		now := time.Unix(1400, 0)
		st := elemState{}
		st.prepareDrag(DragOptions{Bounds: &DragBounds{MaxX: 240}, Elastic: elastic, Momentum: true, Animation: Snappy})
		d := st.dragMotion
		d.handle(pointer.Event{Kind: pointer.Press, Time: time.Second}, now, 1, false)
		d.handle(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(225, 0), Time: time.Second + time.Microsecond}, now, 1, true)
		d.handle(pointer.Event{Kind: pointer.Release, Position: f32.Pt(225, 0), Time: time.Second + 2*time.Microsecond}, now, 1, true)
		for i := range 180 {
			m, _ := d.sample(now.Add(time.Duration(i)*time.Second/60), false, 1)
			x := m.Transform(f32.Point{}).X
			if x < -100*elastic || x > 240+100*elastic {
				t.Fatalf("elastic %v escaped: %v", elastic, x)
			}
			if x != d.values[0].Get() {
				t.Fatal("displayed and re-grab positions differ")
			}
		}
		if d.values[0].Animating() || d.values[0].Get() != 240 {
			t.Fatal("bounded release did not settle")
		}
	}
}
