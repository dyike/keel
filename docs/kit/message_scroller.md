# MessageScroller

按稳定消息 ID 虚拟化的对话滚动区，只构建可见消息及附近的预读区域。

```go
sc := kit.MessageScroller(ids, 120, func(cx *el.Context, i int) el.Element {
    return renderMessage(cx, messages[i])
}).OnReachTop(loadOlder)
// 插入历史、新增、删除或重排消息后更新 ID 顺序：
sc.SetKeys(ids)
// 数据加载后定位未读消息；也可在第一次渲染前调用：
found := sc.ScrollToMessage(unreadID)
_ = found // ID 不存在时返回 false，保留原定位请求
// 用户发送消息时明确跳到最新：
sc.ScrollToEnd()
```

IDs 按从旧到新排列，必须非空且唯一，重复值会在修改前 panic。构造和 SetKeys 都复制切片。第二个参数是未测量消息的估计高度（dp）；真实高度由内容决定，支持不同长度的回答和图片。

- 位于底部时自动跟随新增消息和流式增高；上滚阅读后保留当前消息及其屏幕位置，显示“回到最新”。
- SetKeys 插入历史时，按稳定 ID 保留阅读位置；可见区上方的消息增高，也会修正滚动位置。无需再调用 HistoryPrepended。
- 可见消息的高度每帧检查；更改屏幕外内容后调用 `Invalidate(ids...)`，不传 ID 则清除全部高度缓存。
- `ScrollToMessage(id)` 以最小滚动量显示消息；首次渲染前也可请求，超过视口高度的消息显示开头。消息跳转与 `ScrollToEnd` 以最后一次有效请求为准。未读标记由应用维护。
- `IsScrolledUp(cx)` 查询最近绘制的视口下方是否还有内容；`IsFollowingTail(cx)` 查询实际自动跟随状态。首次绘制前分别为 false 和默认 true，待处理的消息跳转会暂停跟随。
- `SetFollow(false)` 关闭自动跟随；构造后立即关闭会从头阅读，已显示时保留位置。显式 `ScrollToEnd` 仍可跳到末尾，但不修改这个开关。开启跟随时，手动滚回底部会恢复跟随。
- `OnReachTop` 在到达顶部时加载一批历史，停在顶部不会重复调用。加载后至少离开顶部一个视口距离才重新允许触发，避免小批次插入与滚轮惯性形成重复请求。
- `SetDisabled` 禁用滚动、消息内容和“回到最新”，也暂停顶部加载回调。组件撑满父容器给定的空间。

消息中的 Markdown 仍由调用方保存为 Doc，选择范围和流式解析状态不会因虚拟化重建。单篇回答内拖选到视口边缘可继续滚动，释放后停止；不同消息的文本选择相互独立。

Agent：容器角色 log，当前可见的消息逐条列出。验证：`go run ./examples/components -section message_scroller`；完整流式与选择流程用 `go run ./examples/chat`。

“回到最新”按钮默认启用，`LatestButton(false)` 隐藏它而不改变滚动状态；`LatestLabel` 设置文字及可访问名称，空字符串恢复当前语言。`LatestRenderer` 接收每帧新建的默认 Button，可修改变体、图标、尺寸、Content 和 Appearance，也可返回另一 Button。组件复制返回值，保留内部 ID 和跳转动作；nil 配置或返回 nil 恢复默认。自定义内容限展示元素。

```go
sc.LatestLabel("查看新消息").LatestRenderer(func(b *kit.ButtonView) *kit.ButtonView {
    return b.Variant(kit.ButtonPrimary).Outline(true).Size(32)
}).LatestTransition(250 * time.Millisecond)
```

按钮默认使用 150ms 淡入淡出；`LatestTransition(0)` 立即切换，负时长忽略，减少动画优先。退出期间立即禁止交互，结束后移除。默认保留右下角文字按钮，与 GPUI 的圆形图标按钮外观不同。
