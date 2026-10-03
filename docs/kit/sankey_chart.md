# SankeyChart

桑基图展示有向无环流量。节点按输入顺序稳定排列，边使用节点索引，节点吞吐量为入流、出流之和的较大值。流带在起点与终点颜色间渐变，悬停节点时淡化无关流带。

```go
chart := kit.SankeyChart(
    []kit.SankeyNode{{Name:"营收"}, {Name:"成本"}, {Name:"利润"}},
    []kit.SankeyLink{{Source:0, Target:1, Value:55}, {Source:0, Target:2, Value:45}},
).Title("收入分配").Height(320).NodePadding(20).NodeRadius(2)
if err := chart.Error(); err != nil { /* 首次数据无效 */ }
if err := chart.SetData(nodes, links); err != nil { /* 保留原图 */ }
```

构造失败时 Error 返回原因，并渲染错误内容。SetData 检查索引、自环、循环、负数、NaN/Infinity 及总量溢出；失败返回错误并保留旧数据，成功后清除构造错误。数据和可选颜色均复制，回调收到的颜色也为副本。零流量边保留在数据表及拓扑中，但不画流带。

配置：

- NodeAlign：SankeyAlignJustify 默认把汇点放在最右列；Left 使用拓扑深度，Right 按到终点的最长距离，Center 将根节点靠近其下一层。
- NodeWidth 默认 10dp，NodePadding 默认 16dp，NodeRadius 默认 0；窄/矮空间会压缩间隔以保持节点在图内。
- ValueScale 默认为 SankeyValueScaleLinear；Sqrt 压缩节点高度差。两端流带宽度分别按原始流量占各节点吞吐量的比例分配，不改变提示数值。零吞吐节点最多保留 2dp 高度。
- Iterations 为 0–64，默认 6；使用邻居加权位置松弛和稳定输入顺序消除节点重叠，不实现 d3 的全部重排序策略。
- SankeyNode.Color 覆盖节点颜色；LinkOpacity 为 0–1，默认 0.3。尚无流带最小宽度配置，极小正流量可能小于一个像素。
- Format 控制数值；Labels、TooltipContent 提供自定义展示元素，接收节点索引、名称和原始吞吐量。默认标签在节点旁至多 100dp，长文字截断；密集节点的文字仍可能互相遮挡。
- 图表角色为 figure，每个节点报告名称及吞吐量；数据表提供 Source、Target、Value。SetDisabled 禁用悬停和切换。没有节点拖动和悬停过渡动画。

示例：`go run ./examples/components -section sankey_chart`。真机视觉另行验收。
