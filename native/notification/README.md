# native/notification

English | [简体中文](README.zh-CN.md)

Independent system notification interface, currently implements macOS 14+, cgo, `.app` with bundle identifier, Linux desktop D-Bus notification service, and Windows tray bubble notification. WebAssembly, iOS/Android, and macOS cgo-less builds return `native.ErrUnsupported`; macOS plain `go run` is also not supported. Linux does not require X11, and Wayland also communicates through the session bus.

```go
notification.RequestPermission(func(err error) {
    if err != nil { return }
    notification.Post(notification.Message{
        ID: "download/report", Title: "下载完成", Body: "report.pdf 已保存。",
    }, func(err error) {
        // err == nil means that the system accepted the request, but does not mean that the banner has been displayed.
    })
})
// Recall by the same ID after Post completes:
notification.Remove("download/report", func(err error) {})
```

- `Available()` only checks platform implementation and application package identity, not authorization, signature trust, or system presentation policy.
- `RequestPermission` applies for alert permission; `Post` does not automatically pop up the permission prompt, and returns `ErrPermissionDenied` without authorization. NSError of macOS retains domain, code and localization description; NotificationsNotAllowed of `UNErrorDomain` is mapped to `ErrPermissionDenied`, other system errors wrap `ErrFailed`, use `errors.Is` to determine the classification, do not compare the complete error string.
- IDs are shared across the entire application, are non-empty and cannot contain NUL; at least one of the body or title is non-empty, and all strings must be legal UTF-8. Re-delivery with the same ID replaces existing requests by the system.
- The completion callback is executed in a separate goroutine and nil can be passed. UI modifications should be put into `core.Update`; do not block the main thread while waiting for asynchronous completion.
- Operations with the same ID should wait for the previous one to complete before executing to avoid asynchronous delivery and recall interleaving. Remove simultaneously withdraws pending delivery and delivered items; the system does not confirm the withdrawal completion, and a successful callback only means that a withdrawal call has been issued.
- macOS must run the app event loop; permissions dialogs, focus mode, system settings, and signature trusts affect actual display. macOS has registered the notification center delegate to request the front-end Banner/List for notifications of this module; system settings can still prohibit display. Use `window.Window.Activate(token)` to bring the window to the front (the activation token brought by the Linux system click can be handed to it directly, see kit/notifier.md); cross-boot callback recovery is not implemented.
- Windows uses the tray bubble (`Shell_NotifyIcon`), and Windows 10/11 displays it as a system notification banner. It does not require AppUserModelID, packaging or start menu shortcuts, nor cgo; `RequestPermission` does not pop up authorization and succeeds directly. Only one message is displayed at the same time: delivering a new ID will replace the current notification, and the click callback of the old notification will become invalid; withdrawing the replaced ID will not do anything. When the notification is displayed, there is an application icon in the tray (take the first icon resource of the executable file, if not, use the system default icon), the icon will be removed after the notification times out, is closed, clicked or withdrawn, and this item in the action center will also disappear. Clicking the notification or tray icon runs `OnClick`; there is no activation token. Truncate and add ellipsis when the title exceeds 63 UTF-16 units and the body exceeds 255. The system may not display the focus assistant/do not disturb mode.
- kit.Notifier is accessed through NoticeSystemBackend; `examples/notification` provides an adapter to handle system/application-only system addition, same-ID update and withdrawal. The macOS example accesses kit click, close, and Window.Raise through NoticeSystemInteractiveBackend; the Linux example also uses the interactive backend, requiring the notification service to declare actions.

`examples/notification` can be run manually. Build the application package first, then open it:

```sh
mkdir -p /tmp/KeelNotification.app/Contents/MacOS
go build -o /tmp/KeelNotification.app/Contents/MacOS/KeelNotification ./examples/notification
cat > /tmp/KeelNotification.app/Contents/Info.plist <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>com.keel.notification-demo</string>
<key>CFBundleName</key><string>KeelNotification</string>
<key>CFBundleExecutable</key><string>KeelNotification</string>
<key>CFBundlePackageType</key><string>APPL</string>
</dict></plist>
PLIST
codesign --force --deep --sign - /tmp/KeelNotification.app
open /tmp/KeelNotification.app
```

This temporary app package is for development validation; official distribution requires a healthy signature and a trusted installation location. Click "Apply for Notification Permission", allow it to be delivered, and switch to other applications to observe the notification center; submit it again to verify the replacement, and then click "Withdraw Task Notification" to verify that it disappears. Notifier serializes system requests and displays real errors; "In-app and system notification" mode verifies in-app timeouts do not withdraw system notifications. `-check` only prints the platform/package identity check results, does not apply for permissions, and does not send notifications.

The example retains the last 12 events, displayed in order of occurrence, and the window content is scrollable. When clicking on the system notification in both modes, you can check the sequence of "request window brought to front → in-app notification closed → corresponding task opened"; if the in-app notification has timed out, the closing record will not be generated repeatedly. Subsequent withdrawal results will not overwrite these records. System-only mode does not have an in-app close callback; requesting prefix logging only proves that the call was made, and you still need to observe whether the window actually appears in the foreground.

