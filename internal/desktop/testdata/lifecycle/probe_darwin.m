//go:build darwin && cgo

#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>
#include <unistd.h>

static int passed;

static void sendApplicationEvent(AEEventID eventID) {
	pid_t pid = getpid();
	NSAppleEventDescriptor *target = [NSAppleEventDescriptor descriptorWithDescriptorType:typeKernelProcessID bytes:&pid length:sizeof(pid)];
    NSAppleEventDescriptor *event = [NSAppleEventDescriptor appleEventWithEventClass:kCoreEventClass
        eventID:eventID targetDescriptor:target returnID:kAutoGenerateReturnID transactionID:kAnyTransactionID];
    OSStatus status = AESendMessage(event.aeDesc, NULL, kAENoReply, kAEDefaultTimeout);
    if (status != noErr) NSLog(@"FAIL: queued Apple Event status=%d", (int)status);
}

static NSUInteger visibleWindows(void) {
    NSUInteger count = 0;
    for (NSWindow *window in NSApp.windows) {
        if ([window isKindOfClass:NSClassFromString(@"GUIWindow")] && window.isVisible) count++;
    }
    return count;
}

void startLifecycleProbe(void) {
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, 3000 * NSEC_PER_MSEC), dispatch_get_main_queue(), ^{
        NSUInteger before = visibleWindows();
        if (before != 2) {NSLog(@"FAIL: expected two visible windows"); return;}
        [NSApp hide:nil];
        dispatch_after(dispatch_time(DISPATCH_TIME_NOW, 200 * NSEC_PER_MSEC), dispatch_get_main_queue(), ^{
            if (![NSApp isHidden] && visibleWindows() != 0) { NSLog(@"FAIL: application windows did not hide"); return; }
            sendApplicationEvent(kAEReopenApplication);
            dispatch_after(dispatch_time(DISPATCH_TIME_NOW, 200 * NSEC_PER_MSEC), dispatch_get_main_queue(), ^{
                if ([NSApp isHidden] || visibleWindows() != before) { NSLog(@"FAIL: application windows did not restore"); return; }
                passed = 1;
                // Use Dock's actual Apple Event protocol. The probe does not
                // post a wake event or synthesize keyboard/mouse input.
                sendApplicationEvent(kAEQuitApplication);
            });
        });
    });
}

int lifecycleProbePassed(void) { return passed; }
