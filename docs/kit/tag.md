# Tag

显示分类或状态的标签，当前版本不可移除、不可选择，不占键盘焦点。

```go
label := kit.Tag("已完成").Tone(kit.Success)
return label.Render(cx)
```

`Tag(text)` 返回 `*TagView`；`SetText` 更新内容；链式 `Tone` 使用 Neutral（默认）、Info、Success、Warning、Danger。文字在窄容器里换行；主题色在 Render 读取。普通内容长度可以改变标签尺寸，它不是固定占位角标。

Agent 角色 `tag`，名字为文字，value 为语义级别。后续移除操作须等待 el 焦点和禁用接口 review。

验证：`go run ./examples/components -section tag -theme dark`，支持 light；测试覆盖混排、窄宽、缩放和语义更新。