The automatic test covers the parameter verification and the denied path of the unpackaged process, and does not pop up the system permission window. System authorization and real notification display/replacement/withdrawal still require the above manual acceptance.

Interface basis: [Apple UNUserNotificationCenter](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter).


Linux uses the running `org.freedesktop.Notifications` service. `Available()` will check the session bus and service capabilities, and may wait for a bus response; `RequestPermission` performs the same check asynchronously without popping up the permission dialog box. If there is no service, ErrUnsupported will be returned. If the protocol call fails, ErrFailed will be returned. Notify uses the system's default timeout, no icon/action, and application name Keel; the text is escaped according to the body-markup capability and the plain text meaning is retained.

The business ID is mapped to the numeric ID returned by the service. Repeated Post uses replaces_id, and Remove calls CloseNotification. Delete the mapping after receiving NotificationClosed; clear the old mapping when the service owner changes to avoid sending the old ID to the restarted service. Mappings are only retained within the current process and old notifications cannot be initiated across processes. Linux has implemented the ActionInvoked default action callback; the kit sample connection window Raise has passed the ActivationToken to OnActivate; the window back-end consumption token and actual placement are still pending access and desktop acceptance. Real desktop demonstration still requires running the sample for acceptance. Protocol testing uses controllable bus surrogates and does not represent completion of Linux desktop acceptance.

Agreement based on: [Freedesktop Desktop Notifications](https://specifications.freedesktop.org/notification/latest/protocol.html).


`Message.OnClick` can be set on macOS. After receiving the system default open action, the callback is executed once in an independent goroutine; UI modification requires `core.Update`. The example "native click callback" can be manually verified. Windows tray notifications and Linux services that declare actions also support OnClick; Linux services and other platforms that do not support actions return ErrUnsupported and will not silently discard callbacks. Ordinary notifications are still delivered according to the capabilities of each platform.

The first Post installs this module's delegate in the main queue; when there is another delegate, it returns ErrConflict and does not overwrite it. The app should not replace the delegate later. Only notifications marked with this module will request foreground display or trigger a callback, and other notifications will not be processed. Here only the response after delivery by the current process is received, and cold start/cross-start recovery is not implemented.

Use the new callback for the same ID replacement; the old callback will be restored if the replacement fails, and the late result of the old request will not overwrite the updated registration. Callback released after click, successful Remove, or replacement without callback. Manually ignoring/closing the banner in the notification center may not result in the default opening action. The app should remove the banner in a timely manner and release the remaining callbacks. Remove cannot cancel the execution of a callback after it has started execution.

Verification scope: compilation, unpackaged process rejection path, registry replacement/rollback/concurrent consumption and race testing; real notification banner and system click acceptance have not yet been completed.


The Linux default action uses the `default` identifier, listens for ActionInvoked and consumes the callback corresponding to the business ID once; it only accepts signals from the current service owner. NotificationClosed, successful Remove, replacement without callback, and service restart all release the registration. The signal uses a sequential processor to retain the processing order after opening and closing; the callback is executed in an independent goroutine outside the connection lock and can continue to be delivered or withdrawn. Failed replacement does not lose the original callback. ActivationToken is passed to OnActivate as optional activation data, as detailed below; receiving the token does not mean that it has been successfully prepended.

2026-10-03 Development acceptance: The sample window and event record of the temporary signature .app display normally; the permission request returns `native: operation failed: status 7`. Therefore, the successful authorization, system banner, replacement/withdrawal or click forwarding were not verified this time; after retaining the system error details, it reappeared as `UNErrorDomain (1): Notifications are not allowed for this application`, which has been corrected to a permission error classification. This information cannot alone distinguish system settings, application identity or signature issues, and successful authorization is still pending acceptance.

Error classified by: [Apple notificationsNotAllowed](https://developer.apple.com/documentation/usernotifications/unerror/code/notificationsnotallowed).

## Activation token

`Message.OnActivate(func(Activation))` receives the default open action and optional `Activation.Token`; if OnClick is set at the same time, OnActivate will be called first on the same callback goroutine, and then OnClick will be called. Platforms/notification services without tokens pass empty strings and still respond to clicks normally.

Linux temporarily stores the ActivationToken signal according to the notification ID, and takes it out and consumes it once when ActionInvoked arrives. Only signals of the current service owner, specified object path, and correct type/length are accepted; tokens are released by shutdown, recall, successful replacement, and service owner updates. Failed replacement/withdraw retains the original notification status. The unknown action consumes its prepended token but does not call the default open callback.

[freedesktop notification protocol](https://specifications.freedesktop.org/notification/latest-single/#signals) specifies that the token may or may not be sent before ActionInvoked; it may be an X11 startup ID or a Wayland xdg-activation token. Applications should use this opacity value on a per-window backend. This interface does not change process environment variables, nor does it activate the window itself; currently Keel Window.Raise does not accept tokens yet.

Protocol alias testing covers isolation, optional tokens, signal validation, single consumption, and lifecycle; these tests do not represent pre-acceptance with Linux Notification Desktop and Wayland.
