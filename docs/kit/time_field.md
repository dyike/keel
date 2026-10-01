# TimeField

输入一天中的时间，格式为 HH:MM。

```go
start := kit.TimeField("开始时间")
start.SetValue(9*time.Hour + 30*time.Minute)
```

- 接受 `9:30`、`0930`、`930` 等写法，按回车或离开输入框时规整为 `09:30`；不是合法时间的文字会恢复为原值。
- `Value()` 返回从午夜开始的 `time.Duration`，精确到分钟；`SetValue` 超过 24 小时会取余。
- `SetDisabled`、`SetError`。

Agent：角色 `textbox`，`value` 为 `HH:MM`。

验证：`go run ./examples/components -section time_field`，加 `-theme dark` 检查深色。
