package main

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/el"
)

func TestMessageCopyReadsLatestDocument(t *testing.T) {
	m := &message{doc: newDocument("original")}
	view := el.Embed(m.copy())
	m.doc = newDocument("updated answer")
	var router input.Router
	var ops op.Ops
	frame := func() {
		ops.Reset()
		view.Layout(layout.Context{Ops: &ops, Now: time.Now(), Source: router.Source(), Constraints: layout.Constraints{Max: image.Pt(400, 300)}, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}})
		router.Frame(&ops)
	}
	frame()
	p := f32.Pt(20, 12)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
	frame()
	if _, data, ok := router.WriteClipboard(); !ok || string(data) != "updated answer" {
		t.Fatalf("copied stale answer: %q %v", data, ok)
	}
}
