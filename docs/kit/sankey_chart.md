# SankeyChart

English | [简体中文](sankey_chart.zh-CN.md)

Sankey diagram shows directed acyclic flow. Nodes are arranged stably in input order, edges use node indexes, and node throughput is the larger of the sum of incoming and outgoing flows. The flow band gradients between the start and end colors, and irrelevant flow bands fade when the node is hovered.

```go
chart := kit.SankeyChart(
    []kit.SankeyNode{{Name:"Revenue"}, {Name:"Cost"}, {Name:"Profit"}},
    []kit.SankeyLink{{Source:0, Target:1, Value:55}, {Source:0, Target:2, Value:45}},
).Title("Revenue allocation").Height(320).NodePadding(20).NodeRadius(2)
if err := chart.Error(); err != nil { /* Initial data is invalid */ }
if err := chart.SetData(nodes, links); err != nil { /* Keep the previous chart */ }
```

When construction fails, Error returns the reason and renders the error content. SetData checks index, self-loop, loop, negative number, NaN/Infinity and total overflow; if it fails, it returns an error and retains the old data. If it succeeds, it clears the construction error. The data and optional colors are copied, as is the color received by the callback. Zero-flow edges are retained in the data table and topology, but flow bands are not drawn.

Configuration:

- NodeAlign: SankeyAlignJustify defaults to placing the sink point in the rightmost column; Left uses the topological depth, Right presses the longest distance to the end point, and Center places the root node close to its lower level.
- NodeWidth defaults to 10dp, NodePadding defaults to 16dp, and NodeRadius defaults to 0; narrow/short spaces will compress the intervals to keep nodes within the graph.
- ValueScale defaults to SankeyValueScaleLinear; Sqrt compresses node height differences. The width of the flow belts at both ends is allocated according to the proportion of the original traffic to the throughput of each node, and the prompt value does not change. Zero-throughput nodes retain a maximum height of 2dp.
- Iterations is 0–64, default is 6; uses neighbor-weighted positional relaxation and stable input order to eliminate node overlap, and does not implement the full reordering strategy of d3.
- SankeyNode.Color overrides node color; LinkOpacity is 0–1, default 0.3. MinLinkWidth can set the minimum visual width of positive flow. The default value is 0 to maintain the proportion; if not set, the minimum positive flow may be less than one pixel.
- Format controls the value; Labels and TooltipContent provide custom display elements and receive node index, name and raw throughput. Default labels are at most 100dp next to nodes, with long text truncated. Conflicts are detected based on the actual layout boundaries, and hovering nodes are retained first, in descending order of throughput; labels that cannot fit or are less than 2dp away from reserved labels are omitted. Custom labels also participate in detection, and the text does not move next to other nodes. Each node still provides name and throughput semantics, and hover and data tables retain complete data.
- The chart role is figure, and each node reports the name and throughput; the data table provides Source, Target, and Value. SetDisabled disables hover and toggle. No node dragging; support for hover transitions.

Example: `go run ./examples/components -section sankey_chart`. Real machine vision will be subject to separate acceptance.

`MinLinkWidth(dp)` accepts limited values from 0–64dp, illegal values are ignored. When enabled, the height is allocated according to the sum of the widths of each node's in/out ports, and then the proportional scale is compressed so that each column fits into the frame; the original value, sorting, and prompts remain unchanged. If the entire column cannot accommodate the minimum width of all ports, the minimum width is reduced uniformly. Zero-value streambands are not drawn; there is no guarantee that the requested width or one physical pixel will still be reached in limited space. It is suitable for both Linear/Sqrt. The thickening of small flows will change the visual proportion, which is suitable for displays that need to highlight small branches.

Example enabling 3dp lower limit. Testing covers 40 sets of random DAGs, 20/200dp height, Linear/Sqrt, traffic across five orders of magnitude and 1000 isolated nodes, checking node and port boundaries and limited coordinates. Native vision is still pending acceptance.

`HoverAnimation(false)` turns off hover transition, enabled by default. The emphasis state changes with three 150ms slow-outs, and the rapid target change continues from the current weight; the target is displayed immediately when the animation is reduced. Pie chart transition selected border, Sankey chart transition irrelevant stream band transparency, line/area chart transition focus line and point, bar/candle chart transition category background, radar chart transition focus. Hit and cue data are updated immediately, no waiting for animations. Update data to clean up old transitions. The shared animation test covers intermediate frame weights, continuity when switching targets, end frames and reduced motion; the look and feel of the native animation of each picture is still to be accepted.

Dense label test covers 30 nodes, 180/600dp width, 1×/2×, custom 36dp high label and hover priority; light-dark window rendering with Agent hit regression passes. The native window check is not performed yet due to the current Mac lock screen.
