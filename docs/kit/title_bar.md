# TitleBar

English | [简体中文](title_bar.zh-CN.md)

An application title bar for borderless windows. macOS can use actual system traffic lights while the application supplies its title and toolbar controls.

```go
bar := kit.TitleBar("Editor").
    Leading(toggleSidebar).
    Trailing(searchBox, shareButton)

window.Open(window.Options{Title: "Editor", Frameless: true, NativeTrafficLights: true,
    TrafficLightLayout: &window.TrafficLightLayout{Height: kit.TitleBarHeight, Left: 12, Spacing: 20},
    Content: el.Root(app)})
// Put bar.Render(cx) at the top of the app's Render
```

- **Drag**: Press and hold the blank space of the title bar to drag the window, and the area where the title text is located is also considered blank. Dragging is done by the system, so behaviors such as window snapping and cross-screen movement are consistent with native windows.
- **Window Button**:
  - macOS: the example enables AppKit’s native close, minimize and full-screen buttons. TitleBar only reserves their space; AppKit owns appearance, hover behavior and actions. Without `NativeTrafficLights`, the legacy drawn buttons remain, with Zoom as the third action. The title is centered.
  - Windows, Linux: On the right side are minimize, maximize/restore, and close. The close button displays red when hovered. The title is to the left.
  - Drawn buttons call `Minimize`, `ToggleMaximize`, and `Close` of the window where it is located through `core.CurrentWindow()`, and automatically switches to "Restore" when the maximized status changes.
- **Apply your own content**: The content of `Leading` is placed behind the window button, and the content of `Trailing` is placed on the right. These contents are outside the drag area and can be clicked and entered normally.
- When the window is not `Frameless`, the system title bar is still there. The TitleBar only draws the title and application content, which is equivalent to a page header. It does not display window buttons and does not register the drag area.
- Height is `kit.TitleBarHeight` (38dp).

Agent: The role of the title bar is `banner`, and the name is the title; the window buttons are named "Close", "Minimize", "Maximize" or "Restore", and the application content is listed separately.

Verification: `go run ./examples/frameless -lang en` opens a real borderless window for dragging the title bar and clicking window buttons; `go run ./examples/components -section title_bar -lang en` shows the header style in an ordinary window.

Title uses secondary text color when window is out of focus, macOS window buttons are grayed out; restored after reactivation. Slices for front and rear slots are copied, and empty slots are ignored.

macOS double-click only works on the currently drawn middle drag area, excluding front and rear slots and window buttons. The native local event monitor handles double-clicks before Gio initiates the second drag: minimize when the system is set to minimize, do nothing when set to do nothing, and toggle zoom otherwise. The registration is removed when the title bar is hidden, the area to which it belongs is disabled, or the window is destroyed. Automated windows test scaling via normal double-click callbacks; real window dragging is still left to AppKit.

Native traffic lights appear in system accessibility, not Keel off-screen snapshots. Double-clicks on Gio controls in the title-bar band but outside its drag region are delivered directly to Gio, preventing transparent title-bar zoom from also firing; system buttons keep AppKit dispatch. On 2026-11-10, an Apple silicon real window verified search-button double-click without zoom and sidebar drag/collapse restoration. Native-window tests check standard NSWindow button identity, actions and fixed positions after collapse/resize.

Repeated clicks use the same boundary: the third and later clicks cannot fall through to system zoom. Only the second click in the blank drag area performs the double-click action; controls continue receiving normal clicks. Both AppShell collapse modes were checked in a real macOS window with three and six consecutive clicks, preserving window size.

Known validation gap: the full real-window regression on 2026-11-10 still reported native buttons becoming hidden or being repositioned after programmatic resizing, and a 2.5dp horizontal centering offset. Passing repeated-click checks does not resolve these issues; the full native-window regression is not passing.
