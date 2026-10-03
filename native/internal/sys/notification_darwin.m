//go:build darwin && !ios && cgo

#import <Foundation/Foundation.h>
#import <UserNotifications/UserNotifications.h>
#include <stdint.h>

extern void keel_notification_done(uintptr_t token, int code);

int keel_notification_available(void) {
    @autoreleasepool {
        NSBundle *bundle = NSBundle.mainBundle;
        return bundle.bundleIdentifier.length > 0 &&
            [bundle.bundleURL.pathExtension.lowercaseString isEqualToString:@"app"];
    }
}

void keel_notification_permission(uintptr_t token) {
    if (!keel_notification_available()) { keel_notification_done(token, 2); return; }
    dispatch_async(dispatch_get_main_queue(), ^{
        [UNUserNotificationCenter.currentNotificationCenter
            requestAuthorizationWithOptions:UNAuthorizationOptionAlert
            completionHandler:^(BOOL granted, NSError *error) {
                keel_notification_done(token, error ? 7 : (granted ? 0 : 1));
            }];
    });
}

void keel_notification_post(const char *id, const char *title, const char *body, uintptr_t token) {
    if (!keel_notification_available()) { keel_notification_done(token, 2); return; }
    @autoreleasepool {
        // Copy all C input before returning to Go; the queued block owns it.
        NSString *identifier = [NSString stringWithUTF8String:id];
        UNMutableNotificationContent *content = [UNMutableNotificationContent new];
        content.title = [NSString stringWithUTF8String:title];
        content.body = [NSString stringWithUTF8String:body];
        dispatch_async(dispatch_get_main_queue(), ^{
            UNUserNotificationCenter *center = UNUserNotificationCenter.currentNotificationCenter;
            [center getNotificationSettingsWithCompletionHandler:^(UNNotificationSettings *settings) {
                if (settings.authorizationStatus != UNAuthorizationStatusAuthorized &&
                    settings.authorizationStatus != UNAuthorizationStatusProvisional) {
                    keel_notification_done(token, 1);
                    return;
                }
                UNNotificationRequest *request = [UNNotificationRequest requestWithIdentifier:identifier content:content trigger:nil];
                [center addNotificationRequest:request withCompletionHandler:^(NSError *error) {
                    keel_notification_done(token, error ? 7 : 0);
                }];
            }];
        });
    }
}

void keel_notification_remove(const char *id, uintptr_t token) {
    if (!keel_notification_available()) { keel_notification_done(token, 2); return; }
    @autoreleasepool {
        NSString *identifier = [NSString stringWithUTF8String:id];
        dispatch_async(dispatch_get_main_queue(), ^{
            UNUserNotificationCenter *center = UNUserNotificationCenter.currentNotificationCenter;
            [center removePendingNotificationRequestsWithIdentifiers:@[identifier]];
            [center removeDeliveredNotificationsWithIdentifiers:@[identifier]];
            keel_notification_done(token, 0);
        });
    }
}
