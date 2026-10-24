// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
)

// TestPendingReportsUndeliveredInput checks Source.Pending (Keel patch): no
// events before input, input until the frame that delivers it, and none
// after a frame whose handlers did not move.
func TestPendingReportsUndeliveredInput(t *testing.T) {
	var r Router
	tag := new(int)
	frame := func() {
		var ops op.Ops
		area := clip.Rect(image.Rect(0, 0, 100, 100)).Push(&ops)
		event.Op(&ops, tag)
		area.Pop()
		r.Frame(&ops)
	}
	drain := func() {
		for {
			if _, ok := r.Event(pointer.Filter{Target: tag, Kinds: pointer.Press | pointer.Release | pointer.Enter | pointer.Leave | pointer.Move}); !ok {
				return
			}
		}
	}
	frame()
	drain()
	frame()
	drain()
	if r.Source().Pending() {
		t.Fatal("pending before any input")
	}
	r.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(10, 10)})
	if !r.Source().Pending() {
		t.Fatal("queued press not pending")
	}
	drain()
	frame()
	r.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(10, 10)})
	if !r.Source().Pending() {
		t.Fatal("queued release not pending")
	}
	drain()
	frame()
	drain()
	frame()
	if r.Source().Pending() {
		t.Fatal("pending after the input was delivered and the scene kept still")
	}
	if (Source{}).Pending() {
		t.Fatal("zero Source pending")
	}
}
