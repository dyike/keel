//go:build darwin && !ios && cgo

#import <AppKit/AppKit.h>
extern void keel_scroll_event(int precise, int active, int momentum, int ended);

// keel_watch_scroll observes scroll events before Gio's view receives them,
// reporting the device and gesture phase that Gio drops.
void keel_watch_scroll(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        static id monitor;
        monitor = [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskScrollWheel handler:^NSEvent *(NSEvent *event) {
            NSEventPhase phase = event.phase, momentum = event.momentumPhase;
            int active = (phase & (NSEventPhaseBegan | NSEventPhaseStationary | NSEventPhaseChanged | NSEventPhaseMayBegin)) != 0;
            int ended = (phase & (NSEventPhaseEnded | NSEventPhaseCancelled)) != 0;
            int coasting = (momentum & (NSEventPhaseBegan | NSEventPhaseChanged)) != 0;
            keel_scroll_event(event.hasPreciseScrollingDeltas, active, coasting, ended);
            return event;
        }];
    });
}
