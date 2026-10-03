# Message

对话中的一条消息：头像、内容、操作栏。

```go
kit.Message("我", text).User()                  // 用户消息：靠右的气泡，不显示头像
kit.Message("AI 助手", answerDoc).Actions(copy) // 其他人：头像 + 整宽内容 + 操作栏
```

- 非用户消息占满宽度，适合长篇 Markdown 回答。头像显示作者名字的首字母。
- `Actions(views...)` 显示在内容下方，比如复制、重试按钮。流式输出还没结束时，通常先不加操作。

Agent：每条消息是 `article`，名字是作者，内容和操作按钮单独列出。

验证：`go run ./examples/components -section message`，加 `-theme dark` 检查深色。

`SetState(MessageSending, "")` 显示发送中；`SetState(MessageFailed, reason)` 显示失败原因。`OnRetry(fn)` 在失败时显示重试按钮，点击先转为发送中，再调用一次业务回调；业务完成后设置 `MessageReady`，再次失败则设置 `MessageFailed`。

用户和助手消息都支持 `Actions`，传入切片会被复制，空槽位会被忽略。操作栏是否在流式输出期间显示由应用决定。`SetDisabled(true)` 禁止消息内的操作、反应和重试。

```go
msg.Reactions(kit.MessageReaction{Name: "有帮助", Count: 2}).
    OnReaction(func(index int, active bool) { /* 保存到服务端 */ })
```

反应使用可切换按钮，点击更新当前用户的选中状态与计数，再调用回调。`Reactions` 复制数据；后续服务端结果可再次调用它覆盖。未提供回调的反应只读。Agent 可读取 `article` 的 `sending` / `failed` 状态，以及反应按钮的选中状态。

`Avatar(view)` 替换头像，传 nil 隐藏；`DefaultAvatar()` 恢复默认：助手显示姓名首字母，用户消息不显示头像。用户消息显式配置头像时放在右侧。头像目前按消息顶部对齐，与 GPUI 的正文底边对齐不同。

`Header(view)` 放在正文上方，`Footer(view)` 放在状态、操作和反应区之后；传 nil 清除。头尾支持任意 View 和可交互控件，默认继承小号、弱化文字样式，用户消息靠右。`Content(view)` 独立替换正文，nil 清空正文但保留其他分区。消息禁用状态覆盖所有插槽。

```go
msg.Avatar(kit.Avatar("Alice").Size(32)).
    Header(kit.Label("Alice · 10:24")).
    Footer(kit.Button("回复", reply))
```

插入、删除头像或头尾时，正文的身份保持稳定，输入内容和焦点不变。普通用户气泡和显式 Bubble 的头尾默认使用 `theme.SpaceLg` 水平缩进；普通助手正文保持无缩进。完整消息行可用 [MessageGroup](message_group.md) 分组。

`Bubble(surface)` 安装显式气泡，避免 User 消息再包一层气泡。组件以副本渲染，气泡对齐跟随消息；修改原气泡的变体会在下一帧体现，不会反向修改原实例。传 nil 清空正文；`Content(view)` 恢复普通正文模式（User 自动包气泡）。

```go
surface := kit.Bubble(answer).Variant(kit.BubbleGhost)
msg.Bubble(surface).Header(kit.Label("系统消息")).Footer(kit.Label("刚刚"))
msg.HeaderInset(true).FooterInset(false)
msg.ResetContentInsets() // 恢复自动规则
```

显式 Ghost 气泡自动取消头尾缩进；`HeaderInset`、`FooterInset` 分别覆盖继承规则。普通 Content 内部的自定义 View 不参与 Ghost 检测。入口接受单个 Bubble，多段内容可在该气泡内部组合；尚无上游多个 typed bubble 混排并合并元数据的 MessageContent 部件。
