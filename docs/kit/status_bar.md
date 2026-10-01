# StatusBar

并排显示主状态和详情，由父布局决定放置位置，不自动固定到窗口底部。

```go
bar := kit.StatusBar("连接正常", "共 123 条记录")
return bar.Render(cx)
```

构造函数返回 `*StatusBarView`。`SetStatus`、`SetDetail` 更新文字；两栏平均分配可用宽度，长文案在各自栏内换行。颜色跟随主题，无操作按钮、键盘行为或自动播报。

Agent 角色 status，名字为主状态，value 为详情。验证：`go run ./examples/components -section status-bar -theme dark`，支持 light。
