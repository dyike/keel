package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

// RegionOpen reports whether a side region is shown. The center is always
// open.
func (v *DockView) RegionOpen(s DockSide) bool {
	if closed := v.regionClosed(s); closed != nil {
		return !*closed
	}
	return true
}

// SetRegionOpen collapses or reopens a side region. Its panels keep their
// tabs, splits and size, and a closed region is part of Layout.
func (v *DockView) SetRegionOpen(s DockSide, open bool) {
	closed := v.regionClosed(s)
	if closed == nil || *closed == !open {
		return
	}
	v.cancelResize()
	v.drag = dockDrag{}
	*closed = !open
	if !open {
		if z := v.layout.Zoomed; z != "" && v.where(z) == int(s) {
			v.layout.Zoomed = ""
		}
	}
}

// ToggleRegion flips a side region and reports the change to OnLayoutChange.
func (v *DockView) ToggleRegion(s DockSide) {
	if v.regionClosed(s) == nil || v.disabled {
		return
	}
	v.SetRegionOpen(s, !v.RegionOpen(s))
	v.changed()
}

// RegionButton is a toggle for a side region, selected while it is open, to
// put in a title bar or toolbar. Build it once and keep it; it reads the
// region's state on every Render.
func (v *DockView) RegionButton(s DockSide) el.View {
	if v.regionClosed(s) == nil {
		return el.ViewFunc(func(*el.Context) el.Element { return nil })
	}
	text := locale.Current()
	name := map[DockSide]string{DockLeft: text.DockLeftRegion, DockRight: text.DockRightRegion, DockBottom: text.DockBottomRegion}[s]
	b := Button("", func() { v.ToggleRegion(s) }).Name(name).Variant(ButtonGhost).Size(28)
	b.icon = dockRegionIcon(s)
	return el.ViewFunc(func(cx *el.Context) el.Element {
		b.SetSelected(v.RegionOpen(s))
		b.SetDisabled(v.disabled)
		return b.Render(cx)
	})
}

func (v *DockView) regionClosed(s DockSide) *bool {
	switch s {
	case DockLeft:
		return &v.layout.LeftClosed
	case DockRight:
		return &v.layout.RightClosed
	case DockBottom:
		return &v.layout.BottomClosed
	}
	return nil
}

// dockRegionIcon outlines a window with the region's side filled in.
func dockRegionIcon(s DockSide) *IconView {
	pane := map[DockSide]string{
		DockLeft:   `<rect x="3" y="4" width="6" height="16" rx="1"/>`,
		DockRight:  `<rect x="15" y="4" width="6" height="16" rx="1"/>`,
		DockBottom: `<rect x="3" y="14" width="18" height="6" rx="1"/>`,
	}[s]
	icon, err := SVGIcon([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">` +
		`<path fill-rule="evenodd" d="M4 3h16a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2zm0 2v14h16V5z"/>` + pane + `</svg>`))
	if err != nil {
		panic(err) // the markup is fixed
	}
	return icon
}
