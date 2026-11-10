// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"image"
	"testing"
)

type invalidationDriver struct {
	stubDriver
	calls int
}

func (d *invalidationDriver) Invalidate() { d.calls++ }

func TestInvalidateDuringFrame(t *testing.T) {
	d := new(invalidationDriver)
	w := &Window{driver: d, mayInvalidate: true}
	w.Invalidate()
	w.Invalidate()
	if d.calls != 1 {
		t.Fatalf("pending invalidations = %d, want 1", d.calls)
	}
	w.coalesced.frame = &frameEvent{}
	if _, ok := w.nextEvent(); !ok {
		t.Fatal("frame not delivered")
	}
	// An external update arriving after layout has started cannot be
	// included in this frame. It must request another one.
	w.Invalidate()
	if d.calls != 2 {
		t.Fatalf("invalidation during a delivered frame was lost: calls = %d", d.calls)
	}
	w.Invalidate()
	if d.calls != 2 {
		t.Fatal("duplicate next-frame invalidation was not coalesced")
	}
}

func TestCoalescedFrameRetainsSurfaceRefresh(t *testing.T) {
	w := &Window{driver: new(invalidationDriver)}
	w.processEvent(frameEvent{FrameEvent: FrameEvent{Size: image.Pt(800, 600)}, Sync: true})
	w.processEvent(frameEvent{FrameEvent: FrameEvent{Size: image.Pt(1000, 750)}})
	if !w.lastFrame.sync || !w.coalesced.frame.Sync || w.lastFrame.size != image.Pt(1000, 750) {
		t.Fatal("animation frame discarded pending resize refresh")
	}
}
