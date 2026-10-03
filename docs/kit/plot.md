# Plot

数值型的 x / y 绘图，可以交互探索数据。

```go
p := kit.Plot(
    kit.PlotSeries{Name: "sin(x)", Points: wave},
    kit.PlotSeries{Name: "实测", Points: samples},
).Lines().Title("信号")
```

- 鼠标：滚轮以指针位置为中心缩放；拖动平移；双击或点"复位"恢复显示全部数据。
- 键盘：获得焦点后，+ / − 缩放，方向键平移，0 复位。
- 悬停时会选中 12dp 范围内最近的数据点，放大显示，并弹出提示框显示名称和坐标。
- 默认画成散点；`Lines()` 用 2dp 的线把各点连起来，只放大显示被选中的那个点。点的颜色取 `theme.Chart`，外面有一圈背景色的描边，重叠时也能分清。
- `View()` / `SetView()` 读取或设置当前可见范围，`Reset()` 适配全部数据，`SetSeries` 替换数据，`Format(fn)` 设置数值写法。

Agent：角色 `figure`；"复位"是一个按钮。

验证：`go run ./examples/components -section plot`，加 `-theme dark` 检查深色。

构造和 `SetSeries` 复制系列及点切片；包含 NaN/Inf 坐标的点不参与适配、拾取或绘制，折线在该处断开。空数据和全无效数据显示空状态。`SetView` 忽略非有限或倒置范围。

缩放保持端点有限且有序，达到相对精度下限（约当前坐标尺度的 10⁻¹²）后停止继续放大；平移或缩小若会溢出则保留原范围。滚轮单次倍率限制为 0.01–100。极大或极小数值默认使用科学记数法。

`SetDisabled(true)` 禁用指针、复位与键盘；拖动取消或拖动中禁用所属区域会回到拖动开始的视图。更换数据会清除旧的悬停选点。标题和图例在窄窗口换行，提示宽度不超过绘图区。

自定义图表可使用独立 `ui/plot` 包，见[公共绘图基础件](plot_primitives.md)。
