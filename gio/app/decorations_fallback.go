// SPDX-License-Identifier: Unlicense OR MIT

//go:build (!darwin || ios) && !windows

package app

import (
	"fmt"
	"image"

	"github.com/dyike/keel/gio/font/gofont"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget"
	"github.com/dyike/keel/gio/widget/material"
)

// Keel patch: only platforms that may need client-side decorations carry
// their Material widgets, font collection and text shaper.
type fallbackDecorations struct {
	*material.Theme
	*widget.Decorations
}

func (w *Window) initDecorations() unit.Dp {
	// Measure decoration height.
	deco := new(widget.Decorations)
	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Regular()))
	decoStyle := material.Decorations(theme, deco, 0, "")
	gtx := layout.Context{
		Ops: new(op.Ops),
		// Measure in Dp.
		Metric: unit.Metric{},
	}
	// Allow plenty of space.
	gtx.Constraints.Max.Y = 200
	dims := decoStyle.Layout(gtx)
	decoHeight := unit.Dp(dims.Size.Y)
	w.decorations.Theme = theme
	w.decorations.Decorations = deco
	return decoHeight
}

func (w *Window) updateDecorationMode(mode WindowMode) {
	w.decorations.Decorations.Maximized = mode == Maximized
}

func (w *Window) fallbackDecorate() bool {
	cnf := w.decorations.Config
	return w.decorations.enabled && !cnf.Decorated && cnf.Mode != Fullscreen && !w.nocontext
}

// decorate the window if enabled and returns the corresponding Insets.
func (w *Window) decorate(e FrameEvent, o *op.Ops) image.Point {
	if !w.fallbackDecorate() {
		return image.Pt(0, 0)
	}
	deco := w.decorations.Decorations
	allActions := system.ActionMinimize | system.ActionMaximize | system.ActionUnmaximize |
		system.ActionClose | system.ActionMove
	style := material.Decorations(w.decorations.Theme, deco, allActions, w.decorations.Config.Title)
	// Update the decorations based on the current window mode.
	var actions system.Action
	switch m := w.decorations.Config.Mode; m {
	case Windowed:
		actions |= system.ActionUnmaximize
	case Minimized:
		actions |= system.ActionMinimize
	case Maximized:
		actions |= system.ActionMaximize
	case Fullscreen:
		actions |= system.ActionFullscreen
	default:
		panic(fmt.Errorf("unknown WindowMode %v", m))
	}
	gtx := layout.Context{
		Ops:         o,
		Now:         e.Now,
		Source:      e.Source,
		Metric:      e.Metric,
		Constraints: layout.Exact(e.Size),
	}
	// Update the window based on the actions on the decorations.
	opts, acts := splitActions(deco.Update(gtx))
	if len(opts) > 0 {
		w.driver.Configure(opts)
	}
	if acts != 0 {
		w.driver.Perform(acts)
	}
	style.Layout(gtx)
	// Offset to place the frame content below the decorations.
	decoHeight := gtx.Dp(w.decorations.Config.decoHeight)
	if w.decorations.currentHeight != decoHeight {
		w.decorations.currentHeight = decoHeight
		w.coalesced.cfg = &ConfigEvent{Config: w.effectiveConfig()}
	}
	return image.Pt(0, decoHeight)
}
