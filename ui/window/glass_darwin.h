#include <stdint.h>

typedef struct keel_glass_context keel_glass_context;
typedef struct {
    uintptr_t device, queue;
    int pixel_format;
} keel_glass_api;

int keel_glass_supported(void);
int keel_liquid_glass_supported(void);
keel_glass_context *keel_glass_create(uintptr_t view, int style, double radius);
int keel_glass_ready(keel_glass_context *context, keel_glass_api *api);
uintptr_t keel_glass_drawable(keel_glass_context *context, int width, int height);
uintptr_t keel_glass_texture(uintptr_t drawable);
void keel_glass_present(keel_glass_context *context, uintptr_t drawable, int present);
void keel_glass_release(keel_glass_context *context);
