// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/input"
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

// stubDriver accepts what processing input asks of a driver.
type stubDriver struct{ driver }

func (stubDriver) SetAnimating(bool)                   {}
func (stubDriver) SetCursor(pointer.Cursor)            {}
func (stubDriver) Invalidate()                         {}
func (stubDriver) ShowTextInput(bool)                  {}
func (stubDriver) EditorStateChanged(_, _ editorState) {}

// The observer sees each frame's tree, and input queued from another
// goroutine reaches handlers before the next event.
func TestObserveSemanticsAndQueueInput(t *testing.T) {
	w := new(Window)
	w.driver = stubDriver{}
	tag := new(int)
	var ops op.Ops
	area := clip.Rect(image.Rect(10, 20, 110, 60)).Push(&ops)
	event.Op(&ops, tag)
	semantic.Button.Add(&ops)
	semantic.LabelOp("Save").Add(&ops)
	area.Pop()
	filter := pointer.Filter{Target: tag, Kinds: pointer.Press | pointer.Release}
	w.queue.Event(filter)
	w.queue.Frame(&ops)
	w.lastFrame.size = image.Pt(200, 100)
	var seen []string
	var size image.Point
	w.ObserveSemantics(func(nodes []input.SemanticNode, s image.Point) {
		for _, n := range nodes {
			if n.Desc.Label != "" {
				seen = append(seen, n.Desc.Label)
			}
		}
		size = s
	})
	w.observeFrame()
	if len(seen) != 1 || seen[0] != "Save" || size != image.Pt(200, 100) {
		t.Fatalf("observer saw %v at %v", seen, size)
	}
	p := f32.Pt(60, 40)
	done := make(chan struct{})
	go func() {
		w.QueueInput(
			pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p},
			pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p},
			pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p},
		)
		close(done)
	}()
	<-done
	w.nextEvent()
	var kinds []pointer.Kind
	for {
		e, ok := w.queue.Event(filter)
		if !ok {
			break
		}
		kinds = append(kinds, e.(pointer.Event).Kind)
	}
	if len(kinds) != 2 || kinds[0] != pointer.Press || kinds[1] != pointer.Release {
		t.Fatalf("queued input delivered %v", kinds)
	}
	w.ObserveSemantics(nil)
	seen = nil
	w.observeFrame()
	if seen != nil {
		t.Fatal("observer called after removal")
	}
}
