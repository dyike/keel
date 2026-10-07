//go:build darwin && !ios && cgo && keelnativeqa
#import <AppKit/AppKit.h>
static NSMenuItem *keel_qa_find(NSMenu *menu,NSString *name) {
 for(NSMenuItem *item in menu.itemArray) {
  if([item.representedObject isEqualToString:name]) return item;
  NSMenuItem *found=keel_qa_find(item.submenu,name);if(found)return found;
 }
 return nil;
}
void keel_qa_menu(const char *command) {
 NSString *name=[[NSString alloc]initWithUTF8String:command];
 dispatch_async(dispatch_get_main_queue(),^{
  NSMenuItem *item=keel_qa_find(NSApp.mainMenu,name);
  if(item && item.enabled) [NSApp sendAction:item.action to:item.target from:item];
 });
 [name release];
}

extern void keel_native_qa_wake(void);
void keel_qa_ime(int commit) {
 dispatch_async(dispatch_get_main_queue(),^{
  // The fixture may run while another app owns keyboard focus.
  // Deliver input to the application's own window rather than messaging a nil responder.
  NSWindow *window=NSApp.keyWindow;
  if(!window) {
   for(NSWindow *candidate in NSApp.windows) {
    if(candidate.isVisible && candidate.canBecomeKeyWindow) {
     [candidate makeKeyAndOrderFront:nil];window=candidate;break;
    }
   }
  }
  id view=window.firstResponder;
  if(commit) {
   [view insertText:@"你好中文" replacementRange:NSMakeRange(NSNotFound,0)];
  } else {
   [view setMarkedText:@"imepreedit" selectedRange:NSMakeRange(10,0) replacementRange:NSMakeRange(NSNotFound,0)];
  }
  keel_native_qa_wake();
 });
}

// Test fixture: read AppKit's actual Dock icon on its owning thread.
int keel_qa_icon(const char *path, const char *source, const char *reference) {
 NSString *output=[[NSString alloc]initWithUTF8String:path];
 NSString *input=[[NSString alloc]initWithUTF8String:source];
 NSString *expected=[[NSString alloc]initWithUTF8String:reference];
 __block int ok=0;
 dispatch_sync(dispatch_get_main_queue(),^{
  NSBitmapImageRep *rep=[NSBitmapImageRep imageRepWithData:NSApp.applicationIconImage.TIFFRepresentation];
  rep=[rep bitmapImageRepByConvertingToColorSpace:NSColorSpace.sRGBColorSpace renderingIntent:NSColorRenderingIntentDefault];
  NSData *png=[rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
  ok=[png writeToFile:output atomically:YES];
  NSImage *sourceImage=[[NSImage alloc]initWithContentsOfFile:input];
  NSBitmapImageRep *sourceRep=[NSBitmapImageRep imageRepWithData:sourceImage.TIFFRepresentation];
  sourceRep=[sourceRep bitmapImageRepByConvertingToColorSpace:NSColorSpace.sRGBColorSpace renderingIntent:NSColorRenderingIntentDefault];
  NSData *sourcePNG=[sourceRep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];
  ok=ok && [sourcePNG writeToFile:expected atomically:YES];
  [sourceImage release];
 });
 [output release];
 [input release];
 [expected release];
 return ok;
}
