//go:build darwin && !ios && cgo && !nometal

#import <AppKit/AppKit.h>
#import <Metal/Metal.h>
#import <QuartzCore/CAMetalLayer.h>
#include <stdatomic.h>
#include "glass_darwin.h"

struct keel_glass_context {
    atomic_int ready;
    NSView *view, *backdrop;
    CAMetalLayer *layer;
    id<MTLDevice> device;
    id<MTLCommandQueue> queue;
};

// All input stays with GioView, including clicks on empty glass. Native
// material views are only used for compositing, never as input responders.
@interface KeelFrostedBackdrop : NSVisualEffectView
@end
@implementation KeelFrostedBackdrop
- (NSView *)hitTest:(NSPoint)point { return nil; }
@end
#if __MAC_OS_X_VERSION_MAX_ALLOWED >= 260000
API_AVAILABLE(macos(26.0))
@interface KeelLiquidBackdrop : NSGlassEffectView
@end
@implementation KeelLiquidBackdrop
- (NSView *)hitTest:(NSPoint)point { return nil; }
@end
#endif

int keel_glass_supported(void) { return NSClassFromString(@"CAMetalLayer") != Nil; }
int keel_liquid_glass_supported(void) {
#if __MAC_OS_X_VERSION_MAX_ALLOWED >= 260000
    if (@available(macOS 26.0, *)) return NSClassFromString(@"NSGlassEffectView") != Nil;
#endif
    return 0;
}

keel_glass_context *keel_glass_create(uintptr_t handle, int style, double radius) {
    keel_glass_context *ctx = calloc(1, sizeof(*ctx));
    if (!ctx) return NULL;
    atomic_init(&ctx->ready, 0);
    ctx->view = (NSView *)CFRetain((CFTypeRef)handle);
    dispatch_async(dispatch_get_main_queue(), ^{
        @autoreleasepool {
            NSView *view = ctx->view;
            NSWindow *window = view.window;
            if (!window || ![view.layer isKindOfClass:CAMetalLayer.class]) {
                atomic_store_explicit(&ctx->ready, -1, memory_order_release);
                return;
            }
            ctx->device = MTLCreateSystemDefaultDevice();
            ctx->queue = [ctx->device newCommandQueue];
            ctx->layer = [[CAMetalLayer alloc] init];
            if (!ctx->device || !ctx->queue) {
                atomic_store_explicit(&ctx->ready, -1, memory_order_release);
                return;
            }
            ctx->layer.device = ctx->device;
            ctx->layer.pixelFormat = MTLPixelFormatBGRA8Unorm_sRGB;
            ctx->layer.opaque = NO;
            ctx->layer.presentsWithTransaction = NO;
            ctx->layer.contentsScale = window.backingScaleFactor;
            view.layer.opaque = NO;
            window.opaque = NO;
            window.backgroundColor = NSColor.clearColor;

            NSView *content = [[NSView alloc] initWithFrame:view.bounds];
            content.wantsLayer = YES;
            content.layer = ctx->layer;
            content.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;

            // Behind-window vibrancy supplies the desktop sampling surface.
            // Liquid Glass then renders its optical treatment above that
            // surface, with Gio inside its documented contentView slot.
            KeelFrostedBackdrop *backdrop = [[KeelFrostedBackdrop alloc] initWithFrame:view.bounds];
            backdrop.material = NSVisualEffectMaterialSidebar;
            backdrop.blendingMode = NSVisualEffectBlendingModeBehindWindow;
            backdrop.state = NSVisualEffectStateFollowsWindowActiveState;
            backdrop.wantsLayer = YES;
            if (radius > 0) {
                backdrop.layer.cornerRadius = radius;
                backdrop.layer.masksToBounds = YES;
            }
            BOOL liquid = NO;
#if __MAC_OS_X_VERSION_MAX_ALLOWED >= 260000
            if (@available(macOS 26.0, *)) {
                if (style != 2) {
                    NSGlassEffectView *glass = [[KeelLiquidBackdrop alloc] initWithFrame:view.bounds];
                    glass.style = style == 1 ? NSGlassEffectViewStyleClear : NSGlassEffectViewStyleRegular;
                    if (radius > 0) glass.cornerRadius = radius;
                    glass.contentView = content;
                    glass.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
                    [backdrop addSubview:glass];
                    [glass release];
                    liquid = YES;
                }
            }
#endif
            if (!liquid) [backdrop addSubview:content];
            [content release];
            backdrop.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
            // Keep GioView as NSWindow.contentView: Gio's delegates rely on
            // that identity for input, focus, close and fullscreen events.
            [view addSubview:backdrop];
            ctx->backdrop = backdrop;
            atomic_store_explicit(&ctx->ready, 1, memory_order_release);
        }
    });
    return ctx;
}

int keel_glass_ready(keel_glass_context *ctx, keel_glass_api *api) {
    int ready = atomic_load_explicit(&ctx->ready, memory_order_acquire);
    if (ready == 1) {
        api->device = (uintptr_t)ctx->device;
        api->queue = (uintptr_t)ctx->queue;
        api->pixel_format = (int)MTLPixelFormatBGRA8Unorm_sRGB;
    }
    return ready;
}

uintptr_t keel_glass_drawable(keel_glass_context *ctx, int width, int height) {
    @autoreleasepool {
        ctx->layer.drawableSize = CGSizeMake(width, height);
        return (uintptr_t)[[ctx->layer nextDrawable] retain];
    }
}
uintptr_t keel_glass_texture(uintptr_t drawable) {
    return (uintptr_t)[(id<CAMetalDrawable>)drawable texture];
}
void keel_glass_present(keel_glass_context *ctx, uintptr_t handle, int present) {
    @autoreleasepool {
        id<CAMetalDrawable> drawable = (id<CAMetalDrawable>)handle;
        if (present) {
            id<MTLCommandBuffer> buffer = [ctx->queue commandBuffer];
            [buffer commit];
            [buffer waitUntilScheduled];
            [drawable present];
        }
        [drawable release];
    }
}
void keel_glass_release(keel_glass_context *ctx) {
    // Enqueued after installation, including when the window closes before
    // installation finishes. Never wait for AppKit with the frame lock held.
    dispatch_async(dispatch_get_main_queue(), ^{
        [ctx->backdrop removeFromSuperview];
        [ctx->backdrop release];
        [ctx->layer release];
        [ctx->queue release];
        [ctx->device release];
        [ctx->view release];
        free(ctx);
    });
}
