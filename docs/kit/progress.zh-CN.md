# Progress

[English](progress.md) | 简体中文

带标签的进度条。

```go
p := kit.Progress("导入订单")
p.SetValue(0.42)          // 显示 42%
p.SetIndeterminate(true)  // 不知道总量时显示来回滑动的进度块
```

- `SetValue` 会把值限制在 0–1（NaN 归零），并退出不确定模式。开启减少动画时，不确定进度静止显示。

Agent：角色 `progressbar`，`value` 是百分比或 `indeterminate`。

验证：`go run ./examples/components -section progress`，加 `-theme dark` 检查深色。

圆形进度及中心内容见 [ProgressCircle](progress_circle.zh-CN.md)。

`Height(dp)` 设置条形高度，范围 1–128dp，默认 8；`Color(c)` 覆盖填充色并关闭该实例的主题渐变；`Rounded(dp)` 设置轨道和进度块圆角，0 为直角。非法数值忽略。`TrackStyle(func(*el.DivEl))` 在每帧默认样式之后设置轨道背景、边框等；nil 恢复默认，不要保留元素引用。标签和百分比不受轨道样式影响；进度块填满轨道的内部高度。配置同时适用于确定和不确定模式。

确定进度更新使用 200ms 平滑过渡（`kit.ProgressDuration`），连续更新从当前显示位置衔接。首次显示、从不确定模式返回和减少动画时立即显示目标值。`Value()`、百分比及 Agent 语义始终返回目标值，只有图形插值。动画使用帧时钟，不创建定时器或 goroutine。
