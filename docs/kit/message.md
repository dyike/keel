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
