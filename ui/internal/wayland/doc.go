// Package wayland calls libwayland-client through purego, without cgo, on
// the application's own Wayland connection (the one Gio opened). It is
// ui/window's way to reach what Gio does not expose: scroll sources and
// xdg-activation. libwayland 1.20+ is loaded at run time; build with
// -tags nowayland to leave it out.
//
// Requests use wl_proxy_marshal_array_flags (no variadic calls); events of
// every proxy go through one dispatcher (wl_proxy_add_dispatcher) created
// once, routed by a token kept as the proxy's user data. Work happens on
// private event queues, so Gio's own event handling is untouched.
package wayland
