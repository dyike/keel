// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/semantic"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
)

// An accessibility bridge finds a button in the tree and presses it.
func TestAppendSemanticsAndProcessInput(t *testing.T) {
	w := new(Window)
	tag := new(int)
	var ops op.Ops
	area := clip.Rect(image.Rect(10, 20, 110, 60)).Push(&ops)
	event.Op(&ops, tag)
	semantic.Button.Add(&ops)
	semantic.LabelOp("Save").Add(&ops)
	area.Pop()
	// Register the filter, as a frame's layout does before the frame.
	w.queue.Event(pointer.Filter{Target: tag, Kinds: pointer.Press | pointer.Release})
	w.queue.Frame(&ops)

	var button *semantic.ClassOp
	var bounds image.Rectangle
	for _, n := range w.AppendSemantics(nil) {
		if n.Desc.Class == semantic.Button && n.Desc.Label == "Save" {
			button, bounds = &n.Desc.Class, n.Desc.Bounds
		}
	}
	if button == nil {
		t.Fatal("no Save button in the semantic tree")
	}
	if want := image.Rect(10, 20, 110, 60); bounds != want {
		t.Fatalf("button bounds %v, want %v", bounds, want)
	}
	c := bounds.Min.Add(bounds.Max).Div(2)
	p := f32.Pt(float32(c.X), float32(c.Y))
	w.ProcessInput(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p})
	w.ProcessInput(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p})
	if !w.ProcessInput(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p}) {
		t.Fatal("release not handled")
	}
	var kinds []pointer.Kind
	for {
		e, ok := w.queue.Event(pointer.Filter{Target: tag, Kinds: pointer.Press | pointer.Release})
		if !ok {
			break
		}
		kinds = append(kinds, e.(pointer.Event).Kind)
	}
	if len(kinds) != 2 || kinds[0] != pointer.Press || kinds[1] != pointer.Release {
		t.Fatalf("button received %v, want press and release", kinds)
	}
}
