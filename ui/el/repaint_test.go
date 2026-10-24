package el

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/internal/uitest"
)

// ring paints a bar rotated by AmbientPhase, as kit's spinner does, and
// reports the phase it painted.
func ring(phases *[]float32) Element {
	return Widget(core.Func(func(gtx core.C) core.D {
		phase := AmbientPhase(gtx, time.Second)
		*phases = append(*phases, phase)
		size := gtx.Constraints.Min
		c := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
		defer op.Affine(f32.Affine2D{}.Rotate(c, phase*2*math.Pi)).Push(gtx.Ops).Pop()
		defer clip.Rect(image.Rect(size.X/2-2, 0, size.X/2+2, size.Y/2)).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, color.NRGBA{R: 200, A: 255})
		return core.D{Size: size}
	})).Size(Dp(40))
}

func TestPaintOnlyFramesRepaintRetainedTree(t *testing.T) {
	tick := AmbientFrameInterval
	now := time.Unix(100, 0)
	var renders int
	var phases []float32
	label := "first"
	view := ViewFunc(func(cx *Context) Element {
		renders++
		return Div().Child(Text(label), Div().Size(Dp(60)).OnClick(func() {}), ring(&phases))
	})
	root := Root(view)
	root.RetainFrames()
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		root.Layout(gtx)
	})
	step := func(d time.Duration) {
		now = now.Add(d)
		h.Frame()
	}
	expect := func(what string, wantRenders, wantRepaints int) {
		t.Helper()
		if renders != wantRenders || root.repaints != wantRepaints {
			t.Fatalf("%s: %d renders, %d repaints; want %d, %d", what, renders, root.repaints, wantRenders, wantRepaints)
		}
	}
	expect("first frame", 1, 0)
	step(tick)
	expect("ambient tick", 1, 1)
	if n := len(phases); n < 2 || phases[n-1] == phases[n-2] {
		t.Fatalf("repaint did not advance the phase: %v", phases)
	}
	step(tick)
	expect("next ambient tick", 1, 2)
	step(tick / 4)
	expect("frame before the tick", 2, 2)
	step(tick)
	expect("tick after a full frame", 2, 3)

	core.Update(func() { label = "second" })
	step(tick)
	expect("after core.Update", 3, 3)
	if !hasText(root.mainTree, "second") {
		t.Fatal("full frame after core.Update did not render the change")
	}

	now = now.Add(tick)
	h.Move(30, 30)
	expect("frame with input", 4, 3)
	step(tick)
	expect("frame after input", 5, 3)
	step(tick)
	expect("quiet tick", 5, 4)

	loop.Changed()
	step(tick)
	expect("after a window change", 6, 4)
}

// hasText reports whether a text node under n shows s.
func hasText(n *Node, s string) bool {
	if n == nil {
		return false
	}
	if n.isText && n.text == s {
		return true
	}
	for _, c := range n.children {
		if hasText(c.node(), s) {
			return true
		}
	}
	return false
}

// Roots a window does not own, and views that read the frame time while
// rendering, always render.
func TestPaintOnlyFramesNeedWindowAndPaintTimeAnimation(t *testing.T) {
	now := time.Unix(200, 0)
	var phases []float32
	for _, tc := range []struct {
		name   string
		retain bool
		view   View
	}{
		{"embedded root", false, ViewFunc(func(*Context) Element { return ring(&phases) })},
		{"render-time animation", true, ViewFunc(func(cx *Context) Element {
			_ = cx.Now()
			return ring(&phases)
		})},
	} {
		root := Root(tc.view)
		if tc.retain {
			root.RetainFrames()
		}
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Now = now
			root.Layout(gtx)
		})
		for range 4 {
			now = now.Add(AmbientFrameInterval)
			h.Frame()
		}
		if root.repaints != 0 {
			t.Errorf("%s: %d paint-only frames, want none", tc.name, root.repaints)
		}
	}
}

// A paint-only frame draws what a full frame at the same time draws.
func TestPaintOnlyFramePixelsMatchFullFrame(t *testing.T) {
	gpu, err := headless.NewWindow(400, 300)
	if err != nil {
		t.Skip(err)
	}
	defer gpu.Release()
	now := time.Unix(300, 0)
	var phases []float32
	root := Root(ViewFunc(func(*Context) Element {
		return Div().P(20).Child(Text("Loading"), ring(&phases))
	}))
	root.RetainFrames()
	var shot *image.RGBA
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		shot = image.NewRGBA(image.Rect(0, 0, 400, 300))
		if err := gpu.Screenshot(shot); err != nil {
			t.Fatal(err)
		}
	})
	now = now.Add(AmbientFrameInterval * 7)
	h.Frame()
	if root.repaints != 1 {
		t.Fatalf("%d paint-only frames, want 1", root.repaints)
	}
	repainted := shot
	// The same frame time again, rendered in full.
	loop.Changed()
	h.Frame()
	if root.repaints != 1 {
		t.Fatal("forced frame repainted")
	}
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			if a, b := repainted.RGBAAt(x, y), shot.RGBAAt(x, y); a != b {
				t.Fatalf("pixel (%d,%d): repaint %v, full frame %v", x, y, a, b)
			}
		}
	}
}

// benchmarkFrames lays out a spinner beside 60 rows of text on ambient ticks,
// with or without paint-only frames.
func benchmarkFrames(b *testing.B, retain bool) {
	var phases []float32
	root := Root(ViewFunc(func(*Context) Element {
		phases = phases[:0]
		rows := Div().Col().Gap(4)
		for i := range 60 {
			rows.Child(Div().Row().Gap(8).Child(Text("Order"), Text(string(rune('A'+i%26))), Text("Pending")))
		}
		return Div().Col().P(16).Gap(12).Child(ring(&phases), rows)
	}))
	if retain {
		root.RetainFrames()
	}
	now := time.Unix(400, 0)
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		root.Layout(gtx)
	})
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		now = now.Add(AmbientFrameInterval)
		h.Frame()
	}
	b.StopTimer()
	if retain && root.repaints < b.N-1 {
		b.Fatalf("%d paint-only frames of %d", root.repaints, b.N)
	}
}

func BenchmarkAmbientFrameFull(b *testing.B)      { benchmarkFrames(b, false) }
func BenchmarkAmbientFramePaintOnly(b *testing.B) { benchmarkFrames(b, true) }
