//go:build linux && !android && !nowayland

// xdg-activation-v1 from wayland-protocols (staging), written out by hand:
// Activate needs one request on Gio's own connection and surface.

#include <string.h>
#include <wayland-client-core.h>
#include "activation_wayland.h"

extern const struct wl_interface wl_registry_interface;
extern const struct wl_interface wl_surface_interface;
extern const struct wl_interface wl_seat_interface;

static const struct wl_interface xdg_activation_token_v1_interface;

static const struct wl_interface *activation_types[] = {
	&xdg_activation_token_v1_interface, // get_activation_token: id
	NULL, &wl_surface_interface,        // activate: token, surface
	NULL, &wl_seat_interface,           // token set_serial: serial, seat
	&wl_surface_interface,              // token set_surface: surface
};

static const struct wl_message token_requests[] = {
	{"set_serial", "uo", activation_types + 3},
	{"set_app_id", "s", activation_types + 1},
	{"set_surface", "o", activation_types + 5},
	{"commit", "", activation_types + 1},
	{"destroy", "", activation_types + 1},
};
static const struct wl_message token_events[] = {
	{"done", "s", activation_types + 1},
};
static const struct wl_interface xdg_activation_token_v1_interface = {
	"xdg_activation_token_v1", 1, 5, token_requests, 1, token_events,
};

static const struct wl_message activation_requests[] = {
	{"destroy", "", activation_types + 1},
	{"get_activation_token", "n", activation_types + 0},
	{"activate", "so", activation_types + 1},
};
static const struct wl_interface xdg_activation_v1_interface = {
	"xdg_activation_v1", 1, 3, activation_requests, 0, NULL,
};

enum { DISPLAY_GET_REGISTRY = 1, REGISTRY_BIND = 0, ACTIVATION_DESTROY = 0, ACTIVATION_ACTIVATE = 2 };

struct lookup { struct wl_proxy *registry, *activation; };

static void global(void *data, struct wl_proxy *registry, uint32_t name, const char *interface, uint32_t version) {
	struct lookup *l = data;
	if (!l->activation && strcmp(interface, xdg_activation_v1_interface.name) == 0)
		l->activation = wl_proxy_marshal_flags(registry, REGISTRY_BIND, &xdg_activation_v1_interface, 1, 0, name, xdg_activation_v1_interface.name, 1, NULL);
}
static void global_remove(void *data, struct wl_proxy *registry, uint32_t name) {}
static const struct {
	void (*global)(void *, struct wl_proxy *, uint32_t, const char *, uint32_t);
	void (*global_remove)(void *, struct wl_proxy *, uint32_t);
} registry_listener = {global, global_remove};

// keel_wayland_activate asks the compositor to activate surface with token.
// It returns 0 on success, 1 when the compositor lacks xdg-activation, and
// -1 on a connection error.
int keel_wayland_activate(void *display, void *surface, const char *token) {
	struct wl_display *d = display;
	struct wl_event_queue *queue = wl_display_create_queue(d);
	if (!queue)
		return -1;
	struct wl_proxy *wrapper = wl_proxy_create_wrapper(d);
	if (!wrapper) {
		wl_event_queue_destroy(queue);
		return -1;
	}
	wl_proxy_set_queue(wrapper, queue);
	struct lookup l = {0};
	int ret = -1;
	l.registry = wl_proxy_marshal_flags(wrapper, DISPLAY_GET_REGISTRY, &wl_registry_interface, wl_proxy_get_version(wrapper), 0, NULL);
	if (!l.registry)
		goto done;
	wl_proxy_add_listener(l.registry, (void (**)(void))&registry_listener, &l);
	if (wl_display_roundtrip_queue(d, queue) < 0)
		goto done;
	if (!l.activation) {
		ret = 1;
		goto done;
	}
	wl_proxy_marshal_flags(l.activation, ACTIVATION_ACTIVATE, NULL, 1, 0, token, surface);
	wl_proxy_marshal_flags(l.activation, ACTIVATION_DESTROY, NULL, 1, WL_MARSHAL_FLAG_DESTROY);
	ret = wl_display_flush(d) < 0 ? -1 : 0;
done:
	if (l.registry)
		wl_proxy_destroy(l.registry);
	wl_proxy_wrapper_destroy(wrapper);
	wl_event_queue_destroy(queue);
	return ret;
}
