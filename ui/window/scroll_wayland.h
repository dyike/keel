#include <stdint.h>
struct keel_wlscroll;
struct keel_wlscroll *keel_wlscroll_open(void *display, uintptr_t handle);
void keel_wlscroll_poll(struct keel_wlscroll *s);
void keel_wlscroll_close(struct keel_wlscroll *s);
