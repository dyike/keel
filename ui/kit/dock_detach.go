package kit

import "slices"

// OnDetach lets panels leave the dock for a window of their own: their menu
// gains "Open in new window", and dragging a tab out of the dock detaches it.
// fn should open a window showing p.View; when that window closes, call
// reattach (from a UI callback, or through core.Update) to put the panel
// back where it was. kit does not open windows itself, so the app chooses
// the window's size and options.
//
//	d.OnDetach(func(p kit.DockPanel, reattach func()) {
//	    window.Open(window.Options{Title: p.Title, Content: el.Root(p.View), OnClose: reattach})
//	})
func (v *DockView) OnDetach(fn func(p DockPanel, reattach func())) *DockView {
	v.onDetach = fn
	return v
}

// Detach moves a shown panel out of the dock into a window of its own, by
// calling the OnDetach function. It does nothing without one.
func (v *DockView) Detach(id string) {
	if v.onDetach == nil || v.disabled || !v.Visible(id) {
		return
	}
	v.cancelResize()
	v.drag = dockDrag{}
	if v.layout.Zoomed == id {
		v.layout.Zoomed = ""
	}
	v.layout.Detached = append(v.layout.Detached, id)
	if s := v.where(id); s >= 0 {
		v.fixActive(DockSide(s))
	}
	v.changed()
	v.onDetach(v.panels[id], func() { v.reattach(id) })
}

// Detached lists the panels now in windows of their own. A restored layout
// does not reopen windows: SetLayout puts detached panels back in the dock.
func (v *DockView) Detached() []string { return slices.Clone(v.layout.Detached) }

func (v *DockView) reattach(id string) {
	if !slices.Contains(v.layout.Detached, id) {
		return
	}
	v.layout.Detached = slices.DeleteFunc(v.layout.Detached, func(s string) bool { return s == id })
	if s := v.where(id); s >= 0 {
		n := findDockGroup(*v.tree(DockSide(s)), id)
		if n != nil {
			n.Active = id
		}
		_, active := v.side(DockSide(s))
		*active = id
		v.fixActive(DockSide(s))
	} else {
		v.Move(id, DockLeft)
	}
	v.changed()
}
