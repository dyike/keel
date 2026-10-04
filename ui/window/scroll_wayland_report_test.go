package window

import (
	"testing"

	"github.com/dyike/keel/ui/core"
)

func TestReportWaylandScroll(t *testing.T) {
	defer core.ReportScrollGesture(core.ScrollDeviceUnknown, false, false, false)
	defer fingersDown.Store(false)
	start := core.CurrentScrollGesture().Ended
	reportWaylandScroll(wlAxisFinger, true, false)
	if g := core.CurrentScrollGesture(); g.Device != core.ScrollDeviceTrackpad || !g.Active || !g.Phases {
		t.Fatal("fingers scrolling", g)
	}
	// The lift often comes in a frame of its own, naming no source.
	reportWaylandScroll(wlAxisNone, false, true)
	if g := core.CurrentScrollGesture(); g.Active || g.Ended != start+1 {
		t.Fatal("lift ends the gesture", g)
	}
	reportWaylandScroll(wlAxisNone, false, true)
	if g := core.CurrentScrollGesture(); g.Ended != start+1 {
		t.Fatal("a stray stop is not a second lift")
	}
	reportWaylandScroll(wlAxisWheel, true, false)
	if g := core.CurrentScrollGesture(); g.Device != core.ScrollDeviceWheel {
		t.Fatal("wheel", g)
	}
	reportWaylandScroll(wlAxisContinuous, true, false)
	if g := core.CurrentScrollGesture(); g.Device != core.ScrollDeviceWheel {
		t.Fatal("a continuous source changes nothing", g)
	}
}
