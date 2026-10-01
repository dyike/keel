# Command

命令面板：一个搜索框加一列命令，输入时逐步筛选。

```go
palette := kit.Command(
    kit.CommandItem{Title: "新建订单", Group: "订单", Shortcut: "mod+n", Action: newOrder},
    kit.CommandItem{Title: "打开设置", Action: openSettings},
)
// Render 中：
cx.Shortcut("mod+k", palette.Toggle)
root.Child(palette.Render(cx))
```

- 匹配规则：前缀匹配排第一，其次是子串匹配，再次是"字符按顺序出现"的模糊匹配，例如"设置"能找到"打开设置"，"nwo"能找到"New window"。
- ↑ ↓ 移动高亮，回车执行，Esc 或点击外部关闭。执行命令时先关闭面板，再调用 Action。
- 打开时焦点在搜索框里。面板是模态的，靠近窗口顶部显示。
- `Shortcut` 只用于显示，不会注册快捷键。快捷键要像上面的例子一样，自己用 `cx.Shortcut` 绑定。
- `Toggle`、`Value()` / `SetValue(bool)`、`SetItems`。需要 `el.Root`。

Agent：面板是名为"命令面板"的 `dialog`，搜索框是 `textbox`，命令是 `option`，`selected` 表示当前高亮。

验证：`go run ./examples/components -section command`，加 `-theme dark` 检查深色。
