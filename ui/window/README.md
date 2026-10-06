# ui/window

English | [简体中文](README.zh-CN.md)

Window: open, close, bring to front (including activation token), window shortcut keys, event loop, off-screen screenshots, as well as read system preferences and complete platform events that Gio does not have.

| File | Responsibility |
| --- | --- |
| `window.go` | `Open`, `Main`, `Options` (including `Overlay`), `Window` (`Raise`, `Activate`, `WaylandDisplay`…) |
| `shortcut.go` | Shortcut key analysis and distribution |
| `icon*.go` | `SetIcon`: The runtime application icon (macOS Dock, Windows window, X11 `_NET_WM_ICON`), the shape is cut out by `internal/appicon` according to the platform specification |
| `root.go` | Window root view: background, scroll, 24dp margins |
| `position_*` | First display centered; macOS calculated based on available screen area, other platforms use Gio actions |
| `screenshot.go` | `Screenshot` Off-screen rendering to PNG |
| `decorations*.go` | Linux compositor is drawn by Keel when not drawing title bar |
| `titlebar_darwin.*` | Title bar drag area of macOS borderless window |
| `activation_*` | `Activate(token)`: Wayland uses xdg-activation, X11 writes the startup ID and then requests activation |
| `motion_*` | System preferences: reduce dynamic effects (macOS), scrollbar auto-hide (macOS, Windows), write in `theme` |
| `scroll_darwin.m`, `scroll_wayland*` | The device and gesture stages of scrolling (trackpad hand lift, scroll wheel) are not provided by Gio and are left to `core.ReportScrollGesture` |
| `automation.go` | Automation mode: memory window, semantic snapshot, simulated click input scrolling |
| `automation_server.go` | Automation protocol: JSON request on `KEEL_AUTOMATION` socket |
| `testdata/raise` | Real window deadlock regression testing |

- **Dependencies**: `core`, `theme`, `internal/appicon` (icon shape, shared with scaffolding), and `jezek/xgb` (X11 activation) and libwayland-client (Gio natively linked) on Linux. Does not rely on `el`, `kit`: the window only recognizes the `core.Widget` interface, and the system preferences are handed over to el through `theme`.
- **Used by**: Application code. `cmd/keel-mcp` drives it through the socket protocol and does not reference its code.

When the environment variable `KEEL_AUTOMATION=1` (or socket path) is set to start the application, each window will have an additional shadow window for Agent operations; the real window will be displayed as usual, and the Agent's operations will be reflected on the screen in real time. Adding `KEEL_HEADLESS=1` will not display the window, see [Agent end-to-end test](../../docs/automation.md).

```go
window.Open(window.Options{Title: "Hello", Content: page})
window.Main() // Exit the process after the last window is closed
```

Before changing the code here, read [Architecture · Cannot wait for the main thread](../../docs/architecture.md#avoid-waiting-for-the-main-thread-while-holding-the-lock) in the lock. See [Window and Application](../../docs/app.md) for details.

`Main` for macOS subscribes to NSWorkspace's accessibility display preferences and scrollbar styles, reading "Reduce Dynamic Effects" and "Show Scrollbars" on startup, and updating `theme.ReducedMotion` and `theme.SystemScrollbarsAutoHide` when they change. The AppKit callback is handed to the background consumer through a bounded queue, and then the topic is updated in `core.Update` to avoid the main thread waiting for the frame lock. "Auto-hide scroll bars" is read once when Windows starts. Windowless and off-screen automation modes do not install native observers, Linux currently uses application settings.
