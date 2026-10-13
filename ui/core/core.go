// Package core is the foundation every UI module builds on: what a Widget is,
// how a component runs a user callback, and how other goroutines change the UI.
//
// Threading has one rule. All windows render, and all callbacks run, under one
// lock, so callbacks may change any component directly. Code on other goroutines
// (timers, network, hotkey callbacks) must wrap changes in Update.
package core

import (
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"image"

	"github.com/dyike/keel/ui/internal/loop"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Widget is anything that can lay itself out. Every component and container is one.
type Widget interface {
	Layout(gtx C) D
}

// ViewportWidget can avoid painting content outside its parent's visible area.
// Bounds are in widget-local pixels, independent of its full layout constraints.
// An empty viewport requests measurement only; layout dimensions stay natural.
type ViewportWidget interface {
	Widget
	LayoutViewport(gtx C, visible image.Rectangle) D
}

// ViewportFunc adapts a layout function that retains the parent's visible area.
type ViewportFunc func(gtx C, visible image.Rectangle) D

func (f ViewportFunc) Layout(gtx C) D {
	return f(gtx, image.Rectangle{Max: gtx.Constraints.Max})
}
func (f ViewportFunc) LayoutViewport(gtx C, visible image.Rectangle) D { return f(gtx, visible) }

// Func adapts a plain Gio layout function to Widget.
type Func func(gtx C) D

func (f Func) Layout(gtx C) D { return f(gtx) }

// Update queues fn under the frame lock, where it may change components safely.
// Normally the next frame runs it; resident mode also drains without frames.
// It is safe from any goroutine, including from callbacks, and returns at once.
func Update(fn func()) { loop.Post(fn) }

// Call runs a user callback from inside Layout and redraws every window,
// because the callback may have changed components that were already drawn.
// Components must route every user callback through it.
func Call(gtx C, fn func()) {
	if fn == nil {
		return
	}
	fn()
	gtx.Execute(op.InvalidateCmd{})
	loop.InvalidateAll()
}
