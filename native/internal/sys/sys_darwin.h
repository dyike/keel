#include <stdint.h>
#include <stddef.h>
typedef struct { uint32_t id; double x,y,w,h; size_t pw,ph; int primary; } keel_display;
int keel_permission(int kind, int request);
int keel_displays(keel_display **out, uint32_t *count);
int keel_capture(uint32_t id, void **data, size_t *length);
int keel_position(double *x, double *y);
int keel_move(double x,double y);
int keel_click(int button);
int keel_key(unsigned short key,int down);
int keel_hotkey_register(uint32_t id,unsigned short key,unsigned int modifiers,void **ref);
int keel_hotkey_unregister(void *ref);
