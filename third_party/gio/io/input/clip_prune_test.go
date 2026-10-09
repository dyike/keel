// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"image"
	"math/rand"
	"reflect"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
)

func TestDrawingClipsPreserveRouting(t *testing.T) {
	// Compare to the original collector, which kept every clip, including
	// nested and overlapping regions with inherited cursors and semantics.
	for seed := int64(0); seed < 100; seed++ {
		var original, compact pointerQueue
		a, b := pointerCollector{q: &original}, pointerCollector{q: &compact}
		a.Reset()
		b.Reset()
		rng := rand.New(rand.NewSource(seed))
		var walk func(int)
		walk = func(depth int) {
			for i := 0; i < 8; i++ {
				x, y := rng.Intn(40), rng.Intn(40)
				bounds := image.Rect(x, y, x+10+rng.Intn(40), y+10+rng.Intn(40))
				kind := areaKind(rng.Intn(2))
				a.pushArea(kind, bounds)
				b.pushArea(kind, bounds)
				switch rng.Intn(7) {
				case 0:
					tag := new(int)
					var sa, sb pointerHandler
					a.inputOp(tag, &sa)
					b.inputOp(tag, &sb)
				case 1:
					a.cursor(pointer.CursorPointer)
					b.cursor(pointer.CursorPointer)
				case 2:
					a.semanticLabel("label")
					b.semanticLabel("label")
				case 3:
					a.actionInputOp(system.ActionMove)
					b.actionInputOp(system.ActionMove)
				}
				if depth > 0 && rng.Intn(3) == 0 {
					walk(depth - 1)
				}
				// Original pop: only restores the stack, without pruning.
				n := len(a.nodeStack)
				a.state.nodePlusOne = a.nodeStack[n-1].node + 1
				a.nodeStack = a.nodeStack[:n-1]
				b.popArea()
			}
		}
		walk(3)
		for y := float32(0); y < 80; y += 3 {
			for x := float32(0); x < 80; x += 3 {
				p := f32.Pt(x, y)
				hits := func(q *pointerQueue) ([]event.Tag, pointer.Cursor) {
					var tags []event.Tag
					cursor := q.hitTest(p, func(n *hitNode) bool {
						if n.tag != nil {
							tags = append(tags, n.tag)
						}
						return true
					})
					return tags, cursor
				}
				ah, ac := hits(&original)
				bh, bc := hits(&compact)
				if !reflect.DeepEqual(ah, bh) || ac != bc {
					t.Fatalf("seed %d at %v: hits/cursor differ", seed, p)
				}
				aa, aok := original.ActionAt(p)
				ba, bok := compact.ActionAt(p)
				if aa != ba || aok != bok {
					t.Fatalf("seed %d at %v: actions differ", seed, p)
				}
				as, aok := original.SemanticAt(p)
				bs, bok := compact.SemanticAt(p)
				if as != bs || aok != bok {
					t.Fatalf("seed %d at %v: semantics differ", seed, p)
				}
			}
		}
	}
}

func TestDrawingClipsDoNotAccumulate(t *testing.T) {
	var q pointerQueue
	c := pointerCollector{q: &q}
	c.Reset()
	c.pushArea(areaRect, image.Rect(0, 0, 100, 100))
	c.semanticLabel("text")
	for i := 0; i < 10000; i++ {
		c.pushArea(areaRect, image.Rect(i%100, 0, i%100+1, 10))
		c.popArea()
	}
	c.popArea()
	if len(q.areas) != 2 || len(q.hitTree) != 2 {
		t.Fatalf("drawing clips retained: %d areas, %d hits", len(q.areas), len(q.hitTree))
	}
}
