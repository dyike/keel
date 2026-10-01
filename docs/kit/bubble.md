# Bubble

聊天气泡。

```go
kit.Bubble(text).Mine()  // 自己发的：靠右、主色
kit.Bubble(text)         // 别人发的：靠左、浅色底
```

- 宽度最多占父容器的 75%。内容是任意 View。
- 只要气泡本身时用 Bubble；要带头像和操作栏的完整一条消息，用 Message。

验证：`go run ./examples/components -section bubble`，加 `-theme dark` 检查深色。
