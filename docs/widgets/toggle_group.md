# 状态按钮组

```go
align := widget.ToggleGroup("左","中","右").OnChange(onChange)
align.SetValues("中")
format := widget.ToggleGroup("粗体","斜体","下划线").Multiple().Ghost()
format.SetItemDisabled(1,true)
```
默认单选；再次激活当前项会取消。Multiple 可选择任意子集。`Values` 返回显示顺序的新切片；`SetValues` 静默更新、忽略未知值，单选只保留显示顺序中首个匹配项。重复选项被忽略。方向键跳过禁用项，Home/End 跳到首尾，Space/Enter 切换；`SetDisabled` 禁用整个组。`Size`、`Ghost` 配置所有组内按钮。

验证入口：`go run ./examples/components -section toggle_group`。
