// Package uitest drives UI code without a window: it lays it out through a real
// Gio input router, so tests click and type as the platform would. The viewport
// is 400×300 with 1dp = 1px.
package uitest

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
)

type Harness struct {
	Router input.Router
	ops    op.Ops
	layout func(gtx core.C)
}

// New lays out w at the origin once.
func New(w core.Widget) *Harness { return NewFunc(func(gtx core.C) { w.Layout(gtx) }) }

// NewFunc lays out any frame function once, e.g. a window's own layout.
func NewFunc(fn func(gtx core.C)) *Harness {
	h := &Harness{layout: fn}
	h.Frame()
	return h
}

// Frame runs pending updates and lays out one frame.
func (h *Harness) Frame() {
	h.ops.Reset()
	gtx := layout.Context{Ops: &h.ops, Now: time.Now(), Source: h.Router.Source(),
		Constraints: layout.Constraints{Max: image.Pt(400, 300)}, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	loop.Lock()
	loop.Drain()
	h.layout(gtx)
	loop.Unlock()
	h.Router.Frame(&h.ops)
}

// Click presses and releases the primary button at (x, y), then renders.
func (h *Harness) Click(x, y float32) {
	p := f32.Pt(x, y)
	h.Router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p},
	)
	h.Frame()
}

// Key presses and releases a key with modifiers, then renders.
func (h *Harness) Key(name key.Name, mods key.Modifiers) {
	h.Router.Queue(
		key.Event{Name: name, Modifiers: mods, State: key.Press},
		key.Event{Name: name, Modifiers: mods, State: key.Release},
	)
	h.Frame()
}

// Type inserts text into the focused editor, then renders.
func (h *Harness) Type(s string) {
	h.Router.Queue(key.EditEvent{Text: s})
	h.Frame()
}
