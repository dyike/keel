# Chart

English | [简体中文](chart.zh-CN.md)

Line charts and bar charts for categorical data have only one vertical axis.

```go
sales := kit.LineChart(months,
    kit.Series{Name: "East China", Values: east},
    kit.Series{Name: "North China", Values: north},
).Title("Monthly sales (CNY 10,000)")

mix := kit.BarChart([]string{"Q1", "Q2", "Q3", "Q4"}, online, stores).Stacked().Height(180)
```

- The series colors take the 8 classified colors of `theme.Chart` in sequence. The i-th series is always the i-th color. When adding or subtracting series, the colors of other series remain unchanged. This set of colors has been verified using the Color Vision Deficiency Calibration Tool on both light and dark backgrounds.
- When there are two or more series, the legend is displayed; when there is only one series, the title already explains what is drawn, and the legend is not displayed.
- When hovering, the histogram will highlight the group it is in, and the line chart will display cross lines and data points; at the same time, a prompt box will pop up to list the values of all series at that position.
- The "View Data Table" button switches the same data to a Table display. The contrast ratio of several classification colors to the background is less than 3:1. The table view is a remedy for them, and it is also convenient for screen reading and data reading by the Agent.
- Pillars are 24dp wide at most, 4dp rounded at the top, and grow from the same baseline; leave a 2dp gap between adjacent pillars. There is also a 2dp gap between each segment when stacking, and only the outermost segment has rounded corners. Line width 2dp. Gridlines are 1px light solid lines.
- The vertical axis scale is rounded (1, 2, 5 × 10ⁿ), and the numbers include thousandths. When the horizontal axis labels cannot be placed, they will be automatically displayed at intervals to ensure no overlap.
- `Format(fn)` uniformly sets the axis scale, prompt box and writing method of values in the table; `SetData` replaces the data.
- Dual vertical axes are not supported: please draw two graphs for data with two different dimensions.

Agent: role `figure`, the name is the title (if there is no title, it is the name of each series), `value` is "number of items x number of series"; after switching to the table, each row is listed with `row`.

Verify: `go run ./examples/components -section chart`, add `-theme dark` to check the dark theme.

The legend is a focusable switch. Click or press Space/Enter to hide and restore the series; after hiding, the vertical axis is recalculated and the colors remain in the original series position. `SetDisabled(true)` disables legend, view switching and hover.

Both constructor and `SetData` copy categories, series, and numeric slices. `SetData` again will restore all series and rebuild the data table column names. Missing values, NaN, Inf are displayed as `—`, do not participate in the coordinate range, and do not connect gaps; true 0s are still drawn normally. Positive and negative stacks are accumulated separately, and extreme data avoids floating point overflow.

Dense polylines are divided into buckets by pixels, retaining the first and last, maximum and minimum values of each bucket. 100,000 points still retain spikes, and the number of drawing points in the continuous interval is approximately four times the viewing width; the classification label only generates the number that can be accommodated by the current width. Data tables and hovers retain original data.

`AreaChart(labels, series...)` Fills the space between the polyline and the zero baseline with a translucent color, supports negative values, gaps, legends, hovers and data tables; multiple series areas are displayed overlappingly. `Stacked()` only works with bar charts.

## Axis, linetype and prompt configuration

`YDomain(lo,hi)` Fixed exact axis domain, limited requirements and lo < hi; AutoDomain restores automatic range. YTickCount(2–50) distributes points evenly between both ends, 0 restores automatic beautification scale; XTickCount(1–100) allocates labels between the first and last categories, 0 automatically thins out according to width. GridColumns adds internal vertical lines, and GridDashed controls grid dashed lines. ReferenceLines accepts values, colors, and optional text. Reference lines beyond the axis range are hidden. Column/area charts still have zero as the baseline under fixed axes, and the plot is clipped to the frame.

Curve supports ChartCurveLinear (default), ChartCurveStepAfter, and ChartCurveSmooth. Smooth uses piecewise monotonic smoothstep sampling that does not span missing values or extend beyond the longitudinal extent of adjacent endpoints; not equivalent to the GPUI's default interpolation. SeriesStyle(index, style) can set stroke, fill, line width and data vertices independently, and the color pointer will be copied; zero line width uses 2dp. Changing data does not clear the style index.

TooltipContent receives the value/default text/color of the current category index, label and visible series, and can return the display element; nil returns to default. Format still controls the default axis/prompt/data table. CandlestickChart also provides axis, grid, reference line and prompt configuration. The prompt series are open, high, low and close in order.

See also [RadarChart](radar_chart.md) and [SankeyChart](sankey_chart.md). See below for the newly configured automatic verification scope. The new interface does not mean that the visual or API is completely consistent.

## Axis label layout

`Gutter(ChartGutter{Left, Right, Top, Bottom})` uses dp to specify the reserved dimensions on the four sides of the Cartesian drawing area. `AutoGutter()` restores the default Left=52, Bottom=18, and the rest are zero. Height still represents the height of the drawing area; external titles, controls, and partition spacing are additional. All four values must be within the range of 0–4096 and limited, otherwise the entire configuration will remain unchanged. A Bottom of zero hides category labels, and a custom smaller Bottom will crop the label area.

`YLabelsInside(true)` places the vertical axis label on the left side of the drawing area. When Gutter is not explicitly used, the outer left label column is automatically removed; when Gutter is explicitly used, the value given by the caller is maintained. Labels may cover the data, with guide text and tips drawn above the labels. The horizontal axis labels are aligned with the actual plot area width and left and right margins. The above interface is also used for CandlestickChart and does not affect RadarChart/PieChart.

