# MessageScroller

按稳定消息 ID 虚拟化的对话滚动区，只构建可见消息及附近的预读区域。

```go
sc := kit.MessageScroller(ids, 120, func(cx *el.Context, i int) el.Element {
    return renderMessage(cx, messages[i])
}).OnReachTop(loadOlder)
// 插入历史、新增、删除或重排消息后更新 ID 顺序：
sc.SetKeys(ids)
// 用户发送消息时明确跳到最新：
sc.ScrollToEnd()
```

IDs 按从旧到新排列，必须非空且唯一，重复值会在修改前 panic。构造和 SetKeys 都复制切片。第二个参数是未测量消息的估计高度（dp）；真实高度由内容决定，支持不同长度的回答和图片。

- 位于底部时自动跟随新增消息和流式增高；上滚阅读后保留当前消息及其屏幕位置，显示“回到最新”。
- SetKeys 插入历史时，按稳定 ID 保留阅读位置；可见区上方的消息增高，也会修正滚动位置。无需再调用 HistoryPrepended。
- 可见消息的高度每帧检查；更改屏幕外内容后调用 `Invalidate(ids...)`，不传 ID 则清除全部高度缓存。
- `SetFollow(false)` 从头阅读并关闭自动跟随；显式 ScrollToEnd 仍可跳到末尾。
- `OnReachTop` 在到达顶部时加载一批历史，停在顶部不会重复调用。加载后至少离开顶部一个视口距离才重新允许触发，避免小批次插入与滚轮惯性形成重复请求。
- `SetDisabled` 禁用滚动、消息内容和“回到最新”，也暂停顶部加载回调。组件撑满父容器给定的空间。

消息中的 Markdown 仍由调用方保存为 Doc，选择范围和流式解析状态不会因虚拟化重建。单篇回答内拖选到视口边缘可继续滚动，释放后停止；不同消息的文本选择相互独立。

Agent：容器角色 log，当前可见的消息逐条列出。验证：`go run ./examples/components -section message_scroller`；完整流式与选择流程用 `go run ./examples/chat`。
