# Chart

分类数据的折线图和柱状图，只有一条纵轴。

```go
sales := kit.LineChart(months,
    kit.Series{Name: "华东", Values: east},
    kit.Series{Name: "华北", Values: north},
).Title("月度销售额（万元）")

mix := kit.BarChart([]string{"Q1", "Q2", "Q3", "Q4"}, online, stores).Stacked().Height(180)
```

- 系列颜色依次取 `theme.Chart` 的 8 个分类色，第 i 个系列始终是第 i 个颜色，增减系列时其他系列的颜色不变。这套颜色在浅色和深色背景上都用色觉缺陷校验工具验证过。
- 有两个及以上系列时显示图例；只有一个系列时，标题已经说明画的是什么，不显示图例。
- 悬停时，柱状图会高亮所在的那一组，折线图显示十字线和数据点；同时弹出提示框，列出该位置所有系列的值。
- "查看数据表"按钮把同样的数据切换成 Table 显示。有几个分类色对背景的对比度低于 3:1，表格视图是给它们的补救，也方便读屏和 Agent 读取数据。
- 柱子最宽 24dp，顶端 4dp 圆角，从同一条基线长出；相邻柱子之间留 2dp 间隙。堆叠时各段之间也留 2dp 间隙，只有最外侧一段是圆角。线宽 2dp。网格线是 1px 的浅色实线。
- 纵轴刻度取整（1、2、5 × 10ⁿ），数字带千分位。横轴标签放不下时会自动隔几个显示，保证不重叠。
- `Format(fn)` 统一设置轴刻度、提示框和表格中数值的写法；`SetData` 替换数据。
- 不支持双纵轴：两种量纲不同的数据请画成两张图。

Agent：角色 `figure`，名字是标题（没有标题时是各系列名），`value` 为"项数x系列数"；切换到表格后，各行以 `row` 列出。

验证：`go run ./examples/components -section chart`，加 `-theme dark` 检查深色。

图例是可聚焦的开关，点击或按空格/Enter 隐藏与恢复系列；隐藏后重新计算纵轴，颜色仍按原系列位置。`SetDisabled(true)` 禁用图例、视图切换和悬停。

构造和 `SetData` 都复制分类、系列及数值切片。再次 `SetData` 会恢复全部系列并重建数据表列名。缺少的值、NaN、Inf 显示为 `—`，不参与坐标范围，也不会连接缺口；真实的 0 仍正常绘制。正负堆叠分别累计，极端数据避免浮点溢出。

密集折线按像素分桶，保留各桶的首尾、最大和最小值。10 万点仍保留尖峰，连续区间的绘图点数约为视宽的四倍；分类标签只生成当前宽度能容纳的数量。数据表和悬停保留原始数据。

`AreaChart(labels, series...)` 在折线与零基线之间填半透明颜色，支持负值、缺口、图例、悬停和数据表；多系列面积重叠显示。`Stacked()` 只适用于柱状图。

## 轴、线型与提示配置

`YDomain(lo,hi)` 固定精确轴域，要求有限且 lo < hi；AutoDomain 恢复自动范围。YTickCount(2–50) 在两端之间均匀布点，0 恢复自动美化刻度；XTickCount(1–100) 在首末类别间分配标签，0 按宽度自动抽稀。GridColumns 添加内部竖线，GridDashed 控制网格虚线。ReferenceLines 接收数值、颜色、可选文字，超出轴域的参考线隐藏。柱/面积图在固定轴域下仍以零为基线，绘图裁剪到图框。

Curve 支持 ChartCurveLinear（默认）、ChartCurveStepAfter 和 ChartCurveSmooth。Smooth 使用逐段单调 smoothstep 采样，不跨越缺失值，也不超出相邻端点的纵向范围；不等同于 GPUI 的默认插值。SeriesStyle(index, style) 可独立设描边、填充、线宽与数据顶点，颜色指针会复制；零线宽使用 2dp。改变数据不清除样式索引。

TooltipContent 接收当前类别索引、标签和可见系列的数值/默认文字/颜色，可返回展示元素；nil 恢复默认。Format 仍控制默认轴/提示/数据表。CandlestickChart 同样提供轴、网格、参考线与提示配置，提示系列依次为开、高、低、收。

另见 [RadarChart](radar_chart.md) 和 [SankeyChart](sankey_chart.md)。仍未实现上游所有选项：如轴内标签/可配置 gutter、预留未来点位、柱图四向对齐/逐柱渐变、最小流带宽度及各图悬停过渡。新增接口不代表视觉或 API 完全一致。
