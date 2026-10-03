# Notifier

按位置分组显示窗口内通知，默认在右上角。

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
- 悬停：指针停在通知上时，这条通知不会消失；移开后继续剩余时间，键盘焦点进入通知或通知被禁用/遮挡时也暂停。
- 数量：每个位置最多同时显示 `kit.MaxNotifications`（5）条，超出的排队，前面的消失后依次显示。
- `Notify` 返回 id，`Dismiss(id)` 移除指定通知（不论在显示还是排队），`Len()` 返回显示加排队的总数。
- 每条通知都有关闭按钮。通知不抢焦点，也不处理 Esc，下面的对话框仍然可以用 Esc 关闭。
- `Notify`、`Update`、`Dismiss` 必须在 UI 帧锁内调用，也就是在回调里，或者用 `core.Update` 包起来。需要 `el.Root`。

Agent：每条通知的角色是 `status`，名字是标题，`value` 是 Tone 名称（neutral/info/success/warning/danger）；关闭按钮名为"关闭 标题"。

验证：`go run ./examples/components -section notifier`。

`Update(id, Notice)` 原位替换通知、保留 ID 和队列位置，重新开始该条通知的超时；不存在的 ID 返回 false。后台任务用 `core.Update` 包住调用。队列中的通知从实际显示时才开始计时；窄窗口会限制通知栈宽高，长通知栈可以滚动。

`Notifier.Placement(position)` 设置容器默认位置，现有未指定位置的通知也会跟随移动。`Notice.Placement` 为单条通知覆盖位置；默认零值 `NoticeDefault` 跟随容器。容器传 NoticeDefault 恢复右上角，非法默认值忽略，单条非法值按跟随默认处理。

支持 NoticeTopLeft、NoticeTopCenter、NoticeTopRight、NoticeLeftCenter、NoticeRightCenter、NoticeBottomLeft、NoticeBottomCenter、NoticeBottomRight。每个位置独立排队，组内保持发送顺序从上往下排列；更新 Notice 可将其移入其他位置。移动默认位置不重置超时，Update 仍按原约定重启该条超时。

```go
n.Placement(kit.NoticeBottomRight)
n.Notify(kit.Notice{Title: "下载完成", Placement: kit.NoticeBottomLeft})
```

默认位置属于 Notifier 实例，不写入全局主题。多个位置的通知栈分别限制在窗口范围内，不自动避让其他栈；窄窗口同时启用多个位置时可能重叠。
