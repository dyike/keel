// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland

package app

import (
	"gioui.org/internal/ops"
	"gioui.org/op"
	"gioui.org/op/paint"
	"image/color"
	"testing"
)

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

func TestWaylandRunKeepsCurrentFrameOpsAlive(t *testing.T) {
	w := &window{w: &callbacks{w: new(Window)}, runWakeup: func() {}}
	reset := false
	var frame, glyphs op.Ops
	m := op.Record(&glyphs)
	paint.ColorOp{Color: color.NRGBA{A: 255}}.Add(&glyphs)
	m.Stop().Add(&frame)
	w.Run(func() {
		glyphs.Reset()
		reset = true
		w.ProcessEvent(wakeupEvent{})
	})
	var reader ops.Reader
	reader.Reset(&frame.Internal)
	if _, ok := reader.Decode(); !ok {
		t.Fatal("current frame was reset by background window action")
	}
	if _, ok := w.Event().(wakeupEvent); !ok {
		t.Fatal("queued window action was not executed by Event")
	}
	if !reset {
		t.Fatal("queued reset did not execute")
	}
}
