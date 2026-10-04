//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>
extern void keel_motion_changed(int reduce);

void keel_watch_motion(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSWorkspace *workspace = NSWorkspace.sharedWorkspace;
        static id observer;
        observer = [workspace.notificationCenter
            addObserverForName:NSWorkspaceAccessibilityDisplayOptionsDidChangeNotification
            object:workspace queue:NSOperationQueue.mainQueue
            usingBlock:^(NSNotification *notification) {
                keel_motion_changed(workspace.accessibilityDisplayShouldReduceMotion);
            }];
        keel_motion_changed(workspace.accessibilityDisplayShouldReduceMotion);
    });
}

extern void keel_scrollers_changed(int overlay);

int keel_overlay_scrollers(void) { return NSScroller.preferredScrollerStyle == NSScrollerStyleOverlay; }

// Overlay scrollers ("Show scroll bars: automatically / when scrolling")
// hide at rest; legacy ones ("always") stay.
void keel_watch_scrollers(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        static id observer;
        observer = [NSNotificationCenter.defaultCenter
            addObserverForName:NSPreferredScrollerStyleDidChangeNotification
            object:nil queue:NSOperationQueue.mainQueue
            usingBlock:^(NSNotification *notification) {
                keel_scrollers_changed(keel_overlay_scrollers());
            }];
        keel_scrollers_changed(keel_overlay_scrollers());
    });
}
