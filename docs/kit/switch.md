# Switch

立即生效的开关，比如"接收通知"。

```go
notify := kit.Switch("接收通知", true).OnChange(func(on bool) { save(on) })
```

- 点击、Space、Enter 切换；`Value()` / `SetValue(bool)`，`SetValue` 不触发回调；`SetDisabled`。

Agent：角色 `switch`，`checked` 表示开关状态。

验证：`go run ./examples/components -section switch`，加 `-theme dark` 检查深色。
