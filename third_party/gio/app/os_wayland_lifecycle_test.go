// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland

package app

import "testing"

// Exercise the real Wayland Event path without connecting to a compositor.
func TestWaylandEventAcknowledgesViewBeforeDestroy(t *testing.T) {
	w := &window{w: &callbacks{w: new(Window)}}
	cleaned, released := false, false
	w.teardown.begin(func() {
		w.ProcessEvent(WaylandViewEvent{})
	}, func() {
		if !cleaned {
			t.Fatal("native teardown preceded consumer cleanup")
		}
		released = true
		w.ProcessEvent(DestroyEvent{})
	})
	view, ok := w.Event().(WaylandViewEvent)
	if !ok || view.Valid() || released {
		t.Fatal("first Event must return the invalid view with handles alive")
	}
	cleaned = true // Keel releases its proxies before calling Event again.
	if _, ok := w.Event().(DestroyEvent); !ok || !released {
		t.Fatal("next Event must finish teardown and return DestroyEvent")
	}
}
