# ToggleGroup

一排开关按钮。默认单选，`Multiple()` 后可多选。

```go
align := kit.ToggleGroup("左对齐", "居中", "右对齐")
style := kit.ToggleGroup("B", "I", "U").Multiple().OnChange(func(on []string) { … })
```

- 单选模式下，再点一次已按下的选项会取消选择。
- `Value()` 按选项顺序返回已按下的选项（返回副本），`SetValue(values...)` 不触发回调；单选模式只保留第一个。`SetDisabled`。

Agent：容器角色 `group`，每个按钮是 `toggle`。

验证：`go run ./examples/components -section toggle_group`，加 `-theme dark` 检查深色。
