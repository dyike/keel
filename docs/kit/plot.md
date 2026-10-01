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
