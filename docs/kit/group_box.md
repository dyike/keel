# GroupBox

English | [简体中文](group_box.zh-CN.md)

`kit.GroupBox("Notification settings").Description("Choose how to receive notifications").Child(views...)` displays the title and description outside the content box, with borders, rounded corners, and a Surface background.

The constructor only accepts a title. Child accepts el.View and appends it to the content area, and SetChildren replaces the content in batches. Call the subview's Render every frame, preserving subview interaction. SetTitle modifies the title. Agent role group, titled as name, description and content listed separately. The component itself has no keyboard operation.

`Variant` accepts GroupBoxSurface (default, retains Surface background and borders), GroupBoxNormal (no background/borders), GroupBoxFill (Subtle background), and GroupBoxOutline (transparent background and borders). Each appearance preserves content padding.

`Footer(el.View)` adds bottom description or operation outside the content box, aligns with the left side of the title, and has a spacing of 8dp; it inherits the small Muted text by default, and passes nil to remove it. Title, description, and footer are not affected by body styles.

`TitleStyle(func(*el.TextEl))` modifies the title font size, color, font weight, spacing, etc.; `ContentStyle(func(*el.DivEl))` modifies the text background, border, rounded corners, padding and layout after the appearance default value. The callback is executed every frame and can read the current theme; it only modifies the incoming elements and does not retain element references. Pass nil to restore default. The body container maintains a stable identity, and deleting the title or updating the footer does not rebuild the internal input state.

```go
box := kit.GroupBox("Notification settings").Variant(kit.GroupBoxFill).
    Child(kit.Switch("Email notifications", true)).
    Footer(el.ViewFunc(func(*el.Context) el.Element {
        return el.Text("Settings apply to this device only")
    })).
    ContentStyle(func(e *el.DivEl) { e.P(24).Rounded(theme.RadiusLg) })
```

Verification: `go run ./examples/components -section group_box -theme dark`, omit theme to see the light theme.
