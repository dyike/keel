# Marker

用 8dp 圆点和文字说明状态，避免只靠颜色传达含义。

```go
state := kit.Marker("在线").Tone(kit.Success)
return state.Render(cx)
```

`Marker(label)` 返回 `*MarkerView`；`SetLabel` 更新文字，链式 `Tone` 选择语义色，默认 Info。圆点尺寸固定，切换颜色不改变布局；长文字在可用宽度中换行。颜色每帧跟随主题，当前没有闪烁、点击或键盘行为。

Agent 角色 marker，名字是文字，value 为语义级别。验证：`go run ./examples/components -section marker -theme dark`，支持 light。
