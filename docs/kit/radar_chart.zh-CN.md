# RadarChart

[English](radar_chart.md) | 简体中文

`RadarChart(labels, series...)` 返回 `ChartView`，共用 Title、Height、Format、SeriesStyle、TooltipContent、图例与数据表切换。至少三个维度才绘制雷达；有限非负值参与绘图，缺失、负值和非有限值形成缺口，不用零值补齐。

```go
chart := kit.RadarChart([]string{"速度", "质量", "成本"},
    kit.Series{Name: "方案 A", Values: []float64{80, 95, 60}},
    kit.Series{Name: "方案 B", Values: []float64{90, 75, 80}},
).RadarMax(100).GridLevels(5).Title("方案比较")
```

- RadarMax(0) 按可见系列自动取最大正值；固定最大值之外的点压到外环，提示和表格仍报告原数值。
- GridLevels 为 1–20，默认 4；OuterRadius 为 dp，0 自动适配，显式半径也受可用区域约束。
- SeriesStyle 配置各系列描边、填充、线宽和顶点圆点；悬停按最近辐条显示各可见系列数据。
- RadarLabel 可返回自定义展示元素，标签槽宽 72dp；它仍使用原标签作为提示标题。标签排版和默认圆点运动不同于 GPUI，未实现悬停过渡动画。
- SetData 深复制数据并恢复系列可见性；Data 返回副本。SetDisabled 禁用图例、悬停和表格按钮。

示例：`go run ./examples/components -section radar_chart`。浅深色像素与 Agent 检查见测试，原生窗口视觉另行验收。
