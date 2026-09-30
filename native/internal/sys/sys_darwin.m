//go:build darwin && cgo

#import <AppKit/AppKit.h>
#import <ApplicationServices/ApplicationServices.h>
#import <Carbon/Carbon.h>
#import <ScreenCaptureKit/ScreenCaptureKit.h>
#include "sys_darwin.h"
extern void keelHotkeyFired(uint32_t id);

int keel_permission(int kind,int request) {
 @autoreleasepool {
 switch(kind) {
 case 0: if(request) return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)@{(__bridge NSString*)kAXTrustedCheckOptionPrompt:@YES}); return AXIsProcessTrusted();
 case 1: return request ? CGRequestScreenCaptureAccess() : CGPreflightScreenCaptureAccess();
 case 2: return request ? CGRequestListenEventAccess() : CGPreflightListenEventAccess();
 default:return -3;
 }
 }
}
int keel_displays(keel_display **out,uint32_t *count) {
 uint32_t n=0;CGError err=CGGetActiveDisplayList(0,NULL,&n);if(err!=kCGErrorSuccess)return 100+(int)err;
 if(!n){*out=NULL;*count=0;return 0;}
 CGDirectDisplayID *ids=calloc(n,sizeof(*ids));if(!ids)return 100;
 err=CGGetActiveDisplayList(n,ids,&n);if(err!=kCGErrorSuccess){free(ids);return 100+(int)err;}
 keel_display *list=calloc(n,sizeof(*list));if(!list){free(ids);return 100;}
 for(uint32_t i=0;i<n;i++){CGRect b=CGDisplayBounds(ids[i]);list[i]=(keel_display){ids[i],b.origin.x,b.origin.y,b.size.width,b.size.height,CGDisplayPixelsWide(ids[i]),CGDisplayPixelsHigh(ids[i]),CGDisplayIsMain(ids[i])};}
 free(ids);*out=list;*count=n;return 0;
}
int keel_capture(uint32_t id,void **data,size_t *length) {
 @autoreleasepool {
 if([NSThread isMainThread])return 4;
 if(!CGPreflightScreenCaptureAccess())return 1;
 if(!CGDisplayIsActive(id))return 3;
 dispatch_semaphore_t done=dispatch_semaphore_create(0);
 __block NSData *png=nil;__block int status=100;
 [SCShareableContent getShareableContentExcludingDesktopWindows:NO onScreenWindowsOnly:YES completionHandler:^(SCShareableContent *content,NSError *error){
  @autoreleasepool {
  if(error||!content){dispatch_semaphore_signal(done);return;}
  SCDisplay *display=nil;for(SCDisplay *d in content.displays){if(d.displayID==id){display=d;break;}}
  if(!display){status=3;dispatch_semaphore_signal(done);return;}
  SCContentFilter *filter=[[SCContentFilter alloc] initWithDisplay:display excludingWindows:@[]];
  SCStreamConfiguration *config=[SCStreamConfiguration new];
  config.width=CGDisplayPixelsWide(id);config.height=CGDisplayPixelsHigh(id);config.showsCursor=NO;
  [SCScreenshotManager captureImageWithFilter:filter configuration:config completionHandler:^(CGImageRef image,NSError *captureError){
   @autoreleasepool {
   if(image&&!captureError){NSBitmapImageRep *rep=[[NSBitmapImageRep alloc] initWithCGImage:image];png=[rep representationUsingType:NSBitmapImageFileTypePNG properties:@{}];if(png)status=0;}
   dispatch_semaphore_signal(done);
   }
  }];
  }
 }];
 // Blocks retain their state after a timeout; no stack or Go pointers escape.
 if(dispatch_semaphore_wait(done,dispatch_time(DISPATCH_TIME_NOW,10*NSEC_PER_SEC)))return 5;
 if(status)return status;
 *length=png.length;*data=malloc(*length);if(!*data)return 100;memcpy(*data,png.bytes,*length);return 0;
 }
}
int keel_position(double *x,double *y){CGEventRef e=CGEventCreate(NULL);if(!e)return 100;CGPoint p=CGEventGetLocation(e);CFRelease(e);*x=p.x;*y=p.y;return 0;}
int keel_move(double x,double y){if(!AXIsProcessTrusted())return 1;CGEventRef e=CGEventCreateMouseEvent(NULL,kCGEventMouseMoved,CGPointMake(x,y),kCGMouseButtonLeft);if(!e)return 100;CGEventPost(kCGHIDEventTap,e);CFRelease(e);return 0;}
int keel_click(int button){
 if(!AXIsProcessTrusted())return 1;if(button<0||button>2)return 3;
 double x,y;if(keel_position(&x,&y))return 100;
 CGEventType down[]={kCGEventLeftMouseDown,kCGEventRightMouseDown,kCGEventOtherMouseDown};
 CGEventType up[]={kCGEventLeftMouseUp,kCGEventRightMouseUp,kCGEventOtherMouseUp};
 CGEventRef d=CGEventCreateMouseEvent(NULL,down[button],CGPointMake(x,y),(CGMouseButton)button);
 CGEventRef u=CGEventCreateMouseEvent(NULL,up[button],CGPointMake(x,y),(CGMouseButton)button);
 if(!d||!u){if(d)CFRelease(d);if(u)CFRelease(u);return 100;}
 CGEventSetIntegerValueField(d,kCGMouseEventClickState,1);CGEventSetIntegerValueField(u,kCGMouseEventClickState,1);
 CGEventPost(kCGHIDEventTap,d);CGEventPost(kCGHIDEventTap,u);CFRelease(d);CFRelease(u);return 0;
}
int keel_key(unsigned short key,int down){if(!AXIsProcessTrusted())return 1;CGEventRef e=CGEventCreateKeyboardEvent(NULL,key,down);if(!e)return 100;CGEventPost(kCGHIDEventTap,e);CFRelease(e);return 0;}
static void onMain(dispatch_block_t block){if([NSThread isMainThread])block();else dispatch_sync(dispatch_get_main_queue(),block);}
static EventHandlerRef hotkeyHandler=NULL;
static unsigned int hotkeyCount=0;
static OSStatus hotkeyEvent(EventHandlerCallRef next,EventRef event,void *context){
 EventHotKeyID id;OSStatus s=GetEventParameter(event,kEventParamDirectObject,typeEventHotKeyID,NULL,sizeof(id),NULL,&id);
 if(s!=noErr||id.signature!='KEEL')return eventNotHandledErr;keelHotkeyFired(id.id);return noErr;
}
int keel_hotkey_register(uint32_t id,unsigned short key,unsigned int modifiers,void **ref){
 __block OSStatus result=noErr;__block EventHotKeyRef hotkey=NULL;
 onMain(^{
  if(!hotkeyHandler){EventTypeSpec spec={kEventClassKeyboard,kEventHotKeyPressed};result=InstallEventHandler(GetApplicationEventTarget(),hotkeyEvent,1,&spec,NULL,&hotkeyHandler);if(result!=noErr)return;}
  UInt32 flags=0;if(modifiers&1)flags|=controlKey;if(modifiers&2)flags|=optionKey;if(modifiers&4)flags|=shiftKey;if(modifiers&8)flags|=cmdKey;
  EventHotKeyID ident={'KEEL',id};result=RegisterEventHotKey(key,flags,ident,GetApplicationEventTarget(),0,&hotkey);
  if(result==noErr)hotkeyCount++;else if(!hotkeyCount){RemoveEventHandler(hotkeyHandler);hotkeyHandler=NULL;}
 });
 *ref=hotkey;if(result==eventHotKeyExistsErr)return 6;return result==noErr?0:(int)result;
}
int keel_hotkey_unregister(void *ref){
 __block OSStatus result=noErr;onMain(^{result=UnregisterEventHotKey((EventHotKeyRef)ref);if(result==noErr&&hotkeyCount){hotkeyCount--;if(!hotkeyCount){RemoveEventHandler(hotkeyHandler);hotkeyHandler=NULL;}}});return (int)result;
}
