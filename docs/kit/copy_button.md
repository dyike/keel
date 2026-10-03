# CopyButton

把文字复制到剪贴板，并显示"已复制"作为反馈。

```go
kit.CopyButton(func() string { return order.ID })
```

- 文字在点击时才读取，所以函数总能拿到最新内容。
- 复制后，按钮显示"已复制"和对勾图标，`kit.CopiedFeedback`（1.5 秒）后恢复为"复制"。
- 按钮是 Ghost 样式，高 28dp，支持 Tab 聚焦和 Space / Enter 触发。

Agent：角色是 `button`，名字在"复制"和"已复制"之间切换。

验证：`go run ./examples/components -section copy_button`。

连续点击会在每次复制后重新保留完整的 1.5 秒反馈，不沿用上次的截止时间。`SetDisabled(true)` 禁用复制并清除反馈；祖先禁用也会阻止读取文字和写剪贴板。传入 nil 取值函数时点击不会显示“已复制”。

`OnCopied(func(string))` 在提交剪贴板写入请求后调用，参数是本次提交的原文，不会再次读取取值函数。系统剪贴板没有成功确认信号，因此此回调不代表操作系统已确认写入。禁用或取值函数为 nil 时不调用；传入 nil 可移除回调。

`Content(el.View)` 替换默认图标和文字，保留按钮的键盘操作及“复制”/“已复制”语义名称。内容应为文字、图标等展示元素，避免嵌套输入框或按钮。自定义内容最小高 28dp，可随内容增高；传入 nil 恢复默认按钮。

```go
copy := kit.CopyButton(func() string { return order.ID }).
    OnCopied(func(value string) { lastCopied = value })
copy.Content(el.ViewFunc(func(cx *el.Context) el.Element {
    label := "复制订单号"
    if copy.Copied() {
        label = "订单号已复制"
    }
    return el.Text(label)
}))
```

`Copied()` 提供当前反馈状态。自定义内容与默认按钮共用 1.5 秒计时；每次复制重新计时，`SetDisabled(true)` 清除反馈。
