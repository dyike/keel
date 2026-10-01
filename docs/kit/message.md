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
