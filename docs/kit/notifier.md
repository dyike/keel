# Notifier

English | [简体中文](notifier.zh-CN.md)

Display notifications in the window grouped by location, default in the upper right corner.

```go
n := kit.Notifier()
// In Render: as a direct child element of the outermost element of the root view
root.Child(n.Render(cx))
// Callback in progress:
n.Notify(kit.Notice{Title: "Saved successfully", Body: "Order updated", Tone: kit.ToneSuccess})
// In the background goroutine:
core.Update(func() { n.Notify(kit.Notice{Title: "Sync complete"}) })
```

- If you don’t want to mount it yourself, use the notification container that comes with the window: `kit.WindowNotifier(cx).Notify(kit.Notice{...})`. It is created and hung at the root of the window when it is called for the first time. Afterwards, the same instance is returned to the same window. You can set the position, system delivery, etc. as usual, which is equivalent to GPUI's `window.push_notification`.
- Timeout: Use `kit.NotificationTimeout` (5 seconds) when `Timeout` is 0; it will not disappear automatically when it is a negative number.
- Hover: When the pointer is parked on the notification, the notification will not disappear; it will continue for the remaining time after moving away, and will also be paused when the keyboard focus enters the notification or the notification is disabled/occluded.
- Quantity: Each position can display up to `kit.MaxNotifications` (5) at the same time. If the queue exceeds the number, the previous ones will be displayed in sequence after they disappear.
- `Notify` returns the id, `Dismiss(id)` removes the specified notification (whether it is displayed or queued), `Len()` returns the total number of notifications in management; the pure in-app mode is display plus queuing, and the system mode also includes only system notifications and records in the application that have timed out but are waiting to be withdrawn.
- Each notification has a close button. The notification does not grab focus and does not handle Esc. The following dialog box can still be closed with Esc.
- `Notify`, `Update`, and `Dismiss` must be called within the UI frame lock, that is, in a callback, or wrapped with `core.Update`. Requires `el.Root`.

Agent: The role of each notification is `status`, the name is the title, `value` is the Tone name (neutral/info/success/warning/danger); the close button is named "Close Title".

Verification: `go run ./examples/components -section notifier`.

`Update(id, Notice)` replaces the notification, retains the ID and queue position in place, and restarts the timeout of the notification; returns false for non-existent IDs. Background tasks wrap calls with `core.Update`. Notifications in the queue start timing when they are actually displayed; a narrow window will limit the width and height of the notification stack, and a long notification stack can be scrolled.

`Notifier.Placement(position)` sets the default location of the container, and existing notifications with unspecified locations will also be moved. `Notice.Placement` overrides the position for a single notification; default value of zero `NoticeDefault` follows the container. The container passes NoticeDefault and restores the upper right corner. Illegal default values are ignored. A single illegal value is processed by following the default.

Supports NoticeTopLeft, NoticeTopCenter, NoticeTopRight, NoticeLeftCenter, NoticeRightCenter, NoticeBottomLeft, NoticeBottomCenter, NoticeBottomRight. Each position is queued independently, and the sending order within the group is maintained from top to bottom; updating the Notice can move it to other positions. Moving the default location does not reset the timeout, and Update still restarts the timeout according to the original agreement.

```go
n.Placement(kit.NoticeBottomRight)
n.Notify(kit.Notice{Title: "Download complete", Placement: kit.NoticeBottomLeft})
```

The default location belongs to the Notifier instance and does not write to the global topic. The notification stacks of multiple positions are limited to the window range and do not automatically avoid other stacks; when multiple positions are enabled at the same time in a narrow window, they may overlap.


`Notice.Content` accepts any `el.View` and replaces the Body if it is not nil. The Title is still displayed and serves as the accessible name of the notification; nil restores the normal text. Can place rich text, pictures or interactive controls. `Notice.Action` is an independent View slot below the text, which can hold a fully configured Button or combine multiple operations. The click operation will not automatically close the notification. Call `Dismiss(id)` in the callback when necessary.

