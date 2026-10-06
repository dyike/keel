# PieChart

English | [简体中文](pie_chart.zh-CN.md)

Pie charts and donut charts show proportions of the same set of data.

```go
chart := kit.PieChart(
    kit.PieSlice{Name: "线上", Value: 60},
    kit.PieSlice{Name: "门店", Value: 40},
).Title("收入来源").Donut(.6)
```

`Height(dp)` sets the height of the drawing area; `Donut(fraction)` sets the central hole radius ratio, range 0–0.9, default 0. `Format(fn)` sets the numerical format, keeping the ratio to one decimal place. `SetData` Copies the data and restores the legend to visibility.

The legend can be clicked or toggled with Space/Enter to recalculate the scale for the remaining visible data. Hovered sectors show name, original value, and proportion; center hole misses. Colors are assigned by original index, hiding does not change the remaining item colors. `SetDisabled` disables legend, table switching and hover.

Only finite positive numbers participate in sectors; 0 has no sector, and negative numbers, NaN, and Inf are shown in the table as `—`. The empty state is displayed when all are empty, all are hidden, or all are invalid. The proportion is first scaled by the maximum value, and the two largest floating point numbers will not cause the total to overflow. The data table retains all items, and the proportion of hidden items is 0.

Agent: Outer `figure`, the name is the title; the legend is `toggle` with the selected state. Switch to the data table to read the name, value and proportion.

Verify: `go run ./examples/components -section pie_chart`, add `-theme dark` to check the dark theme. Testing covers sector and center hole hits, legend recalculation, disabling, data copying, null data, outliers, and data tables.

`TooltipContent(func(cx, PieChartTooltip) el.Element)` adds a custom overlay that follows the mouse, and the original bottom reading and data table are retained. The callback parameters include the original Index/Slice, the normalized Share (0–1) of the visible sector, the formatted value/scale and the current theme sector color. Parameters are value snapshots; do not modify data in render callbacks.

Passing nil restores only the bottom reading, and the callback returns nil to hide the overlay by sector. The prompt only displays the content and does not carry interaction; the corresponding prompt will no longer be displayed after leaving the pie chart, entering the donut hole, hiding the sector, updating data, or disabling it. The overlay anchors the mouse position and is constrained by the window size. Add custom prompts to the component library. 1×/2× test covers the proportion after mouse hits, overlay horizontal boundaries, explicit disabling and legend switching; the complete native window interaction is still to be accepted.

`HoverAnimation(false)` turns off hover transition, enabled by default. The emphasis state changes with three 150ms slow-outs, and the rapid target change continues from the current weight; the target is displayed immediately when the animation is reduced. Pie chart transition selected border, Sankey chart transition irrelevant stream band transparency, line/area chart transition focus line and point, bar/candle chart transition category background, radar chart transition focus. Hit and cue data are updated immediately, no waiting for animations. Update data to clean up old transitions. The shared animation test covers intermediate frame weights, continuity when switching targets, end frames and reduced motion; the look and feel of the native animation of each picture is still to be accepted.
