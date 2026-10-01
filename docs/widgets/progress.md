# 进度条

`SetValue` 把值限制在 [0,1]，NaN 按 0 处理。`SetIndeterminate(true)` 显示动画并隐藏百分比；切回时保留之前的 `Value()`。按帧请求重绘，关闭动画后停止。自动化快照的 `value` 为百分比或 `indeterminate`。

验证入口：`go run ./examples/components -section progress`。
