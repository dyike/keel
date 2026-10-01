# Notifier

在窗口右上角堆叠显示通知。

```go
n := kit.Notifier()
// Render 中：作为根视图最外层元素的直接子元素
root.Child(n.Render(cx))
// 回调中：
n.Notify(kit.Notice{Title: "保存成功", Body: "订单已更新", Tone: kit.ToneSuccess})
// 后台 goroutine 中：
core.Update(func() { n.Notify(kit.Notice{Title: "同步完成"}) })
```

- 超时：`Timeout` 为 0 时用 `kit.NotificationTimeout`（5 秒）；为负数时不自动消失。
- 悬停：指针停在通知上时，这条通知不会消失；移开后重新计时。
- 数量：最多同时显示 `kit.MaxNotifications`（5）条，超出的排队，前面的消失后依次显示。
- `Notify` 返回 id，`Dismiss(id)` 移除指定通知（不论在显示还是排队），`Len()` 返回显示加排队的总数。
- 每条通知都有关闭按钮。通知不抢焦点，也不处理 Esc，下面的对话框仍然可以用 Esc 关闭。
- `Notify`、`Dismiss` 必须在 UI 帧锁内调用，也就是在回调里，或者用 `core.Update` 包起来。需要 `el.Root`。

Agent：每条通知的角色是 `status`，名字是标题，`value` 是 Tone 名称（neutral/info/success/warning/danger）；关闭按钮名为"关闭 标题"。

验证：`go run ./examples/components -section notifier`。
