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

连续同组命令前显示分组标题；标题不可选中。`CommandItem.Disabled` 禁止点击和执行，键盘导航跳过禁用项。构造和 `SetItems` 复制切片，应用修改原切片不会改变面板。结果用虚拟列表构建可见行，搜索结果在条目或查询不变时缓存；↑ ↓ 和 PageUp/PageDown 导航并滚动露出高亮，Home/End 保留搜索框的文字编辑行为。关闭后焦点回到触发控件，重新打开重置查询、高亮和滚动位置。`SetDisabled(true)` 关闭并禁止打开。

`OnSearch(func(query string, token uint64))` 切换为异步搜索模式，打开、改词和重试都会生成新 token。工作线程取得结果后通过 `core.Update` 调用 `SetResults(token, items...)` 或 `SetSearchError(token, message)`；返回 false 表示结果已过期或面板已关闭。旧查询不会覆盖新查询，失败后可重试。异步结果按返回顺序展示，不再做本地模糊过滤，适用于语义搜索。加载或失败状态不能执行旧命令。示例 `-section command_async` 输入 error 模拟失败；静态 `-section command` 包含万条结果与禁用项。
