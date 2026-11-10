# AppShell

English | [简体中文](app_shell.zh-CN.md)

A desktop layout with a collapsible left sidebar, a 56dp top title bar, main content and an optional pinned composer. Keep the shell in application state.

```go
shell := kit.AppShell("Workspace").
    Page(kit.AppPage{
        ID: "projects", Title: "Projects", Icon: kit.IconHome,
        Create: func() el.View { return kit.Input("Project name") },
    }).
    Page(kit.AppPage{
        ID: "settings", Title: "Settings", Icon: kit.IconSettings,
        Create: func() el.View { return kit.Switch("Notifications", true) },
    })
window.Open(window.Options{Title: "Workspace", Width: 1000, Height: 680, Frameless: true, NativeTrafficLights: true,
    TrafficLightLayout: &window.TrafficLightLayout{Height: kit.AppShellTitleBarHeight, Left: 16, Spacing: 20},
    Content: el.Root(shell)})
```

`Create` runs once on first visit. Each page keeps its own element root, preserving control and scroll state. Inactive pages are not rendered, animated or exposed to input/automation. Cached pages remain allocated for the lifetime of the shell; use a bounded set of application destinations, not one destination per document.

The first enabled page is selected automatically. Empty/duplicate IDs are ignored. `SetValue(id)` silently selects an existing enabled page; `OnChange` runs for user navigation. `Value` reads the selected ID.

`AppPage.Actions` adds page-specific heading controls; `Actions(view)` adds persistent controls. `Footer(view)` supplies the navigation footer. `SidebarWidth(dp)` sets expanded width; the visible divider supports drag resizing within `SidebarLimits(min, max)` (200–420dp by default). `SidebarSize()` returns the current expanded width for custom navigation. A focused divider supports arrow keys, Home and End. The sidebar toggle sits beside the window controls, and collapse preserves the dragged width; `SetCollapsed` and `Collapsed` control the icon rail. Below 640dp the sidebar becomes an icon rail automatically, restoring the user’s expanded-window preference when widened. `CollapseBelow(dp)` changes the threshold; zero disables automatic collapse. Pages without an icon use a file icon.

Pages receive 24dp padding and a vertical scroll area by default. Set `AppPage.Fill` for editors or split panes that own their sizing and scrolling. A nil factory produces an empty page.

Peer-page changes fade in over 150ms: the old page disappears immediately so text does not blend. `Transition(duration)` changes this duration; zero disables it. Reduced motion skips the fade. There is no initial entrance animation or background timer. This is a flat application shell, not a history router or shared-element transition engine.

Run `go run ./examples/app-shell` for a complete application with retained input. Preview page transitions with `go run ./examples/motion -section pages`, and spring presets with `-section presets`. Run `go run ./examples/components -section app_shell` for the component gallery.

## Custom regions

| Method | Purpose |
| --- | --- |
| `SidebarHeader(view)` | Sidebar toolbar, such as new/search buttons |
| `Sidebar(view)` | Replace page navigation; own scrolling and read `Collapsed()` to adapt to the icon rail |
| `Header(view)` | Replace the title and description with interactive content such as a model selector |
| `Actions(view)` | Persistent title-bar actions; `AppPage.Actions` supplies page-specific actions |
| `Content(view)` | Replace the main content, owning padding and scrolling; nil restores page mode |
| `ContentFooter(view)` | Pin content such as a composer below the main body |
| `Footer(view)` | Sidebar footer, such as a profile entry |

Each region accepts an `el.View`, without editing shell rendering. `Content` creates a retained root at configuration time; keep view state and call it only to replace content. Sidebar width controls navigation and body positioning; the title bar independently reserves space for window controls.

## Desktop window behavior

Use `window.Options{Frameless: true}` for integrated window chrome. Windows and Linux share the same control sizes, order and top-right placement. macOS traffic lights stay fixed at the top left, including when the sidebar collapses. Custom toolbar buttons always follow them. Enable `NativeTrafficLights: true` and set `TrafficLightLayout.Height` to `kit.AppShellTitleBarHeight`, as above, to use the actual AppKit close, minimize and full-screen buttons. AppShell reserves their space and does not draw duplicates. AppKit owns their appearance, hover behavior, actions and accessibility. Without that option, the legacy drawn buttons remain available. The sidebar toggle is a Keel control with a sidebar glyph, not an AppKit button.

The title and blank area drag the window; double-click follows the platform title-bar behavior. Custom headers, actions and window controls are excluded from dragging and receive normal input. With a system title bar, the shell keeps its application header without duplicating window controls.

Applications with wide selectors or many actions should set an appropriate `MinWidth` and adapt custom views to compact layouts. The example uses a 560×420dp minimum. Platform-branch tests cover layout and control callbacks; cross-compilation does not establish Windows or Linux runtime behavior.

`HideSidebarOnCollapse(true)` removes navigation, its footer and the vertical divider when collapsed, giving content the full width. The title-bar toggle remains available, and expanding restores the previous width. The default `false` keeps the icon rail. This also applies to automatic collapse in narrow windows. A 1dp horizontal divider starts at the sidebar’s right edge when expanded and spans the entire window when collapsed. `examples/app-shell` defaults to complete hiding; its top-right collapse-mode selector switches between the icon rail and complete hiding at runtime, including while collapsed.
