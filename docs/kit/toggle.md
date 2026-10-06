# Toggle

English | [简体中文](toggle.zh-CN.md)

Buttons that remain pressed, such as "Bold" in the toolbar.

```go
bold := kit.Toggle("加粗", false).Icon(kit.IconStar).OnChange(func(on bool) { … })
```

- Click, Space, Enter to switch; `Value()` / `SetValue(bool)`; `SetDisabled`.

Agent: Role `toggle`, `selected` indicates whether to press or not.

Verify: `go run ./examples/components -section toggle`, add `-theme dark` to check the dark theme.

`Variant(kit.ToggleGhost)` sets the transparent borderless appearance when not selected; `ToggleOutline` sets the transparent background with borders. `ToggleDefault` retains the original Surface background and borders. The Highlight background color is displayed when selected; the focus border appears only when the Ghost is focused.

`Size(...)` accepts `ToggleSizeXSmall`, `ToggleSizeSmall`, `ToggleSizeMedium`, `ToggleSizeLarge`, and adjusts the height, font size, icon and horizontal blank synchronously. The heights are 24/28/32/40dp respectively, and the default is Medium. Switching styles during operation retains the focus and selection status and does not trigger OnChange. Illegal enumeration values are ignored.
