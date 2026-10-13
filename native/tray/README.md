# native/tray

English | [简体中文](README.zh-CN.md)

Desktop status icons and native popup menus. Depends only on `native` and `native/internal/sys`, never UI or Gio. Applications may combine it with `ui/window` or run a standalone event loop.

```go
icon, err := tray.New(tray.Options{
    Icon: artwork, // image.Image, 1..1024 pixels per axis; copied as PNG
    Template: true,
    Tooltip: "My app",
    Menu: []tray.MenuItem{
        {ID: "open", Title: "Open", OnSelect: func() {
            core.Update(openWindow)
        }},
        {Separator: true},
        {ID: "quit", Title: "Quit", OnSelect: window.Quit},
    },
})
if err != nil { log.Fatal(err) }
window.MainWithOptions(window.MainOptions{
    KeepAlive: true,
    OnQuit: func() { icon.Close() },
})
```

`MenuItem` supports `Children`, `Disabled`, `Checked`, `Separator` and `OnSelect`. Optional string IDs must be unique across the entire tree. Submenus cannot also have an action. Menu trees are copied, limited to 1024 entries and 16 nested levels. Labels and tooltips cannot contain NUL. Setters do not call user callbacks. `SetMenu` replaces the complete menu, invalid updates preserve it, and native IDs are never reused so stale menu events cannot invoke replacement actions.

`SetIcon`, `SetTooltip`, `SetMenu` and `Close` are safe from concurrent goroutines. Callbacks run serially per tray on a Go goroutine, outside native event threads and the UI frame lock. UI writes still require `core.Update`. Callbacks may update or close their own tray; keep long tasks on another goroutine. Close is idempotent and discards queued events; an already running callback may finish. It removes the icon and does not exit the process.

| Platform | Backend and behavior |
| --- | --- |
| macOS | purego calls `NSStatusBar`/`NSMenu`. `Template` marks the image for system appearance tinting. A click opens the attached menu; without a menu it calls `OnActivate`. `New` synchronously uses the main thread: call before `window.Main`, or in a background goroutine outside the frame lock. Setters and Close enqueue changes on the main queue and return before application. |
| Windows | Win32 notification icon, a shared hidden top-level HWND and a sleeping message loop. Activation calls `OnActivate`; context activation opens the native menu. Explorer's `TaskbarCreated` broadcast re-adds live icons. Native resources are released on replacement and removal. |
| Linux | Pure Go session D-Bus, SNI (KDE and freedesktop namespaces) and DBusMenu. Requires a running StatusNotifierHost; without one New returns `native.ErrUnsupported`. Menus and activation follow the host's behavior. Watcher restart triggers re-registration. Works independently of X11/Wayland window handles. |
| Other platforms, iOS, Android, Web | New returns `native.ErrUnsupported`. |

`Template` only affects macOS. Tooltips, icon scaling and placement are controlled by the desktop; Windows tooltips are limited to 127 UTF-16 units. A registered icon may be in the notification overflow or hidden by the desktop. No permissions or cgo are needed.

For a standalone macOS app, call `runtime.LockOSThread()` at the start of `main`, create the tray there and call `tray.Main()` on that same main thread. On Windows/Linux, `tray.Main()` keeps the process alive while the backend handles native events. Do not run `tray.Main` together with `window.Main`. Standalone applications own their exit policy and cleanup; tray does not import a UI lifecycle manager.

Run `go run ./examples/tray`. It starts without a window; use the menu to open one, close it and reopen it. `-smoke` automatically exercises icon/menu updates, windowless reopening and explicit exit. `-screenshot <path>` renders only its companion window. Windows/Linux runtime acceptance still requires those desktops; macOS smoke and cross-compilation are separate checks.
