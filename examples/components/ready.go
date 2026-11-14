package main

import "github.com/dyike/keel/ui/core"

// Notify the web loader after layout has produced drawing operations, rather
// than when Gio merely creates its canvas. Preserve the root sizing contract.
type galleryFirstFrame struct {
	core.Widget
	ready bool
}

func (w *galleryFirstFrame) FillsWindow() bool {
	f, ok := w.Widget.(interface{ FillsWindow() bool })
	return ok && f.FillsWindow()
}
func (w *galleryFirstFrame) Layout(gtx core.C) core.D {
	d := galleryLayout(w.Widget, gtx)
	if !w.ready {
		w.ready = true
		galleryReady()
	}
	return d
}
