# RadarChart

English | [简体中文](radar_chart.zh-CN.md)

`RadarChart(labels, series...)` returns `ChartView`, sharing Title, Height, Format, SeriesStyle, TooltipContent, legend and data table switching. The radar must be drawn in at least three dimensions; finite non-negative values participate in the drawing. Missing, negative and non-finite values form gaps and do not need to be filled with zero values.

```go
chart := kit.RadarChart([]string{"速度", "质量", "成本"},
    kit.Series{Name: "方案 A", Values: []float64{80, 95, 60}},
    kit.Series{Name: "方案 B", Values: []float64{90, 75, 80}},
).RadarMax(100).GridLevels(5).Title("方案比较")
```

- RadarMax(0) automatically takes the maximum positive value according to the visible series; points outside the fixed maximum value are pressed into the outer ring, and the prompts and tables still report the original value.
- GridLevels is 1–20, default is 4; OuterRadius is dp, 0 automatically adapts, and the explicit radius is also constrained by the available area.
- SeriesStyle configures the stroke, fill, line width, and vertex dots of each series; hover to display each visible series data by the nearest spoke.
- RadarLabel returns a custom display element with a label slot width of 72dp; it still uses the original label as the prompt title. Label layout and default dot motion are different from GPUI, and hover transition animation is not implemented.
- SetData deep copies the data and restores series visibility; Data returns the copy. SetDisabled disables legend, hover, and table buttons.

Example: `go run ./examples/components -section radar_chart`. For light and dark pixels and Agent inspection, please refer to the test, and the native window vision shall be subject to separate acceptance.
