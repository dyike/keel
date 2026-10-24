// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nowayland && (!cgo || purego)

package app

// Code generated from wayland_xdg_shell.c, wayland_xdg_decoration.c and
// wayland_text_input.c (wayland-scanner's private code); DO NOT EDIT.
// os_wayland_purego.go builds these interfaces in C memory. Keel patch.

// wlProtocol is one protocol's shared types array and its interfaces.
type wlProtocol struct {
	types      []string // interface names, "" for NULL
	interfaces []wlProtocolInterface
}

type wlProtocolInterface struct {
	name             string
	version          int32
	requests, events []wlProtocolMessage
}

// wlProtocolMessage is a wl_message; types indexes the protocol's types.
type wlProtocolMessage struct {
	name, signature string
	types           int
}

var wlProtocols = []wlProtocol{
	{
		types: []string{"", "", "", "", "xdg_positioner", "xdg_surface", "wl_surface", "xdg_toplevel", "xdg_popup", "xdg_surface", "xdg_positioner", "xdg_toplevel", "wl_seat", "", "", "", "wl_seat", "", "wl_seat", "", "", "wl_output", "wl_seat", "", "xdg_positioner", ""},
		interfaces: []wlProtocolInterface{
			{name: "xdg_wm_base", version: 3, requests: []wlProtocolMessage{{"destroy", "", 0}, {"create_positioner", "n", 4}, {"get_xdg_surface", "no", 5}, {"pong", "u", 0}}, events: []wlProtocolMessage{{"ping", "u", 0}}},
			{name: "xdg_positioner", version: 3, requests: []wlProtocolMessage{{"destroy", "", 0}, {"set_size", "ii", 0}, {"set_anchor_rect", "iiii", 0}, {"set_anchor", "u", 0}, {"set_gravity", "u", 0}, {"set_constraint_adjustment", "u", 0}, {"set_offset", "ii", 0}, {"set_reactive", "3", 0}, {"set_parent_size", "3ii", 0}, {"set_parent_configure", "3u", 0}}, events: nil},
			{name: "xdg_surface", version: 3, requests: []wlProtocolMessage{{"destroy", "", 0}, {"get_toplevel", "n", 7}, {"get_popup", "n?oo", 8}, {"set_window_geometry", "iiii", 0}, {"ack_configure", "u", 0}}, events: []wlProtocolMessage{{"configure", "u", 0}}},
			{name: "xdg_toplevel", version: 3, requests: []wlProtocolMessage{{"destroy", "", 0}, {"set_parent", "?o", 11}, {"set_title", "s", 0}, {"set_app_id", "s", 0}, {"show_window_menu", "ouii", 12}, {"move", "ou", 16}, {"resize", "ouu", 18}, {"set_max_size", "ii", 0}, {"set_min_size", "ii", 0}, {"set_maximized", "", 0}, {"unset_maximized", "", 0}, {"set_fullscreen", "?o", 21}, {"unset_fullscreen", "", 0}, {"set_minimized", "", 0}}, events: []wlProtocolMessage{{"configure", "iia", 0}, {"close", "", 0}}},
			{name: "xdg_popup", version: 3, requests: []wlProtocolMessage{{"destroy", "", 0}, {"grab", "ou", 22}, {"reposition", "3ou", 24}}, events: []wlProtocolMessage{{"configure", "iiii", 0}, {"popup_done", "", 0}, {"repositioned", "3u", 0}}},
		},
	},
	{
		types: []string{"", "zxdg_toplevel_decoration_v1", "xdg_toplevel"},
		interfaces: []wlProtocolInterface{
			{name: "zxdg_decoration_manager_v1", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}, {"get_toplevel_decoration", "no", 1}}, events: nil},
			{name: "zxdg_toplevel_decoration_v1", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}, {"set_mode", "u", 0}, {"unset_mode", "", 0}}, events: []wlProtocolMessage{{"configure", "u", 0}}},
		},
	},
	{
		types: []string{"", "", "", "", "wl_surface", "wl_surface", "zwp_text_input_v3", "wl_seat"},
		interfaces: []wlProtocolInterface{
			{name: "zwp_text_input_v3", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}, {"enable", "", 0}, {"disable", "", 0}, {"set_surrounding_text", "sii", 0}, {"set_text_change_cause", "u", 0}, {"set_content_type", "uu", 0}, {"set_cursor_rectangle", "iiii", 0}, {"commit", "", 0}}, events: []wlProtocolMessage{{"enter", "o", 4}, {"leave", "o", 5}, {"preedit_string", "?sii", 0}, {"commit_string", "?s", 0}, {"delete_surrounding_text", "uu", 0}, {"done", "u", 0}}},
			{name: "zwp_text_input_manager_v3", version: 1, requests: []wlProtocolMessage{{"destroy", "", 0}, {"get_text_input", "no", 6}}, events: nil},
		},
	},
}
