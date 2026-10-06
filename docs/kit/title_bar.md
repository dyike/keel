# TitleBar

English | [简体中文](title_bar.zh-CN.md)

The title bar of a borderless window is drawn by the application itself.

```go
bar := kit.TitleBar("Editor").
    Leading(toggleSidebar).
    Trailing(searchBox, shareButton)

window.Open(window.Options{Title: "Editor", Frameless: true, Content: el.Root(app)})
// Put bar.Render(cx) at the top of the app's Render
```

- **Drag**: Press and hold the blank space of the title bar to drag the window, and the area where the title text is located is also considered blank. Dragging is done by the system, so behaviors such as window snapping and cross-screen movement are consistent with native windows.
- **Window Button**:
  - macOS: The three circular buttons on the left are Close, Minimize, and Zoom. The colors follow the system convention. The symbols are displayed when the pointer moves up. The title is centered.
  - Windows, Linux: On the right side are minimize, maximize/restore, and close. The close button displays red when hovered. The title is to the left.
  - The button calls `Minimize`, `ToggleMaximize`, and `Close` of the window where it is located through `core.CurrentWindow()`, and automatically switches to "Restore" when the maximized status changes.
- **Apply your own content**: The content of `Leading` is placed behind the window button, and the content of `Trailing` is placed on the right. These contents are outside the drag area and can be clicked and entered normally.
- When the window is not `Frameless`, the system title bar is still there. The TitleBar only draws the title and application content, which is equivalent to a page header. It does not display window buttons and does not register the drag area.
- Height is `kit.TitleBarHeight` (38dp).

Agent: The role of the title bar is `banner`, and the name is the title; the window buttons are named "Close", "Minimize", "Maximize" or "Restore", and the application content is listed separately.

Verification: `go run ./examples/frameless` open a real borderless window, drag the title bar and click the window button; `go run ./examples/components -section title_bar` check the header style in the ordinary window.

Title uses secondary text color when window is out of focus, macOS window buttons are grayed out; restored after reactivation. Slices for front and rear slots are copied, and empty slots are ignored.

macOS double-click only works on the currently drawn middle drag area, excluding front and rear slots and window buttons. The native local event monitor handles double-clicks before Gio initiates the second drag: minimize when the system is set to minimize, do nothing when set to do nothing, and toggle zoom otherwise. The registration is removed when the title bar is hidden, the area to which it belongs is disabled, or the window is destroyed. Automated windows test scaling via normal double-click callbacks; real window dragging is still left to AppKit.

Verification record (2026-10-02): Off-screen double-click, drag area exclusion control, disabling cleaning and out-of-focus pixel tests passed. When running the native example in the current environment, Gio crashes during the DisplayLink creation phase; the pre-change version also fails. Therefore, AppKit double-click preference and system drag still need to be reviewed in an environment that can run desktop windows.
