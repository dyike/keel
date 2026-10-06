# 公共绘图基础件

[English](plot_primitives.md) | 简体中文

`github.com/dyike/keel/ui/plot` 提供自定义图表的比例尺、数据布局和即时绘制。`kit.Plot` 继续提供带缩放、平移和拾取的成品图；本包由应用掌握布局、主题颜色、输入处理与数据描述。

```go
x := plot.NewBand([]string{"A", "B"}, [2]float64{0, 300}).Padding(.1, .1)
y := plot.ScaleLinear{Domain: [2]float64{0, 100}, Range: [2]float64{200, 0}}
left, ok := x.Map("A")
top, valid := y.Map(60)
if ok && valid {
    plot.Bar{
        Rect: plot.Rectangle{Min: f32.Pt(float32(left), float32(top)),
            Max: f32.Pt(float32(left+x.Bandwidth()), 200)},
        Fill: theme.Chart[0],
    }.Paint(plot.Canvas{Context: gtx, Bounds: image.Rect(20, 10, 320, 210)})
}
```

Canvas 使用本地像素坐标，每次绘制平移到 Bounds.Min，并裁剪到 Bounds。应在 Layout 内创建，不跨帧保存。尺寸需要由调用方按 gtx.Metric 换算；Axis.TextSize 使用 sp。源几何坐标非有限或绝对值超过一千万像素时跳过绘制，避免交给底层图形引擎。

## 比例尺与数据布局

- ScaleLinear：支持反向域/值域、Map、Invert、Clamp；常量域映射到中点。Ticks 返回含端点的均匀刻度，数量限制为 2–1000，不做 nice 刻度取整。非法输入或不能表示的外推结果返回 false。
- NewBand：分类域去重并复制，Map 返回条带低坐标，Bandwidth 返回宽度。Padding 配置内外间距，Align 配置剩余空间分配；修改方法返回新比例尺。反向值域倒置分类顺序，宽度仍为正。
- NewPoint：分类点位置，默认两端对齐，单项居中；Padding 留两端空间。
- NewOrdinal：分类映射到循环使用的离散值域；复制输入切片，缺失分类/空值域返回 false，不自动扩域。值域元素中的引用由调用方管理。
- Stack：输入按系列组织的二维值，输出保留系列/样本索引及 Low/High/Valid。正负值分别从零累计，NaN/缺失样本保留无效槽；无穷或累计溢出整次返回错误。
- Pie：输出保留原始索引、值及起止弧度；允许正反方向最多一圈，零值省略，负数/非有限值报错。按最大值归一化避免大值求和溢出；间隙角限制在每项可用角度内。零角宽条目可保留索引但不绘制。

## 绘制

Bar 支持任意方向矩形和圆角；Line 支持直线、阶梯、逐段 smoothstep 曲线及数据点；Area 填充等长上下边界之间的区域，可描上边线。缺失或非法点打断线/面积，不连接缺口。Arc 绘制扇形/环形，使用有限分段近似圆弧；Dot 和 CrossLine 提供点及十字参考线。

Axis 支持上下左右四方向、自定义刻度标签、线宽、颜色和字号；At 为轴位置，From/To 为基线两端。标签同样受 Canvas 裁剪，调用方需预留边距。低层图形及轴标签不会自动生成 Agent 数据节点，应在周围提供可读标题、数据表或文本；提示框可用 `kit.Tooltip` 或应用浮层组合。此包不自动处理悬停动画，也不等同于上游全部样式配置。

示例将堆叠柱、趋势线、参考线、轴与环图组合：`go run ./examples/components -section plot_primitives`。
