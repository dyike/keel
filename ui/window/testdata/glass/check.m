//go:build darwin && !ios && cgo && !nometal

#import <AppKit/AppKit.h>
#import <QuartzCore/CAMetalLayer.h>

int check_glass(int liquid, int width, int height) {
    __block int ok = 0;
    dispatch_sync(dispatch_get_main_queue(), ^{
        for (NSWindow *window in NSApp.windows) {
            if (![window.title isEqualToString:@"glass acceptance"] || !window.isVisible || window.isMiniaturized) continue;
            NSView *gio = window.contentView;
            if (![NSStringFromClass(gio.class) isEqualToString:@"GioView"]) continue;
            if (window.isOpaque || gio.layer.isOpaque || window.backgroundColor.alphaComponent != 0) continue;
            if (fabs(NSWidth(gio.bounds) - width) > 1 || fabs(NSHeight(gio.bounds) - height) > 1) continue;
            NSVisualEffectView *base = nil;
            for (NSView *child in gio.subviews) if ([child isKindOfClass:NSVisualEffectView.class]) base = (NSVisualEffectView *)child;
            if (!base || base.blendingMode != NSVisualEffectBlendingModeBehindWindow) continue;
            NSView *host = nil;
            NSView *glass = nil;
            for (NSView *child in base.subviews) {
                if ([child isKindOfClass:NSClassFromString(@"NSGlassEffectView")]) glass = child;
                else if ([child.layer isKindOfClass:CAMetalLayer.class]) host = child;
            }
            if (liquid) {
                if (!glass) continue;
                host = [glass valueForKey:@"contentView"];
                if (!NSEqualSizes(glass.bounds.size, gio.bounds.size)) continue;
            } else if (glass) continue;
            if (!host || ![host.layer isKindOfClass:CAMetalLayer.class]) continue;
            CAMetalLayer *layer = (CAMetalLayer *)host.layer;
            if (layer.isOpaque || !layer.device || !NSEqualSizes(host.bounds.size, gio.bounds.size)) continue;
            CGFloat scale = window.backingScaleFactor;
            if (fabs(layer.drawableSize.width - width * scale) > 2 || fabs(layer.drawableSize.height - height * scale) > 2) continue;
            // The native effect must not consume Gio input events.
            NSPoint point = [gio convertPoint:NSMakePoint(30, 80) toView:gio.superview];
            if ([gio hitTest:point] != gio) continue;
            ok = 1;
        }
    });
    return ok;
}
