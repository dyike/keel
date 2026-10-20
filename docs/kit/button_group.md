# ButtonGroup

English | [简体中文](button_group.zh-CN.md)

```go
kit.ButtonGroup(
    kit.Button("Previous page", prev).Outline(true),
    kit.Button("Next page", next).Outline(true),
).Name("Pagination")
```

- Several buttons are connected into one control: only the outer corners are rounded, and a 1dp background gap separates every neighbour without drawing dividers or strokes.
- Each button retains its own click, icon, loaded, and selected states; `Buttons()` removes the button for later modification.
- `Vertical(true)` is arranged vertically, with the upper and lower ends of the circle; `SetDisabled(true)` disables all buttons in the group.
- The Agent role is `group` and its name comes from `Name`.

Verification: `go run ./examples/components -section button_group`.
