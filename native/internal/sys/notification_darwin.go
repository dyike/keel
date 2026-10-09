//go:build darwin && !ios

package sys

import (
	"fmt"
	"sync"

	"github.com/dyike/keel/native"
	"github.com/ebitengine/purego/objc"
)

// Notifications use UNUserNotificationCenter, which only works for apps
// running from a bundle with an identifier.

const (
	notificationUserInfoKey = "keel.notification"
	unAuthorizationAlert    = 1 << 2
	unPresentList           = 1 << 3
	unPresentBanner         = 1 << 4
	unAuthorized            = 2
	unProvisional           = 3
)

func NotificationAvailable() bool {
	load()
	var ok bool
	withPool(func() {
		bundle := send(class("NSBundle"), "mainBundle")
		ext := goString(send(send(send(bundle, "bundleURL"), "pathExtension"), "lowercaseString"))
		ok = send(send(bundle, "bundleIdentifier"), "length") > 0 && ext == "app"
	})
	return ok
}

func ours(notification id) bool {
	info := send(send(send(notification, "request"), "content"), "userInfo")
	return sendBool(send(info, "objectForKey:", uintptr(nsString(notificationUserInfoKey))), "boolValue")
}

var notificationDelegate struct {
	once sync.Once
	obj  id // main thread only
}

// installNotificationDelegate runs on the main queue. It does not replace
// another integration's delegate.
func installNotificationDelegate(center id) bool {
	notificationDelegate.once.Do(func() {
		var protocols []*objc.Protocol
		if p := objc.GetProtocol("UNUserNotificationCenterDelegate"); p != nil {
			protocols = append(protocols, p)
		}
		c, err := objc.RegisterClass("KeelNotificationDelegate", objc.GetClass("NSObject"), protocols, nil, []objc.MethodDef{
			{Cmd: sel("userNotificationCenter:willPresentNotification:withCompletionHandler:"), Fn: func(_ id, _ objc.SEL, _, n id, handler uintptr) {
				var options uintptr
				if ours(n) {
					options = unPresentBanner | unPresentList
				}
				callBlock(handler, options)
			}},
			{Cmd: sel("userNotificationCenter:didReceiveNotificationResponse:withCompletionHandler:"), Fn: func(_ id, _ objc.SEL, _, response id, handler uintptr) {
				n := send(response, "notification")
				if ours(n) && sendBool(send(response, "actionIdentifier"), "isEqualToString:", uintptr(constant("UNNotificationDefaultActionIdentifier"))) {
					if fn := systemNotificationClicks.take(goString(send(send(n, "request"), "identifier"))); fn != nil {
						go fn()
					}
				}
				callBlock(handler)
			}},
		})
		if err == nil {
			notificationDelegate.obj = send(send(id(c), "alloc"), "init")
		}
	})
	d := notificationDelegate.obj
	if d == 0 {
		return false
	}
	if current := send(center, "delegate"); current != 0 && current != d {
		return false
	}
	send(center, "setDelegate:", uintptr(d))
	return true
}

func notificationCenter() id {
	return send(class("UNUserNotificationCenter"), "currentNotificationCenter")
}

func NotificationPermission(done func(error)) {
	if !NotificationAvailable() {
		done(status(2))
		return
	}
	mainAsync(func() {
		blk := objc.NewBlock(func(_ objc.Block, granted bool, err id) {
			if err != 0 {
				done(nsNotificationError(err))
			} else if granted {
				done(nil)
			} else {
				done(status(1))
			}
		})
		send(notificationCenter(), "requestAuthorizationWithOptions:completionHandler:", unAuthorizationAlert, uintptr(blk))
		blk.Release()
	})
}

func NotificationPost(id, title, body string, done func(error)) {
	NotificationPostInteractive(id, title, body, nil, done)
}

func NotificationPostInteractive(ident, title, body string, onClick func(), done func(error)) {
	finish := systemNotificationClicks.register(ident, onClick)
	completed := func(err error) { finish(err); done(err) }
	if !NotificationAvailable() {
		completed(status(2))
		return
	}
	mainAsync(func() {
		center := notificationCenter()
		if !installNotificationDelegate(center) {
			completed(status(6))
			return
		}
		content := send(send(class("UNMutableNotificationContent"), "alloc"), "init")
		send(content, "setTitle:", uintptr(nsString(title)))
		send(content, "setBody:", uintptr(nsString(body)))
		yes := send(class("NSNumber"), "numberWithBool:", 1)
		send(content, "setUserInfo:", uintptr(send(class("NSDictionary"), "dictionaryWithObject:forKey:", uintptr(yes), uintptr(nsString(notificationUserInfoKey)))))
		request := send(send(class("UNNotificationRequest"), "requestWithIdentifier:content:trigger:", uintptr(nsString(ident)), uintptr(content), 0), "retain")
		release(content)
		added := objc.NewBlock(func(_ objc.Block, err id) {
			if err != 0 {
				completed(nsNotificationError(err))
			} else {
				completed(nil)
			}
		})
		settings := objc.NewBlock(func(_ objc.Block, settings id) {
			defer release(request)
			if s := send(settings, "authorizationStatus"); s != unAuthorized && s != unProvisional {
				completed(status(1))
				return
			}
			withPool(func() {
				send(center, "addNotificationRequest:withCompletionHandler:", uintptr(request), uintptr(added))
			})
			added.Release()
		})
		send(center, "getNotificationSettingsWithCompletionHandler:", uintptr(settings))
		settings.Release()
	})
}

func NotificationRemove(ident string, done func(error)) {
	if !NotificationAvailable() {
		done(status(2))
		return
	}
	mainAsync(func() {
		center := notificationCenter()
		ids := send(class("NSArray"), "arrayWithObject:", uintptr(nsString(ident)))
		send(center, "removePendingNotificationRequestsWithIdentifiers:", uintptr(ids))
		send(center, "removeDeliveredNotificationsWithIdentifiers:", uintptr(ids))
		systemNotificationClicks.take(ident)
		done(nil)
	})
}

// nsNotificationError copies an NSError while the system still owns it.
func nsNotificationError(err id) error {
	var e error
	withPool(func() {
		e = notificationFailure(goString(send(err, "domain")), int64(send(err, "code")), goString(send(err, "localizedDescription")))
	})
	return e
}

func notificationFailure(domain string, code int64, description string) error {
	cause := native.ErrFailed
	// UNErrorCodeNotificationsNotAllowed is 1. Check the domain as codes
	// from unrelated NSError domains can have the same numeric value.
	if domain == "UNErrorDomain" && code == 1 {
		cause = native.ErrPermissionDenied
	}
	return fmt.Errorf("%w: %s (%d): %s", cause, domain, code, description)
}