```go
var id int
id = n.Notify(kit.Notice{
    Title: "Connection lost", Timeout: -1,
    Content: kit.Label("Check your connection and try again."),
    Action: kit.Button("Retry", func() {
        retry()
        n.Dismiss(id)
    }).Variant(kit.ButtonPrimary),
})
```

The body and action areas retain separate stable identities; Update can replace or clear both slots. When reusing a stateful View, keep the instance stable and do not put the same interaction instance into multiple notifications at the same time. Custom content should fit within the available width of the notification. When the focus enters any sub-control, the countdown will be paused, and disabling the container will disable the content and operations at the same time; queued notifications will only render content after they are actually displayed.

Different positions belong to different overlays; the keyboard focus of the sub-control is not retained when moving the notification to another position. Notifications that require continuous editing should remain fixed in position.


`Notice.OnClick` is executed when the notification background, title or normal text is clicked, and does not automatically close the notification. After setting, add a button operation area named after Title, support Tab focus and Enter/Space activation, and the outer layer still retains status semantics. Rich content and action buttons handle their respective clicks independently, and the close button does not perform an OnClick. Restore normal notifications after removing the callback.

`Notice.OnClose` Execute once synchronously after the in-app notification is dismissed, overriding close button, timeout, explicit Dismiss, and queued cancel (except system mode only); does not trigger for duplicate or unknown IDs. The callback can safely dismiss the same ID again, update other notifications, or add new notifications. Update replaces the callback but does not trigger a shutdown; final shutdown uses the latest callback. Uninstalling Notifier is not equivalent to Dismiss and does not trigger OnClose. The callback runs within the UI frame lock, the background work should be executed asynchronously, and core.Update is used when writing back.


`NotifyKey(key, Notice)` uses a business string to identify the notification, and its scope is limited to the current Notifier. Repeatedly sending the same non-empty key will replace it in place, return the original ID, retain the queue position and restart the timeout, and the old OnClose will not be triggered. Notice's body, action, callback, and position will be replaced with new values. An empty key is equivalent to a normal Notify and is added each time.

```go
n.NotifyKey("download/report", kit.Notice{Title: "Downloading", Timeout: -1})
n.NotifyKey("download/report", kit.Notice{Title: "Download complete", Tone: kit.ToneSuccess})
n.DismissKey("download/report")
```

`Update(id, Notice)` retains the business key; `DismissKey(key)` deletes the corresponding notification being displayed or queued and returns whether it is found. An empty key returns false. Resending the same key after deletion will assign a new ID and the old timeout will not affect new notifications. Different businesses should set their own key prefixes to avoid conflicts within the same container; Rust types should not be used as identifiers.

`Clear()` first removes all notifications when called, then executes their respective OnClose in the order of the original queue, and returns the removal quantity. The notification added in the callback is retained unless a subsequent callback explicitly deletes it; repeated clearing will not repeat the notification of the removed item. The above methods also require calling within the UI frame lock. See below for the withdrawal behavior of system notifications.


System delivery is accessed through `NoticeSystemBackend`, and the kit does not directly rely on native modules. The `Post(id,title,body,done)` and `Remove(id,done)` interfaces are executed in the work goroutine, and done must be called after each completion (also called on failure). Notifier serially waits for each completion before executing the next request; the backend must not modify the UI directly. See `examples/notification` for native adapters and permission requests.

```go
n := kit.Notifier().SystemBackend(backend, func(r kit.NoticeSystemResult) {
    // Already within the UI frame lock, r.Err can be displayed; Removing distinguishes between delivery and recall.
})
n.NotifyKey("download", kit.Notice{
    Title: "Download complete", Body: "report.pdf has been saved.",
    Delivery: kit.NoticeInAppAndSystem,
})
```

