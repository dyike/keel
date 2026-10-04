//go:build linux && !android && !nowayland

// A second wl_pointer on Gio's connection, on a private queue, to learn what
// Gio drops: whether scrolling comes from a wheel or fingers (axis_source)
// and when fingers lift (axis_stop). Needs wl_seat version 5.

#include <stdlib.h>
#include <string.h>
#include <wayland-client-core.h>
#include "scroll_wayland.h"

extern const struct wl_interface wl_registry_interface;
extern const struct wl_interface wl_seat_interface;
extern const struct wl_interface wl_pointer_interface;
extern void keel_wl_scroll_frame(uintptr_t handle, uint32_t source, int scrolled, int stopped);

enum {
	DISPLAY_GET_REGISTRY = 1,
	REGISTRY_BIND = 0,
	SEAT_GET_POINTER = 0,
	SEAT_RELEASE = 3,
	POINTER_RELEASE = 1,
	SOURCE_NONE = 0xffffffff, // wlAxisNone in Go
};

struct keel_wlscroll {
	struct wl_display *display;
	struct wl_event_queue *queue;
	struct wl_proxy *wrapper, *registry, *seat, *pointer;
	uintptr_t handle;
	uint32_t source; // axis_source of the current frame
	int scrolled, stopped;
};

static void p_enter(void *d, struct wl_proxy *p, uint32_t serial, struct wl_proxy *surface, wl_fixed_t x, wl_fixed_t y) {}
static void p_leave(void *d, struct wl_proxy *p, uint32_t serial, struct wl_proxy *surface) {}
static void p_motion(void *d, struct wl_proxy *p, uint32_t time, wl_fixed_t x, wl_fixed_t y) {}
static void p_button(void *d, struct wl_proxy *p, uint32_t serial, uint32_t time, uint32_t button, uint32_t state) {}
static void p_axis(void *d, struct wl_proxy *p, uint32_t time, uint32_t axis, wl_fixed_t value) {
	((struct keel_wlscroll *)d)->scrolled = 1;
}
static void p_frame(void *d, struct wl_proxy *p) {
	struct keel_wlscroll *s = d;
	if (s->scrolled || s->stopped)
		keel_wl_scroll_frame(s->handle, s->source, s->scrolled, s->stopped);
	s->source = SOURCE_NONE;
	s->scrolled = s->stopped = 0;
}
static void p_axis_source(void *d, struct wl_proxy *p, uint32_t source) { ((struct keel_wlscroll *)d)->source = source; }
static void p_axis_stop(void *d, struct wl_proxy *p, uint32_t time, uint32_t axis) { ((struct keel_wlscroll *)d)->stopped = 1; }
static void p_axis_discrete(void *d, struct wl_proxy *p, uint32_t axis, int32_t discrete) {}
static const struct {
	void (*enter)(void *, struct wl_proxy *, uint32_t, struct wl_proxy *, wl_fixed_t, wl_fixed_t);
	void (*leave)(void *, struct wl_proxy *, uint32_t, struct wl_proxy *);
	void (*motion)(void *, struct wl_proxy *, uint32_t, wl_fixed_t, wl_fixed_t);
	void (*button)(void *, struct wl_proxy *, uint32_t, uint32_t, uint32_t, uint32_t);
	void (*axis)(void *, struct wl_proxy *, uint32_t, uint32_t, wl_fixed_t);
	void (*frame)(void *, struct wl_proxy *);
	void (*axis_source)(void *, struct wl_proxy *, uint32_t);
	void (*axis_stop)(void *, struct wl_proxy *, uint32_t, uint32_t);
	void (*axis_discrete)(void *, struct wl_proxy *, uint32_t, int32_t);
} pointer_listener = {p_enter, p_leave, p_motion, p_button, p_axis, p_frame, p_axis_source, p_axis_stop, p_axis_discrete};

