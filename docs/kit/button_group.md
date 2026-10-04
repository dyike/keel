# ButtonGroup

```go
kit.ButtonGroup(
    kit.Button("上一页", prev).Outline(true),
    kit.Button("下一页", next).Outline(true),
).Name("翻页")
```

- 几个按钮连成一个控件：只圆最外侧的角；描边按钮共用中间的边框，实心按钮之间留一道细缝。
- 每个按钮保留自己的点击、图标、加载和选中状态；`Buttons()` 取出按钮在之后修改。
- `Vertical(true)` 竖向排列，圆上下两端；`SetDisabled(true)` 禁用组内全部按钮。
- Agent 角色为 `group`，名字来自 `Name`。

验证：`go run ./examples/components -section button_group`。