- `NoticeInApp` In-app only, `NoticeSystemOnly` System only, `NoticeInAppAndSystem` both. `NoticeDeliveryDefault` uses the container default; `Notifier.Delivery(mode)` only affects subsequent Notify/Update, and does not migrate existing notifications. The initial default is only within the application.
- The system body only takes Title/Body; rich content, buttons and Tone are not sent. System request is skipped if the title and body are empty. System mode does not automatically apply for permissions; applications should complete platform authorization first.
- When both are delivered at the same time, the in-app timeout will only hide the card and trigger OnClose once, retaining the system notification and business key; resending the same key will update the same system ID and redisplay the card. Dismiss/DismissKey/Clear is required to request withdrawal. Hidden cards will not trigger OnClose repeatedly. Only system notifications do not trigger OnClose and do not occupy the five visible queues in the application.
- The system ID uses random container prefixes and notification serial numbers to prevent different containers/processes from overwriting each other; cross-boot recovery is not possible, and cross-boot withdrawal requires the application to manage its own native interface.
- Existing notifications remain with the backend they were created with; replacing the SystemBackend only affects new notifications, with the resulting callback captured when the request is queued. Please configure the backend before notifying for the first time. Switching existing notifications to in-app only, or updating to empty system text, will recall the previous system notifications.
- If the backend is missing, `ErrNoticeSystemUnavailable` will be returned and will not be silently regarded as success. `SystemError()` is the error of the most recently completed request and will be cleared if successful; the result callback is executed in subsequent UI frames. When delivery fails, the in-app card will still be displayed in both modes; only the system mode will not automatically change to in-app.
- Closing the window/uninstalling Notifier does not automatically withdraw system notifications; clear Clear explicitly if needed. If the backend does not call done, subsequent system requests for the container will be blocked.

Backends that support click responses can implement NoticeSystemInteractiveBackend, see below. The current test verifies the delivery state machine and request sequence, and is not equivalent to the actual display acceptance of macOS/Linux Notification Center.


`NoticeSystemInteractiveBackend` Adds `PostInteractive(id,title,body,activated,done)` to the base backend. After the backend receives the system open action and calls activated, the kit automatically switches back to UI frame processing; even if Notice.OnClick is empty, the open action will be processed. Ordinary backends continue to deliver via Post and cannot provide system click responses.

`OnSystemActivate(fn)` sets the application's window evocation callback, such as `func(){ if !w.Closed() { w.Raise() } }`. In response, first remove the notification from the management list, schedule the system recall, and then call the window awakening, the in-app OnClose that has not yet been executed, and the latest Notice.OnClick in sequence. Duplicate responses, deleted notifications, and notifications that have been switched to purely in-app are ignored. After timeout, the application can still respond to the system opening action, but OnClose will not be called repeatedly; OnClose will not be called in system mode only.

The example selects a native adapter that supports clicks on macOS/Linux/Windows; Linux services require actions capabilities. The example records the token in OnSystemActivation and calls `Window.Activate(token)` in OnSystemActivate: X11 writes the token to the window as the startup ID, and then requests activation with the timestamp in it. The window manager releases the token accordingly instead of just flashing the taskbar; Wayland hands the token to xdg-activation to activate the window's surface; when the synthesizer does not support it and on other platforms, it returns to Raise. Whether the window can actually be brought to the front is determined by the operating system, and native clicks and window awakening still need to be accepted by the real system.

`NoticeSystemActivationBackend` increments `PostActivated(id,title,body,activated,done)`, which activated receives `NoticeActivation{Token: ...}`. When the backend implements two interactive interfaces at the same time, the interface with activation data is preferred; the backend that only implements PostInteractive continues to work, and the activation data is empty.

`OnSystemActivation(func(NoticeActivation))` receives data in the UI frame in the following order: remove notification and schedule recall → OnSystemActivation → OnSystemActivate → OnClose that needs to be executed → OnClick. Tokens are no longer delivered when a duplicate event or notification is deleted. The application hands the token to its own window backend in this callback; the component does not modify environment variables or infer the token format.
