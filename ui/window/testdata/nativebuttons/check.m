//go:build darwin && !ios && cgo
#import <AppKit/AppKit.h>
#include <stdio.h>
#include <math.h>

int check_native_buttons(int action, int report, double height, double left, double offsetY, double spacing) {
 __block int ok = 0;
 dispatch_sync(dispatch_get_main_queue(), ^{
  for (NSWindow *window in NSApp.windows) {
   if (![window.title isEqualToString:@"Native buttons acceptance"]) continue;
   if (action == 4) { ok = window.miniaturized; break; }
   if (action == 5) { [window deminiaturize:nil]; ok = 1; break; }
   if (action == 6) { ok = !window.miniaturized; break; }
   ok = window.titlebarAppearsTransparent && window.titleVisibility == NSWindowTitleHidden && (window.styleMask & NSWindowStyleMaskFullSizeContentView);
   for (NSUInteger kind = NSWindowCloseButton; kind <= NSWindowZoomButton; kind++) {
    NSButton *button = [window standardWindowButton:(NSWindowButton)kind];
    if (!button || button.hiddenOrHasHiddenAncestor || !button.enabled || !button.action || button.window != window) ok = 0;
    NSView *content = window.contentView;
    NSPoint center = [content convertPoint:NSMakePoint(NSMidX(button.bounds), NSMidY(button.bounds)) fromView:button];
    CGFloat top = content.isFlipped ? center.y : NSHeight(content.bounds) - center.y;
    CGFloat expectedX = left + NSWidth(button.frame)/2 + kind*spacing;
    if (fabs(center.x-expectedX) > .5 || fabs(top-height/2-offsetY) > .5) {
     ok = 0;
     if (report) fprintf(stderr, "placement %lu: actual %.1f,%.1f expected %.1f,%.1f\n", (unsigned long)kind, center.x,top,expectedX,height/2+offsetY);
    }
    NSView *frameView = content.superview;
    NSPoint hitPoint = [frameView convertPoint:center fromView:content];
    NSView *hit = [frameView hitTest:hitPoint];
    if (hit != button && ![hit isDescendantOf:button]) {
     ok = 0;
     if (report) fprintf(stderr, "button %lu is not hittable: %s\n", (unsigned long)kind, NSStringFromClass(hit.class).UTF8String);
    }
    if (report) fprintf(stderr, "button %lu: %s / %s, hidden=%d, enabled=%d, action=%s\n", (unsigned long)kind, NSStringFromClass(button.class).UTF8String, NSStringFromClass(button.cell.class).UTF8String, button.hiddenOrHasHiddenAncestor, button.enabled, NSStringFromSelector(button.action).UTF8String);
   }
   NSView *content = window.contentView;
   NSPoint tabPoint = NSMakePoint(230,height/2);
   if (!content.isFlipped) tabPoint.y = NSHeight(content.bounds)-tabPoint.y;
   NSView *tabHit = [content.superview hitTest:[content.superview convertPoint:tabPoint fromView:content]];
   if (tabHit != content && ![tabHit isDescendantOf:content]) {
    ok = 0;
    if (report) fprintf(stderr, "titlebar intercepts content tabs: %s\n", NSStringFromClass(tabHit.class).UTF8String);
   }
   if (action == 1) [window setContentSize:NSMakeSize(720,360)];
   if (action == 2) {
    [[window standardWindowButton:NSWindowMiniaturizeButton] performClick:nil];
   }
   if (action == 3) [[window standardWindowButton:NSWindowCloseButton] performClick:nil];
   break;
  }
 });
 return ok;
}
