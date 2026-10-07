//go:build darwin && !ios && cgo
#import <AppKit/AppKit.h>

int keel_system_dark(void) {
    return [[[NSUserDefaults standardUserDefaults] stringForKey:@"AppleInterfaceStyle"] isEqualToString:@"Dark"];
}
void keel_native_appearance(int mode) {
    dispatch_async(dispatch_get_main_queue(),^{
        if(!NSApp) return;
        NSApp.appearance=mode==0?nil:[NSAppearance appearanceNamed:mode==2?NSAppearanceNameDarkAqua:NSAppearanceNameAqua];
    });
}
