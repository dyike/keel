# ProgressCircle

[English](progress_circle.md) | 简体中文

圆形进度显示，支持真实进度、不确定动画和中心内容。

```go
p := kit.ProgressCircle("导入订单").Size(64)
p.SetValue(.42)
p.Child(el.ViewFunc(func(*el.Context) el.Element {
    return el.Text(fmt.Sprintf("%.0f%%", p.Value()*100))
}))
```

`Value/SetValue` 读写 0–1 的进度；赋值退出不确定模式。NaN 归零，无穷和越界值限制到端点。`SetIndeterminate(true)` 显示旋转弧，减少动画时静止；切回确定模式保留之前的值。

`Size` 设置直径（默认 48dp，忽略非正数和非有限数）；`Color` 显式覆盖前景色，默认随主题切换。`SetLabel` 更新语义名称，`Child` 设置中心视图，传 nil 清空。中心内容应简短，绘制裁剪在组件范围内。窄约束下圆环取可用宽高的较小值居中绘制。

纯展示组件，不接收键盘焦点。Agent 角色为 `progressbar`，名称取 label，值为百分比或 `indeterminate`。应用负责更新状态；后台更新应通过 `core.Update`。

验证：`go run ./examples/components -section progress_circle`，以 `-theme dark` 检查深色，`-width 320 -scale 2 -screenshot /tmp/progress-circle.png` 检查窄布局。单元测试覆盖状态和约束，窗口测试覆盖 Agent 值与动画像素。

确定进度更新使用 200ms 平滑过渡（`kit.ProgressDuration`），连续更新从当前显示位置衔接。首次显示、从不确定模式返回和减少动画时立即显示目标值。`Value()`、百分比及 Agent 语义始终返回目标值，只有图形插值；中心自定义内容由调用方管理。动画使用帧时钟，不创建定时器或 goroutine。