The component library order diagram demonstrates in-axis labels and four-side reservations. 1×/2× layout test covers reserved size, illegal configuration atom rejection, hiding bottom labels and restoring default layout; axis text of native narrow window still requires visual acceptance.

`FutureSlots(n)` reserves n empty positions (0–100000, default 0) after existing categories, suitable for line, area, column and CandlestickChart. Data, horizontal axis labels, hover bands and prompts use the same category width; no prompt will be displayed when the mouse enters an empty space. Empty positions do not generate data table rows, do not participate in vertical axis fields, and do not automatically generate dates. Setting it back to 0 restores filling the drawing area; changing the configuration clears the old hover. Radar charts do not use this configuration. Samples are reserved for two periods. The test covers the real data hits of four-way bar charts and candles, misses in gaps, restoration to default, and unchanged vertical axis domain; the native visual of dense candles still needs to be accepted.

`BarFill(func(ChartBarDatum) ChartBarFill)` Configure padding on a per-column or per-stack segment basis. Parameters include original Series/Index, series name, category label, value, stacking flag, and default series color; returns Color or non-nil Gradient. `ChartBarGradient{Start, End, Direction}` Gradient to Top/Bottom/Left/Right along the current column segment rectangle. The default and illegal directions use Bottom, retaining transparency and rounded corners. Pass nil to restore the series color. This interface only affects the column and does not change the legend, tips, or data. Callbacks may be executed during measurement and plotting and must have no side effects.

The order chart example adds bar-by-bar gradient. The 1×/2× GPU pixel test covers four combinations of cylinder directions and four gradient directions, and also verifies that the stack callback retains the original positive and negative values. The gradient direction uses the screen direction; the cylinder baseline direction is configured through BarAlignment.

`HoverAnimation(false)` turns off hover transition, enabled by default. The emphasis state changes with three 150ms slow-outs, and the rapid target change continues from the current weight; the target is displayed immediately when the animation is reduced. Pie chart transition selected border, Sankey chart transition irrelevant stream band transparency, line/area chart transition focus line and point, bar/candle chart transition category background, radar chart transition focus. Hit and cue data are updated immediately, no waiting for animations. Update data to clean up old transitions. The shared animation test covers intermediate frame weights, continuity when switching targets, end frames and reduced motion; the look and feel of the native animation of each picture is still to be accepted.

## Cylinder direction

`BarAlignment(BarAlignmentBottom/Top/Left/Right)` specifies the side of the cylinder baseline: by default Bottom is upward, Top is downward, Left is to the right, and Right is to the left. Preserve zero baseline for positive and negative values, grouping/stacking, missing data, FutureSlots, rounded corners, and column-by-column padding. The text remains in the forward direction, the categories are arranged from top to bottom in the horizontal direction, and the value axis is in the left and right directions; hover hits are processed under the same coordinate transformation. Switching orientation clears old hovers. Non-column charts ignore this configuration.

The default reservations for landscape orientation are Left=80, Bottom=24, with explicit Gutter taking precedence. YLabelsInside moves the numeric scale into the plot area, with the category labels still on the left. The reference lines, tips and numerical scales use the coordinates of the corresponding directions, and the gradient Top/Bottom/Left/Right maintains the screen orientation.

The quarterly revenue example uses a horizontally stacked chart with a left baseline. Four-way hit and gradient pixels have passed automatic testing; full native layout of guides and text is still pending. See below for fixed total point position and multi-stop gradient.

## Fixed point position and numerical gradient

`PointCount(n)` The fixed category axis accommodates at least n positions (0–100000). When appending data, the existing category positions remain unchanged; after the actual data exceeds n, the axis is expanded and 0 returns to automatic. It is interchangeable with FutureSlots and works with line, area, column and candle charts. XTickCount selects ticks at a fixed total position, displays only text from existing categories, and does not forge labels, values, or tips for future empty spaces. For the same purpose as [GPUI's point_count](https://gpui-kit.com/component/chart/#pinned-axis-and-unfinished-series), Keel continues to use the classification band center coordinates.

`BarGradient(func(ChartBarDatum, ChartBarRange) []ChartColorStop)` Provides a multi-stop gradient along the cylinder's baseline to its tip. `ChartBarRange` includes display axis Min/Max and current segment cumulative value Base/Tip; `ChartToBar(value)` maps the value to the current segment, Base is 0, Tip is 1, and out-of-bounds are allowed. The positive and negative columns and the four directions are drawn according to the true growth direction. The datum.Value of stacked callbacks retains its original value.

Each stop contains Position and Color; to copy and sort the input, the repeated position adopts the last color. Points outside the interval are interpolated according to linear light and premultiplied alpha at the boundary; an empty array or a non-finite position set of fallback series colors. Adjacent color segments are divided according to physical pixel boundaries, and sub-pixel wide color segments are not retained. BarGradient and BarFill replace each other. Pass nil to restore the series color; the callback should have no side effects.

```go
chart.BarGradient(func(d kit.ChartBarDatum, r kit.ChartBarRange) []kit.ChartColorStop {
    return []kit.ChartColorStop{
        {Position: r.ChartToBar(r.Min), Color: theme.Chart[0]},
        {Position: r.ChartToBar((r.Min+r.Max)/2), Color: theme.Chart[1]},
        {Position: r.ChartToBar(r.Max), Color: theme.Chart[2]},
    }
})
```

Automatic verification covers fixed scale when appending, switching between two point configurations, excess data, interval extrapolation, input copies, illegal stops, extreme value mapping, positive and negative stacking endpoints; GPU pixel test covers the color consistency of double rate, four-way positive and negative columns, and translucent whole image mapping on different column heights. Native windows are still pending acceptance.
