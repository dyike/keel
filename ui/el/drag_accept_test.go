package el

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestDragAcceptDecidesOnceAndReportsCancellation(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		for _, accept := range []bool{false, true} {
			decisions := 0
			var delta f32.Point
			var got []DragEvent
			filtered, disabled := true, false
			root := Embed(ViewFunc(func(*Context) Element {
				d := Div().W(Dp(200)).H(Dp(100)).OnDrag(func(e DragEvent) { got = append(got, e) }).Disabled(disabled)
				if filtered {
					d.DragAccept(func(dx, dy float32) bool { decisions++; delta = f32.Pt(dx, dy); return accept })
				}
				return Div().Child(d)
			}))
			h := uitest.NewFunc(func(gtx core.C) { gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}; root.Layout(gtx) })
			send := func(kind pointer.Kind, id pointer.ID, x float32) {
				h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Touch, PointerID: id, Position: f32.Pt(x*scale, 20*scale)})
				h.Frame()
			}
			send(pointer.Press, 1, 20)
			send(pointer.Move, 1, 21)
			if decisions != 0 {
				t.Fatal("decided before threshold", decisions)
			}
			send(pointer.Press, 2, 30)
			send(pointer.Release, 2, 30)
			send(pointer.Move, 1, 30)
			send(pointer.Move, 1, 40)
			send(pointer.Release, 1, 40)
			if decisions != 1 || delta != f32.Pt(10, 0) {
				t.Fatal("decision count/units", scale, decisions, delta)
			}
			ends := 0
			for _, e := range got {
				if e.Kind == DragEnd {
					ends++
					if e.Canceled == accept {
						t.Fatal("wrong end", e, accept)
					}
				}
			}
			if ends != 1 {
				t.Fatal("end count", got)
			}
			// Replacing the filter with ordinary dragging and restoring it must not
			// retain the prior recognizer's pressed state.
			filtered = false
			h.Frame()
			got = nil
			send(pointer.Press, 3, 20)
			send(pointer.Move, 3, 40)
			send(pointer.Release, 3, 40)
			if len(got) == 0 || got[len(got)-1].Canceled || got[len(got)-1].Kind != DragEnd {
				t.Fatal("ordinary drag after filter", got)
			}
			filtered = true
			h.Frame()
			send(pointer.Press, 4, 20)
			send(pointer.Move, 4, 40)
			send(pointer.Release, 4, 40)
			if decisions != 2 {
				t.Fatal("restored filter stale", decisions)
			}
			disabled = true
			h.Frame()
			got = nil
			send(pointer.Press, 5, 20)
			send(pointer.Move, 5, 40)
			send(pointer.Release, 5, 40)
			if len(got) != 0 || decisions != 2 {
				t.Fatal("disabled drag", got, decisions)
			}
		}
	}
}

func TestConditionalDragHonorsMovedAndRemovedChildHitAreas(t *testing.T) {
	childX := float32(40)
	show, disabled := true, false
	clicks, presses := 0, 0
	h := uitest.New(Embed(ViewFunc(func(*Context) Element {
		parent := Div().P(10).W(Dp(200)).H(Dp(100)).DragAccept(func(float32, float32) bool { return true }).OnDrag(func(e DragEvent) {
			if e.Kind == DragStart {
				presses++
			}
		})
		if show {
			parent.Child(Div().W(Dp(40)).H(Dp(30)).Translate(childX, 0).Disabled(disabled).OnClick(func() { clicks++ }))
		}
		return Div().P(10).Child(parent)
	})))
	h.Click(65, 25)
	if clicks != 1 || presses != 0 {
		t.Fatal("translated child press stolen", clicks, presses)
	}
	childX = 100
	h.Frame()
	h.Drag(65, 25, 80, 25)
	if presses != 1 {
		t.Fatal("old child location still blocked", presses)
	}
	h.Click(125, 25)
	if clicks != 2 || presses != 1 {
		t.Fatal("new child location unprotected", clicks, presses)
	}
	disabled = true
	h.Frame()
	h.Drag(125, 25, 140, 25)
	if presses != 2 || clicks != 2 {
		t.Fatal("disabled child blocked drag", presses, clicks)
	}
	show = false
	h.Frame()
	h.Drag(125, 25, 140, 25)
	if presses != 3 {
		t.Fatal("removed child blocked drag", presses)
	}
}
