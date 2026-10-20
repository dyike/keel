# Toggle

English | [简体中文](toggle.zh-CN.md)

Buttons that remain pressed, such as "Bold" in the toolbar.

```go
bold := kit.Toggle("Bold", false).Icon(kit.IconStar).OnChange(func(on bool) { … })
```

- Click, Space, Enter to switch; `Value()` / `SetValue(bool)`; `SetDisabled`.

Agent: Role `toggle`, `selected` indicates whether to press or not.

Verify: `go run ./examples/components -section toggle`, add `-theme dark` to check the dark theme.

All variants are borderless. `ToggleGhost` and `ToggleOutline` use a transparent background when not selected; `ToggleDefault` uses a `Subtle` fill. Selected toggles use `Highlight`, and keyboard/programmatic focus draws the inset focus ring without changing layout.

`Size(...)` accepts `ToggleSizeXSmall`, `ToggleSizeSmall`, `ToggleSizeMedium`, `ToggleSizeLarge`, and adjusts the height, font size, icon and horizontal blank synchronously. The heights are 24/28/32/40dp respectively, and the default is Medium. Switching styles during operation retains the focus and selection status and does not trigger OnChange. Illegal enumeration values are ignored.
