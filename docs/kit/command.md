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

`Inline(true)` 将面板放入普通布局并打开；执行后保持显示，不阻挡外部控件，也不自动抢焦点。需要主动进入时调用 `palette.Focus(cx)`。`SetValue(false)` 或面板上的 Esc 隐藏内联内容；`Inline(false)` 关闭当前内容并恢复默认模态模式，下一次打开才显示弹层。切换会使旧异步请求失效。

`Searchable(false)` 隐藏搜索框、显示全部候选，并停止调用 OnSearch；键盘焦点移到面板框，方向键导航、Enter 执行。切换时清空查询、加载状态和错误，旧结果不再接收；恢复搜索且面板已打开时重新请求。默认搜索行为保留。

`Header(view)`、`Footer(view)` 分别放在搜索框上方和结果下方，加载、失败和空结果时仍显示；nil 移除。每个区域最多占窗口高度的五分之一且不超过 80dp，超出独立滚动；结果区保守预留这部分高度。`Empty(view)` 替换无匹配内容，nil 恢复默认。交互控件应复用实例，以保留焦点和内部状态。

`RenderItem(func(CommandItem, bool) el.View)` 自定义可见候选，第二个参数为当前高亮状态。它替换文字和快捷键展示，外层保留可访问名称、禁用及选中语义；nil 回调或 nil 内容使用默认行。点击展示内容执行命令，嵌套按钮独立处理操作。`RowHeight(dp)` 指定统一虚拟槽位高度，包含上下共 4dp 留白，分组标题共用此高度；0 恢复 36dp，正数最小为 5dp，负数及非有限值忽略。复杂内容需显式设置足够的高度，当前不自动测量每行。

```go
quick := kit.Command(items...).Searchable(false).Inline(true).
    Footer(kit.Button("刷新", refresh))
// 放入普通布局；应用可在事件中用 quick.Focus(cx) 将焦点交给面板。
root.Child(quick.Render(cx))
```

与上游仍有差异：Keel 使用统一行高；尚无完整的选择/确认/取消回调、keywords 和独立分隔项，快捷键仍为显式展示字符串；Esc 保持既有的直接关闭行为，而非先清除查询。内联模式可以放入应用自己的弹层，但外层弹层的关闭由应用管理。