static void seat_capabilities(void *d, struct wl_proxy *seat, uint32_t caps) {
	struct keel_wlscroll *s = d;
	if ((caps & 1) && !s->pointer) { // WL_SEAT_CAPABILITY_POINTER
		s->pointer = wl_proxy_marshal_flags(seat, SEAT_GET_POINTER, &wl_pointer_interface, wl_proxy_get_version(seat), 0, NULL);
		if (s->pointer)
			wl_proxy_add_listener(s->pointer, (void (**)(void))&pointer_listener, s);
	}
}
static void seat_name(void *d, struct wl_proxy *seat, const char *name) {}
static const struct {
	void (*capabilities)(void *, struct wl_proxy *, uint32_t);
	void (*name)(void *, struct wl_proxy *, const char *);
} seat_listener = {seat_capabilities, seat_name};

static void global(void *d, struct wl_proxy *registry, uint32_t name, const char *interface, uint32_t version) {
	struct keel_wlscroll *s = d;
	if (!s->seat && strcmp(interface, "wl_seat") == 0 && version >= 5) {
		s->seat = wl_proxy_marshal_flags(registry, REGISTRY_BIND, &wl_seat_interface, 5, 0, name, wl_seat_interface.name, 5, NULL);
		if (s->seat)
			wl_proxy_add_listener(s->seat, (void (**)(void))&seat_listener, s);
	}
}
static void global_remove(void *d, struct wl_proxy *registry, uint32_t name) {}
static const struct {
	void (*global)(void *, struct wl_proxy *, uint32_t, const char *, uint32_t);
	void (*global_remove)(void *, struct wl_proxy *, uint32_t);
} registry_listener = {global, global_remove};

void keel_wlscroll_close(struct keel_wlscroll *s) {
	if (!s)
		return;
	if (s->pointer)
		wl_proxy_marshal_flags(s->pointer, POINTER_RELEASE, NULL, wl_proxy_get_version(s->pointer), WL_MARSHAL_FLAG_DESTROY);
	if (s->seat)
		wl_proxy_marshal_flags(s->seat, SEAT_RELEASE, NULL, wl_proxy_get_version(s->seat), WL_MARSHAL_FLAG_DESTROY);
	if (s->registry)
		wl_proxy_destroy(s->registry);
	if (s->wrapper)
		wl_proxy_wrapper_destroy(s->wrapper);
	wl_display_flush(s->display);
	if (s->queue)
		wl_event_queue_destroy(s->queue);
	free(s);
}

// keel_wlscroll_open binds a seat and its pointer on display. Events queue
// up as Gio reads the socket; keel_wlscroll_poll handles them.
struct keel_wlscroll *keel_wlscroll_open(void *display, uintptr_t handle) {
	struct keel_wlscroll *s = calloc(1, sizeof *s);
	if (!s)
		return NULL;
	s->display = display;
	s->handle = handle;
	s->source = SOURCE_NONE;
	s->queue = wl_display_create_queue(s->display);
	s->wrapper = s->queue ? wl_proxy_create_wrapper(s->display) : NULL;
	if (!s->wrapper)
		goto fail;
	wl_proxy_set_queue(s->wrapper, s->queue);
	s->registry = wl_proxy_marshal_flags(s->wrapper, DISPLAY_GET_REGISTRY, &wl_registry_interface, wl_proxy_get_version(s->wrapper), 0, NULL);
	if (!s->registry)
		goto fail;
	wl_proxy_add_listener(s->registry, (void (**)(void))&registry_listener, s);
	// One roundtrip finds the seat, the next its capabilities.
	if (wl_display_roundtrip_queue(s->display, s->queue) < 0 || !s->seat || wl_display_roundtrip_queue(s->display, s->queue) < 0 || !s->pointer)
		goto fail;
	return s;
fail:
	keel_wlscroll_close(s);
	return NULL;
}

void keel_wlscroll_poll(struct keel_wlscroll *s) { wl_display_dispatch_queue_pending(s->display, s->queue); }
