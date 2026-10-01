# MessageScroller

对话的滚动区域。

```go
sc := kit.MessageScroller(func(cx *el.Context) []el.Element { return renderMessages(cx) }).
    OnReachTop(loadOlder)
// 发送消息后：
sc.ScrollToEnd()
// 在 loadOlder 里把旧消息插到开头之后：
sc.HistoryPrepended()
```

- 视图位于底部时，新内容到达会自动跟随，流式回答保持可见。用户往上滚动后不再跟随，右下角出现"回到最新"按钮。
- 滚到顶部时调用 `OnReachTop`。插入旧消息后调用 `HistoryPrepended`，画面上的内容不会跳动（底层是 el 的 `KeepBottomOn`）。
- `ScrollToEnd()` 在下一帧跳到最新消息。`SetFollow(false)` 关闭自动跟随，比如从头展示一篇已完成的文档。
- 会撑满父容器给的空间。

Agent：容器角色 `log`，消息逐条单独列出。

验证：`go run ./examples/components -section message_scroller`，加 `-theme dark` 检查深色。
