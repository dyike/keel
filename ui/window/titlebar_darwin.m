//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>
#include <stdint.h>
extern void keel_titlebar_double_click(uintptr_t view, int action);

void keel_titlebar_area(uintptr_t handle, double x, double y, double width, double height) {
    // Gio keeps the handle valid until the next view event. Retain it before
    // asynchronous dispatch so frame code never waits for the main thread.
    NSView *view = (NSView *)CFRetain((CFTypeRef)handle);
    dispatch_async(dispatch_get_main_queue(), ^{
        static NSMapTable<NSView *, NSValue *> *regions;
        static id monitor;
        if (!regions) {
            regions = [NSMapTable strongToStrongObjectsMapTable];
            [regions retain];
            monitor = [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskLeftMouseDown
                handler:^NSEvent *(NSEvent *event) {
                    if (event.clickCount != 2) return event;
                    for (NSView *candidate in regions.keyEnumerator) {
                        if (candidate.window != event.window || !candidate.window) continue;
                        NSPoint p = [candidate convertPoint:event.locationInWindow fromView:nil];
                        if (!candidate.isFlipped) p.y = candidate.bounds.size.height - p.y;
                        if (!NSPointInRect(p, [regions objectForKey:candidate].rectValue)) continue;
                        NSString *action = [NSUserDefaults.standardUserDefaults stringForKey:@"AppleActionOnDoubleClick"];
                        int command = [action isEqualToString:@"None"] ? 0 : ([action isEqualToString:@"Minimize"] ? 1 : 2);
                        if (command) keel_titlebar_double_click((uintptr_t)candidate, command);
                        return nil; // Do not let Gio start another native window drag.
                    }
                    return event;
                }];
        }
        if (width > 0 && height > 0) [regions setObject:[NSValue valueWithRect:NSMakeRect(x,y,width,height)] forKey:view];
        else [regions removeObjectForKey:view];
        CFRelease((CFTypeRef)view);
    });
}
