//go:build ios && cgo

#import <UIKit/UIKit.h>
#import <objc/runtime.h>

// Gio v0.10's app delegate creates a window before a scene exists. When
// explicitly selected in Info.plist, move that work to the scene delegate.
// This compatibility adapter depends on Gio v0.10's Objective-C class names.
static BOOL keel_didFinishLaunching(id self, SEL cmd, UIApplication *application,
                                  NSDictionary *options) {
    return YES;
}

int keel_prepareIOSScene(void) {
    NSDictionary *manifest = NSBundle.mainBundle.infoDictionary[@"UIApplicationSceneManifest"];
    NSArray *configurations = manifest[@"UISceneConfigurations"][@"UIWindowSceneSessionRoleApplication"];
    BOOL selected = NO;
    for (NSDictionary *configuration in configurations) {
        if ([configuration[@"UISceneDelegateClassName"] isEqualToString:@"KeelSceneDelegate"]) selected = YES;
    }
    if (!selected) return 0;
    Class delegate = NSClassFromString(@"_gioAppDelegate");
    SEL selector = @selector(application:didFinishLaunchingWithOptions:);
    Method method = class_getInstanceMethod(delegate, selector);
    if (!method || !NSClassFromString(@"GioViewController")) return 1;
    class_replaceMethod(delegate, selector, (IMP)keel_didFinishLaunching, method_getTypeEncoding(method));
    return 0;
}

@interface KeelSceneDelegate : UIResponder <UIWindowSceneDelegate>
@property (strong, nonatomic) UIWindow *window;
@end

@implementation KeelSceneDelegate
- (void)scene:(UIScene *)scene
    willConnectToSession:(UISceneSession *)session
    options:(UISceneConnectionOptions *)connectionOptions {
    if (![scene isKindOfClass:[UIWindowScene class]]) return;
    self.window = [[UIWindow alloc] initWithWindowScene:(UIWindowScene *)scene];
    // Gio registers a display link in NSRunLoop.currentMode while loading
    // its view. Run that work after UIKit starts the main run loop.
    [self performSelector:@selector(showWindow) withObject:nil afterDelay:0];
}
- (void)showWindow {
    self.window.rootViewController = [[NSClassFromString(@"GioViewController") alloc] init];
    [self.window makeKeyAndVisible];
}
@end
