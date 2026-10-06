# ButtonGroup

English | [简体中文](button_group.zh-CN.md)

```go
kit.ButtonGroup(
    kit.Button("上一页", prev).Outline(true),
    kit.Button("下一页", next).Outline(true),
).Name("翻页")
```

- Several buttons are connected into a control: only the outermost corners are rounded; the stroked buttons share the middle border, and a thin gap is left between the solid buttons.
- Each button retains its own click, icon, loaded, and selected states; `Buttons()` removes the button for later modification.
- `Vertical(true)` is arranged vertically, with the upper and lower ends of the circle; `SetDisabled(true)` disables all buttons in the group.
- The Agent role is `group` and its name comes from `Name`.

Verification: `go run ./examples/components -section button_group`.
