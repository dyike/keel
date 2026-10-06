//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>
#include <stdint.h>
#import <objc/runtime.h>
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

// AppKit lays out close/minimize again when it updates the window. Apply the
// requested geometry after that pass, and after resize/fullscreen transitions.
@interface KeelTrafficLightLayout : NSObject
@property(assign) NSView *view;
@property(assign) CGFloat height, left, offsetY, spacing, systemSpacing, minimumHeight;
- (void)apply;
@end
@implementation KeelTrafficLightLayout
- (void)dealloc {
    [NSNotificationCenter.defaultCenter removeObserver:self];
    [super dealloc];
}
- (void)updated:(NSNotification *)notification {
    [self apply];
}
- (void)apply {
    NSView *view = self.view;
    NSWindow *window = view.window;
    if (!window || (window.styleMask & NSWindowStyleMaskFullScreen)) return;
    NSButton *close = [window standardWindowButton:NSWindowCloseButton];
    if (self.height > 0) {
        NSView *bar = close.superview;
        NSView *container = bar.superview;
        CGFloat barHeight = MAX(self.height, self.minimumHeight);
        NSRect frame = container.frame;
        frame.origin.y = NSMaxY(frame) - barHeight;
        frame.size.height = barHeight;
        if (!NSEqualRects(container.frame,frame)) container.frame = frame;
        frame = bar.frame;
        frame.origin.y = 0;
        frame.size.height = barHeight;
        if (!NSEqualRects(bar.frame,frame)) bar.frame = frame;
    }
    for (NSUInteger kind = NSWindowCloseButton; kind <= NSWindowZoomButton; kind++) {
        NSButton *button = [window standardWindowButton:(NSWindowButton)kind];
        button.hidden = NO;
        if (self.height > 0) {
            CGFloat gap = self.spacing > 0 ? self.spacing : self.systemSpacing;
            NSRect frame = button.frame;
            NSPoint center = NSMakePoint(self.left+NSWidth(frame)/2+kind*gap,
                                        self.height/2+self.offsetY);
            if (!view.isFlipped) center.y = NSHeight(view.bounds)-center.y;
            center = [button.superview convertPoint:center fromView:view];
            frame.origin = NSMakePoint(center.x-NSWidth(frame)/2,center.y-NSHeight(frame)/2);
            if (!NSEqualRects(button.frame,frame)) button.frame = frame;
        }
    }
}
@end

// Reuse NSWindow's own buttons in their existing hierarchy: no replacement
// cells, tracking areas, actions or accessibility implementation.
void keel_native_traffic_lights(uintptr_t handle, int custom, double height, double left, double offsetY, double spacing) {
    NSView *view = (NSView *)CFRetain((CFTypeRef)handle);
    dispatch_async(dispatch_get_main_queue(), ^{
        NSWindow *window = view.window;
        if (window) {
            window.titleVisibility = NSWindowTitleHidden;
            window.titlebarAppearsTransparent = YES;
            if (@available(macOS 11.0, *)) window.titlebarSeparatorStyle = NSTitlebarSeparatorStyleNone;
            static char layoutKey;
            KeelTrafficLightLayout *layout = objc_getAssociatedObject(view,&layoutKey);
            if (!layout) {
                layout = [[KeelTrafficLightLayout alloc] init];
                layout.view = view;
                NSButton *close = [window standardWindowButton:NSWindowCloseButton];
                NSButton *mini = [window standardWindowButton:NSWindowMiniaturizeButton];
                layout.systemSpacing = mini.frame.origin.x-close.frame.origin.x;
                layout.minimumHeight = NSHeight(close.superview.frame);
                objc_setAssociatedObject(view,&layoutKey,layout,OBJC_ASSOCIATION_RETAIN_NONATOMIC);
                for (NSString *name in @[NSWindowDidUpdateNotification, NSWindowDidResizeNotification, NSWindowDidExitFullScreenNotification]) {
                    [NSNotificationCenter.defaultCenter addObserver:layout selector:@selector(updated:) name:name object:window];
                }
                [layout release];
            }
            layout.height = custom ? height : 0;
            layout.left = left;
            layout.offsetY = offsetY;
            layout.spacing = spacing;
            [layout apply];
        }
        CFRelease((CFTypeRef)view);
    });
}
