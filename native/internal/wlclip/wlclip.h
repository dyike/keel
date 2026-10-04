#define KEEL_WLCLIP_MAX_MIMES 64
#define KEEL_WLCLIP_FAILED 1
#define KEEL_WLCLIP_UNSUPPORTED 2

struct keel_wlclip;
struct keel_wlclip *keel_wlclip_open(void *display, int *err);
void keel_wlclip_close(struct keel_wlclip *c);
int keel_wlclip_known(struct keel_wlclip *c);
int keel_wlclip_count(struct keel_wlclip *c);
const char *keel_wlclip_mime(struct keel_wlclip *c, int i);
int keel_wlclip_receive(struct keel_wlclip *c, const char *mime);
