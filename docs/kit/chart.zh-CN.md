# Chart

[English](chart.md) | 简体中文

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

另见 [RadarChart](radar_chart.zh-CN.md) 和 [SankeyChart](sankey_chart.zh-CN.md)。新增配置的自动验证范围见下文。新增接口不代表视觉或 API 完全一致。

## 轴标签布局

`Gutter(ChartGutter{Left, Right, Top, Bottom})` 以 dp 指定笛卡尔绘图区四边的预留尺寸，`AutoGutter()` 恢复默认 Left=52、Bottom=18、其余为零。Height 仍表示绘图区高度；外部标题、控件及分区间距另计。四个值均须在 0–4096 范围内且有限，否则整组配置不变。Bottom 为零隐藏类别标签，自定义的较小 Bottom 会裁剪标签区。

`YLabelsInside(true)` 将纵轴标签置于绘图区左侧，未显式 Gutter 时自动去掉外侧左标签列；显式 Gutter 保持调用方给定值。标签可能覆盖数据，参考线文字和提示在标签上方绘制。横轴标签跟随实际绘图区宽度与左右边距对齐。以上接口也用于 CandlestickChart，不影响 RadarChart/PieChart。

组件库订单图演示轴内标签及四边预留。1×/2× 布局测试覆盖预留尺寸、非法配置原子拒绝、隐藏底部标签和恢复默认布局；原生窄窗口的轴文本仍需视觉验收。

`FutureSlots(n)` 在已有类别后预留 n 个空位置（0–100000，默认 0），适用于折线、面积、柱图及 CandlestickChart。数据、横轴标签、悬停带与提示使用同一类别宽度；鼠标进入空位不显示提示。空位不产生数据表行、不参与纵轴域，也不会自动生成日期。设回 0 恢复铺满绘图区；改变配置清除旧悬停。雷达图不使用此配置。示例预留两期。测试覆盖四向柱图与蜡烛的真实数据命中、空位不命中、恢复默认以及纵轴域不变；密集蜡烛的原生视觉仍待验收。

`BarFill(func(ChartBarDatum) ChartBarFill)` 逐柱或逐堆叠段配置填充。参数含原始 Series/Index、系列名、类别标签、数值、堆叠标志和默认系列色；返回 Color 或非 nil 的 Gradient。`ChartBarGradient{Start, End, Direction}` 沿当前柱段矩形向 Top/Bottom/Left/Right 渐变，默认及非法方向使用 Bottom，保留透明度与圆角。传 nil 恢复系列色。此接口只影响柱体，不改变图例、提示或数据。回调可能在测量和绘制时执行，必须无副作用。

订单图示例加入逐柱渐变。1×/2× GPU 像素测试覆盖四种柱体方向与四种渐变方向的组合，另验证堆叠回调保留正负原值。渐变方向使用屏幕方向；柱体基线方向通过 BarAlignment 配置。

`HoverAnimation(false)` 关闭悬停过渡，默认开启。强调状态以 150ms 三次缓出变化，快速换目标从当前权重继续；减少动画时立即显示目标。饼图过渡选中边框，桑基图过渡不相关流带的透明度，折线/面积图过渡焦点线和点，柱/蜡烛图过渡类别背景，雷达图过渡焦点。命中和提示数据立即更新，不等待动画。更新数据清理旧过渡。共享动画测试覆盖中间帧权重、切换目标时的连续性、结束帧及减少动画；各图的原生动画观感仍待验收。

## 柱体方向

`BarAlignment(BarAlignmentBottom/Top/Left/Right)` 指定柱体基线所在侧：默认 Bottom 向上，Top 向下，Left 向右，Right 向左。保留正负值的零基线、分组/堆叠、缺失数据、FutureSlots、圆角和逐柱填充。文字保持正向，横向时类别沿上到下排列，数值轴沿左右方向；悬停命中在相同坐标变换下处理。切换方向清除旧悬停。非柱图忽略该配置。

横向的默认预留为 Left=80、Bottom=24，显式 Gutter 优先。YLabelsInside 将数值刻度移进绘图区，类别标签仍在左侧。参考线、提示与数值刻度使用对应方向的坐标，渐变的 Top/Bottom/Left/Right 保持屏幕方向。

季度收入示例采用左基线横向堆叠图。四向命中与渐变像素已通过自动测试；参考线和文字的完整原生布局仍待验收。固定总点位与多停靠点渐变见下文。

## 固定点位与数值渐变

`PointCount(n)` 固定分类轴至少容纳 n 个位置（0–100000），追加数据时已有类别位置保持不变；实际数据超过 n 后扩展轴，0 恢复自动。它与 FutureSlots 相互替换，适用于折线、面积、柱图和蜡烛图。XTickCount 在固定总位置上选择刻度，仅展示已有类别的文字，未来空位不伪造标签、数值或提示。与 [GPUI 的 point_count](https://gpui-kit.com/component/chart/#pinned-axis-and-unfinished-series) 用途相同，Keel 继续使用分类带中心坐标。

`BarGradient(func(ChartBarDatum, ChartBarRange) []ChartColorStop)` 提供沿柱体基线到顶端的多停靠点渐变。`ChartBarRange` 含显示轴域 Min/Max 和当前段累计值 Base/Tip；`ChartToBar(value)` 将数值映射到当前段，Base 为 0、Tip 为 1，允许越界。正负柱及四种方向均按真实生长方向绘制。堆叠回调的 datum.Value 保留原值。

每个停靠点含 Position 和 Color；实现复制并排序输入，重复位置采用最后一个颜色。区间外的点在边界按线性光、预乘 alpha 插值；空数组或非有限位置整组回退系列色。相邻色段按物理像素边界切分，不保留亚像素宽的色段。BarGradient 与 BarFill 相互替换，传 nil 恢复系列色；回调应无副作用。

```go
chart.BarGradient(func(d kit.ChartBarDatum, r kit.ChartBarRange) []kit.ChartColorStop {
    return []kit.ChartColorStop{
        {Position: r.ChartToBar(r.Min), Color: theme.Chart[0]},
        {Position: r.ChartToBar((r.Min+r.Max)/2), Color: theme.Chart[1]},
        {Position: r.ChartToBar(r.Max), Color: theme.Chart[2]},
    }
})
```

自动验证覆盖追加时固定刻度、两种点位配置切换、超额数据、区间外插值、输入副本、非法停靠点、极值映射、正负堆叠端点；GPU 像素测试覆盖双倍率、四向正负柱、半透明整图映射在不同柱高上的颜色一致性。原生窗口仍待验收。
