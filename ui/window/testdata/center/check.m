//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>
#include <math.h>
#include <stdio.h>

int check_positions(int move, int report) {
    __block int ok = 1;
    dispatch_sync(dispatch_get_main_queue(), ^{
        int count = 0;
        for (NSWindow *window in NSApp.windows) {
            if (![window.title hasPrefix:@"Center "]) continue;
            count++;
            NSRect r = window.screen.visibleFrame;
            NSRect frame = window.frame;
            NSPoint expected = NSMakePoint(NSMidX(r)-NSWidth(frame)/2, NSMidY(r)-NSHeight(frame)/2);
            if (move == 1) {
                [window setFrameOrigin:NSMakePoint(expected.x+35, expected.y-25)];
                continue;
            }
            if (move == 2) { expected.x += 35; expected.y -= 25; }
            if (fabs(frame.origin.x-expected.x) > 1 || fabs(frame.origin.y-expected.y) > 1) {
                ok = 0;
                if (report) fprintf(stderr, "%s: actual %.1f,%.1f expected %.1f,%.1f\n",
                    window.title.UTF8String, frame.origin.x, frame.origin.y, expected.x, expected.y);
            }
        }
        if (count != 2) {
            ok = 0;
            if (report) fprintf(stderr, "expected two windows, got %d\n", count);
        }
    });
    return ok;
}
