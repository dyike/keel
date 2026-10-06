# SankeyChart

[English](sankey_chart.md) | 简体中文

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
- SankeyNode.Color 覆盖节点颜色；LinkOpacity 为 0–1，默认 0.3。MinLinkWidth 可设置正流量的最小视觉宽度，默认 0 保持比例；未设置时极小正流量可能小于一个像素。
- Format 控制数值；Labels、TooltipContent 提供自定义展示元素，接收节点索引、名称和原始吞吐量。默认标签在节点旁至多 100dp，长文字截断。按实际排版边界检测冲突，优先保留悬停节点、再按吞吐量降序；放不下或与已保留标签相距不足 2dp 的标签省略。自定义标签也参与检测，文字不移动到其他节点旁。每个节点仍提供名称与吞吐量语义，悬停及数据表保留完整数据。
- 图表角色为 figure，每个节点报告名称及吞吐量；数据表提供 Source、Target、Value。SetDisabled 禁用悬停和切换。没有节点拖动；支持悬停过渡。

示例：`go run ./examples/components -section sankey_chart`。真机视觉另行验收。

`MinLinkWidth(dp)` 接受 0–64dp 的有限值，非法值忽略。启用后按每个节点入/出端口的宽度总和分配高度，再压缩比例尺度使各列放入图框；原始值、排序和提示不变。若整列无法容纳所有端口的最小宽度，则统一降低最小宽度。零值流带不绘制；有限空间下不保证仍达到请求的宽度或一个物理像素。Linear/Sqrt 均适用，小流量变粗会改变视觉比例，适合需要突出小分支的展示。

示例启用 3dp 下限。测试覆盖 40 组随机 DAG、20/200dp 高度、Linear/Sqrt、跨五个数量级流量和 1000 个孤立节点，检查节点与端口边界及有限坐标。原生视觉仍待验收。

`HoverAnimation(false)` 关闭悬停过渡，默认开启。强调状态以 150ms 三次缓出变化，快速换目标从当前权重继续；减少动画时立即显示目标。饼图过渡选中边框，桑基图过渡不相关流带的透明度，折线/面积图过渡焦点线和点，柱/蜡烛图过渡类别背景，雷达图过渡焦点。命中和提示数据立即更新，不等待动画。更新数据清理旧过渡。共享动画测试覆盖中间帧权重、切换目标时的连续性、结束帧及减少动画；各图的原生动画观感仍待验收。

密集标签测试覆盖 30 个节点、180/600dp 宽度、1×/2×、自定义 36dp 高标签和悬停优先；浅深色窗口渲染与 Agent 命中回归通过。原生窗口检查因当前 Mac 锁屏暂未执行。
