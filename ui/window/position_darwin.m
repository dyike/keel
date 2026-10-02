//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>
#include <stdint.h>

void keel_center_window(uintptr_t handle) {
    NSView *view = (NSView *)CFRetain((CFTypeRef)handle);
    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *window = view.window;
        NSScreen *screen = window.screen ?: NSScreen.mainScreen;
        if (window && screen && !window.miniaturized &&
            !(window.styleMask & NSWindowStyleMaskFullScreen)) {
            NSRect visible = screen.visibleFrame;
            NSRect frame = window.frame;
            // Include the title bar, preserve the size and the screen origin.
            // Oversized windows keep their top and left edges reachable.
            NSPoint origin = NSMakePoint(
                NSMinX(visible) + MAX(0, (NSWidth(visible) - NSWidth(frame)) / 2),
                NSMaxY(visible) - NSHeight(frame) - MAX(0, (NSHeight(visible) - NSHeight(frame)) / 2));
            [window setFrameOrigin:origin];
        }
        CFRelease((CFTypeRef)view);
    });
}
