//go:build linux && !android && cgo && !nowayland

// Reads the Wayland selection through an application's own wl_display, on a
// private event queue. Only the core protocol is used, marshalled by opcode
// from wayland.xml, so no generated protocol header is needed.

#include <errno.h>
#include <fcntl.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <wayland-client-core.h>
#include "wlclip.h"

extern const struct wl_interface wl_registry_interface;
extern const struct wl_interface wl_seat_interface;
extern const struct wl_interface wl_data_device_manager_interface;
extern const struct wl_interface wl_data_device_interface;
extern const struct wl_interface wl_data_offer_interface;

enum {
	DISPLAY_GET_REGISTRY = 1,
	REGISTRY_BIND = 0,
	MANAGER_GET_DATA_DEVICE = 1,
	DEVICE_RELEASE = 2,
	OFFER_RECEIVE = 1,
	OFFER_DESTROY = 2,
};

struct offer {
	struct wl_proxy *proxy;
	char *mimes[KEEL_WLCLIP_MAX_MIMES];
	int n;
	struct offer *next;
};

struct keel_wlclip {
	struct wl_display *display;
	struct wl_event_queue *queue;
	struct wl_proxy *wrapper, *registry, *seat, *manager, *device;
	uint32_t manager_version;
	struct offer *offers;    // every offer announced, owned until close
	struct offer *selection; // the clipboard's, or NULL when empty
	int selection_known;
};

static struct offer *find_offer(struct keel_wlclip *c, struct wl_proxy *p) {
	for (struct offer *o = c->offers; o; o = o->next)
		if (o->proxy == p)
			return o;
	return NULL;
}

static void offer_mime(void *data, struct wl_proxy *proxy, const char *mime) {
	struct offer *o = find_offer(data, proxy);
	if (o && o->n < KEEL_WLCLIP_MAX_MIMES)
		o->mimes[o->n++] = strdup(mime);
}
static void offer_source_actions(void *data, struct wl_proxy *proxy, uint32_t a) {}
static void offer_action(void *data, struct wl_proxy *proxy, uint32_t a) {}
static const struct {
	void (*offer)(void *, struct wl_proxy *, const char *);
	void (*source_actions)(void *, struct wl_proxy *, uint32_t);
	void (*action)(void *, struct wl_proxy *, uint32_t);
} offer_listener = {offer_mime, offer_source_actions, offer_action};

static void device_data_offer(void *data, struct wl_proxy *device, struct wl_proxy *proxy) {
	struct keel_wlclip *c = data;
	struct offer *o = calloc(1, sizeof *o);
	if (!o) {
		wl_proxy_marshal_flags(proxy, OFFER_DESTROY, NULL, wl_proxy_get_version(proxy), WL_MARSHAL_FLAG_DESTROY);
		return;
	}
	o->proxy = proxy;
	o->next = c->offers;
	c->offers = o;
	wl_proxy_add_listener(proxy, (void (**)(void))&offer_listener, c);
}
static void device_enter(void *data, struct wl_proxy *d, uint32_t serial, struct wl_proxy *surface, wl_fixed_t x, wl_fixed_t y, struct wl_proxy *offer) {}
static void device_leave(void *data, struct wl_proxy *d) {}
static void device_motion(void *data, struct wl_proxy *d, uint32_t time, wl_fixed_t x, wl_fixed_t y) {}
static void device_drop(void *data, struct wl_proxy *d) {}
static void device_selection(void *data, struct wl_proxy *d, struct wl_proxy *offer) {
	struct keel_wlclip *c = data;
	c->selection = offer ? find_offer(c, offer) : NULL;
	c->selection_known = 1;
}
static const struct {
	void (*data_offer)(void *, struct wl_proxy *, struct wl_proxy *);
	void (*enter)(void *, struct wl_proxy *, uint32_t, struct wl_proxy *, wl_fixed_t, wl_fixed_t, struct wl_proxy *);
	void (*leave)(void *, struct wl_proxy *);
	void (*motion)(void *, struct wl_proxy *, uint32_t, wl_fixed_t, wl_fixed_t);
	void (*drop)(void *, struct wl_proxy *);
	void (*selection)(void *, struct wl_proxy *, struct wl_proxy *);
} device_listener = {device_data_offer, device_enter, device_leave, device_motion, device_drop, device_selection};

