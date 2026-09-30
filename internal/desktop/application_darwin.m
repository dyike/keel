//go:build darwin && cgo

#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

static NSArray<NSWindow *> *hiddenWindows;
static __weak NSWindow *hiddenKeyWindow;
static id launchObserver;
static BOOL appKitLaunched;

static void traceApplication(const char *action) {
    if (!getenv("KEEL_APP_TRACE")) return;
    fprintf(stderr, "[keel-app pid=%d] %s running=%d active=%d hidden=%d launched=%d\n",
        getpid(), action, NSApp.isRunning, NSApp.isActive, NSApp.isHidden,
        NSRunningApplication.currentApplication.finishedLaunching);
    fflush(stderr);
}

static void wakeApplication(void) {
    [NSApp postEvent:[NSEvent otherEventWithType:NSEventTypeApplicationDefined
        location:NSZeroPoint modifierFlags:0 timestamp:0 windowNumber:0
        context:nil subtype:0 data1:0 data2:0] atStart:YES];
}

@interface KeelApplication : NSApplication
@end

@implementation KeelApplication
- (void)terminate:(id)sender {
    traceApplication("terminate:");
    // Metal's delegate cancels AppKit termination and sets a quit flag so Go
    // can clean up. A Dock request arrives during nextEvent's blocking wait;
    // wake that wait so the backend can consume the flag without another click.
    [super terminate:sender];
    wakeApplication();
}
- (void)hide:(id)sender {
    traceApplication("hide:");
    [super hide:sender];
    if ([super isHidden]) return;
    // A CLI launch can be refused by Launch Services' application-hide API.
    // Hide the visible Metal windows explicitly, retaining the list so reopen
    // restores them without exposing intentionally hidden utility windows.
    NSMutableArray<NSWindow *> *visible = hiddenWindows ? [hiddenWindows mutableCopy] : [NSMutableArray array];
    if (self.keyWindow) hiddenKeyWindow = self.keyWindow;
    Class metalWindow = NSClassFromString(@"GUIWindow");
    for (NSWindow *window in self.windows) {
        if ([window isKindOfClass:metalWindow] && window.isVisible) {
            window.releasedWhenClosed = NO;
            if (![visible containsObject:window]) [visible addObject:window];
            [window orderOut:nil];
        }
    }
    hiddenWindows = visible;
    wakeApplication();
}
- (void)unhide:(id)sender {
    traceApplication("unhide:");
    [super unhide:sender];
    NSArray<NSWindow *> *windows = hiddenWindows;
    hiddenWindows = nil;
    for (NSWindow *window in windows) {
        // Metal clears the native delegate when destroying a window.
        if (window.delegate) [window orderFront:nil];
    }
    if (hiddenKeyWindow.delegate) [hiddenKeyWindow makeKeyAndOrderFront:nil];
    hiddenKeyWindow = nil;
    wakeApplication();
}
- (void)keelQuitEvent:(NSAppleEventDescriptor *)event reply:(NSAppleEventDescriptor *)reply {
    traceApplication("Apple Event quit");
    [self terminate:nil];
}
- (void)keelReopenEvent:(NSAppleEventDescriptor *)event reply:(NSAppleEventDescriptor *)reply {
    traceApplication("Apple Event reopen");
    [self unhide:nil];
}
@end

int keelApplicationPrepare(void) {
    [KeelApplication sharedApplication];
    if (![NSApp isKindOfClass:[KeelApplication class]]) return 0;
    launchObserver = [[NSNotificationCenter defaultCenter] addObserverForName:NSApplicationDidFinishLaunchingNotification object:NSApp queue:nil usingBlock:^(NSNotification *notification) {
        appKitLaunched = YES;
    }];
    return 1;
}

void keelApplicationFinishLaunch(void) {
    traceApplication("finish launch: before");
    // CLI processes are already "launched" to Launch Services, so Metal skips
    // its AppKit bootstrap. Register AppKit's system event handlers explicitly.
    if (!appKitLaunched && [[NSRunningApplication currentApplication] isFinishedLaunching]) {
        [NSApp finishLaunching];
    }
    NSAppleEventManager *events = [NSAppleEventManager sharedAppleEventManager];
    [events setEventHandler:NSApp andSelector:@selector(keelQuitEvent:reply:)
             forEventClass:kCoreEventClass andEventID:kAEQuitApplication];
    [events setEventHandler:NSApp andSelector:@selector(keelReopenEvent:reply:)
             forEventClass:kCoreEventClass andEventID:kAEReopenApplication];
    // Also expose the conventional application-menu Hide / Cmd+H entry.
    NSMenu *menu = [NSApp.mainMenu itemAtIndex:0].submenu;
    NSMenuItem *hide = [[NSMenuItem alloc] initWithTitle:@"Hide" action:@selector(hide:) keyEquivalent:@"h"];
    hide.target = NSApp;
    [menu insertItem:hide atIndex:MAX(0, menu.numberOfItems - 1)];
    traceApplication("finish launch: ready");
}

void keelApplicationCleanup(void) {
    traceApplication("cleanup");
    hiddenWindows = nil;
    hiddenKeyWindow = nil;
    if (launchObserver) [[NSNotificationCenter defaultCenter] removeObserver:launchObserver];
    launchObserver = nil;
    NSAppleEventManager *events = [NSAppleEventManager sharedAppleEventManager];
    [events removeEventHandlerForEventClass:kCoreEventClass andEventID:kAEQuitApplication];
    [events removeEventHandlerForEventClass:kCoreEventClass andEventID:kAEReopenApplication];
}
