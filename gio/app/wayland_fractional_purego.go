// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nowayland && (!cgo || !giocgo || purego)

package app

import "unsafe"

// Handwritten protocol metadata from fractional-scale-v1.xml and viewporter.xml.
// These share the process-wide dispatcher and native interface allocation.
func init() {
	// xdg_toplevel.configure_bounds was added in version 4. It provides an
	// advisory logical area, without guessing from physical monitor modes.
	for p := range wlProtocols {
		for i := range wlProtocols[p].interfaces {
			iface := &wlProtocols[p].interfaces[i]
			switch iface.name {
			case "xdg_wm_base", "xdg_surface":
				iface.version = 4
			case "xdg_toplevel":
				iface.version = 4
				iface.events = append(iface.events, wlProtocolMessage{"configure_bounds", "4ii", 0})
			}
		}
	}

	wlProtocols = append(wlProtocols,
		wlProtocol{types: []string{"", "wp_fractional_scale_v1", "wl_surface"}, interfaces: []wlProtocolInterface{
			{name: "wp_fractional_scale_manager_v1", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}, {"get_fractional_scale", "no", 1}}},
			{name: "wp_fractional_scale_v1", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}}, events: []wlProtocolMessage{{"preferred_scale", "u", 0}}},
		}},
		wlProtocol{types: []string{"", "", "", "", "wp_viewport", "wl_surface"}, interfaces: []wlProtocolInterface{
			{name: "wp_viewporter", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}, {"get_viewport", "no", 4}}},
			{name: "wp_viewport", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}, {"set_source", "ffff", 0}, {"set_destination", "ii", 0}}},
		}})
}

func bindFractionalGlobal(d *wlDisplay, reg unsafe.Pointer, name uint32, intf string) {
	switch intf {
	case "wp_viewporter":
		d.viewporter = wlRegistryBind(reg, name, wlIface.wp_viewporter, intf, 1)
	case "wp_fractional_scale_manager_v1":
		d.fractionalManager = wlRegistryBind(reg, name, wlIface.wp_fractional_scale_manager_v1, intf, 1)
	}
}

func (w *window) initFractionalScale() {
	d := w.disp
	if d.viewporter == nil || d.fractionalManager == nil {
		return
	}
	w.viewport = marshal(d.viewporter, 1, wlIface.wp_viewport, 0, 0, uintptr(w.surf))
	if w.viewport == nil {
		return
	}
	w.fractional = marshal(d.fractionalManager, 1, wlIface.wp_fractional_scale_v1, 0, 0, uintptr(w.surf))
	if w.fractional == nil {
		marshal(w.viewport, 0, nil, wlMarshalFlagDestroy)
		w.viewport = nil
		return
	}
	wlAddListener(w.fractional, lk_fractional_scale, w.surf)
}

func (w *window) updateViewport() {
	if w.fractionalScale == 0 || w.viewport == nil || w.size.X <= 0 || w.size.Y <= 0 || w.viewportSize == w.size {
		return
	}
	wlSurfaceSetBufferScale(w.surf, 1)
	w.viewportSize = w.size
	marshal(w.viewport, 2, nil, 0, wlArg(uint32(w.size.X)), wlArg(uint32(w.size.Y)))
}

func (w *window) destroyFractionalScale() {
	for _, p := range []unsafe.Pointer{w.fractional, w.viewport} {
		if p != nil {
			marshal(p, 0, nil, wlMarshalFlagDestroy)
		}
	}
	w.fractional, w.viewport = nil, nil
}
func (d *wlDisplay) destroyFractionalScale() {
	for _, p := range []unsafe.Pointer{d.fractionalManager, d.viewporter} {
		if p != nil {
			marshal(p, 0, nil, wlMarshalFlagDestroy)
		}
	}
	d.fractionalManager, d.viewporter = nil, nil
}

func waylandShellVersion(version uint32) uint32 { return min(version, 4) }