static void registry_global(void *data, struct wl_proxy *registry, uint32_t name, const char *interface, uint32_t version) {
	struct keel_wlclip *c = data;
	if (!c->seat && strcmp(interface, "wl_seat") == 0) {
		// Version 1 is enough to name the seat and needs no listener.
		c->seat = wl_proxy_marshal_flags(registry, REGISTRY_BIND, &wl_seat_interface, 1, 0, name, wl_seat_interface.name, 1, NULL);
	} else if (!c->manager && strcmp(interface, "wl_data_device_manager") == 0) {
		c->manager_version = version < 3 ? version : 3;
		c->manager = wl_proxy_marshal_flags(registry, REGISTRY_BIND, &wl_data_device_manager_interface, c->manager_version, 0, name, wl_data_device_manager_interface.name, c->manager_version, NULL);
	}
}
static void registry_global_remove(void *data, struct wl_proxy *registry, uint32_t name) {}
static const struct {
	void (*global)(void *, struct wl_proxy *, uint32_t, const char *, uint32_t);
	void (*global_remove)(void *, struct wl_proxy *, uint32_t);
} registry_listener = {registry_global, registry_global_remove};

void keel_wlclip_close(struct keel_wlclip *c) {
	if (!c)
		return;
	for (struct offer *o = c->offers; o;) {
		struct offer *next = o->next;
		for (int i = 0; i < o->n; i++)
			free(o->mimes[i]);
		wl_proxy_marshal_flags(o->proxy, OFFER_DESTROY, NULL, wl_proxy_get_version(o->proxy), WL_MARSHAL_FLAG_DESTROY);
		free(o);
		o = next;
	}
	if (c->device) {
		if (c->manager_version >= 2)
			wl_proxy_marshal_flags(c->device, DEVICE_RELEASE, NULL, wl_proxy_get_version(c->device), WL_MARSHAL_FLAG_DESTROY);
		else
			wl_proxy_destroy(c->device);
	}
	if (c->manager)
		wl_proxy_destroy(c->manager);
	if (c->seat)
		wl_proxy_destroy(c->seat);
	if (c->registry)
		wl_proxy_destroy(c->registry);
	if (c->wrapper)
		wl_proxy_wrapper_destroy(c->wrapper);
	wl_display_flush(c->display);
	if (c->queue)
		wl_event_queue_destroy(c->queue);
	free(c);
}

struct keel_wlclip *keel_wlclip_open(void *display, int *err) {
	struct keel_wlclip *c = calloc(1, sizeof *c);
	*err = KEEL_WLCLIP_FAILED;
	if (!c)
		return NULL;
	c->display = display;
	c->queue = wl_display_create_queue(c->display);
	c->wrapper = c->queue ? wl_proxy_create_wrapper(c->display) : NULL;
	if (!c->wrapper)
		goto fail;
	wl_proxy_set_queue(c->wrapper, c->queue);
	c->registry = wl_proxy_marshal_flags(c->wrapper, DISPLAY_GET_REGISTRY, &wl_registry_interface, wl_proxy_get_version(c->wrapper), 0, NULL);
	if (!c->registry)
		goto fail;
	wl_proxy_add_listener(c->registry, (void (**)(void))&registry_listener, c);
	if (wl_display_roundtrip_queue(c->display, c->queue) < 0)
		goto fail;
	if (!c->seat || !c->manager) {
		*err = KEEL_WLCLIP_UNSUPPORTED;
		goto fail;
	}
	c->device = wl_proxy_marshal_flags(c->manager, MANAGER_GET_DATA_DEVICE, &wl_data_device_interface, wl_proxy_get_version(c->manager), 0, NULL, c->seat);
	if (!c->device)
		goto fail;
	wl_proxy_add_listener(c->device, (void (**)(void))&device_listener, c);
	// The compositor announces the selection to a focused client's new data
	// device; a second roundtrip lets the offer's MIME types arrive too.
	if (wl_display_roundtrip_queue(c->display, c->queue) < 0 || wl_display_roundtrip_queue(c->display, c->queue) < 0)
		goto fail;
	*err = 0;
	return c;
fail:
	keel_wlclip_close(c);
	return NULL;
}

int keel_wlclip_count(struct keel_wlclip *c) { return c->selection ? c->selection->n : 0; }
const char *keel_wlclip_mime(struct keel_wlclip *c, int i) { return c->selection->mimes[i]; }
int keel_wlclip_known(struct keel_wlclip *c) { return c->selection_known; }

// keel_wlclip_receive asks the selection's owner to write mime into a pipe
// and returns its non-blocking read end, or -1.
int keel_wlclip_receive(struct keel_wlclip *c, const char *mime) {
	int fds[2];
	if (!c->selection || pipe2(fds, O_CLOEXEC) < 0)
		return -1;
	wl_proxy_marshal_flags(c->selection->proxy, OFFER_RECEIVE, NULL, wl_proxy_get_version(c->selection->proxy), 0, mime, fds[1]);
	close(fds[1]);
	if (wl_display_flush(c->display) < 0 && errno != EAGAIN) {
		close(fds[0]);
		return -1;
	}
	fcntl(fds[0], F_SETFL, fcntl(fds[0], F_GETFL) | O_NONBLOCK);
	return fds[0];
}
