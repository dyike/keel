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
