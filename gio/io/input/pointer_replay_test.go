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

func TestUnobservedReleaseSurvivesCommandReplay(t *testing.T) {
	for _, terminal := range []pointer.Kind{pointer.Release, pointer.Cancel} {
		t.Run(terminal.String(), func(t *testing.T) {
			var r Router
			outside, button := new(int), new(int)
			outsideFilter := pointer.Filter{Target: outside, Kinds: pointer.Press}
			buttonFilter := pointer.Filter{Target: button, Kinds: pointer.Press | pointer.Release}
			var ops op.Ops
			a := clip.Rect(image.Rect(0, 0, 100, 100)).Push(&ops)
			event.Op(&ops, outside)
			a.Pop()
			b := clip.Rect(image.Rect(100, 0, 200, 100)).Push(&ops)
			event.Op(&ops, button)
			b.Pop()
			r.Event(outsideFilter)
			r.Event(buttonFilter)
			r.Frame(&ops)
			r.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)}, pointer.Event{Kind: terminal, Source: pointer.Mouse, Position: f32.Pt(20, 20)})
			// An animation asks for its next frame before the view dispatches input.
			r.Source().Execute(op.InvalidateCmd{})
			for {
				if _, ok := r.Event(outsideFilter); !ok {
					break
				}
			}
			r.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(150, 20)}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(150, 20)})
			var got []pointer.Kind
			for {
				e, ok := r.Event(buttonFilter)
				if !ok {
					break
				}
				if p, ok := e.(pointer.Event); ok && p.Kind != pointer.Cancel {
					got = append(got, p.Kind)
				}
			}
			if len(got) != 2 || got[0] != pointer.Press || got[1] != pointer.Release {
				t.Fatalf("old receiver kept capture after %v: %v", terminal, got)
			}
		})
	}
}
