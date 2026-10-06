# Common drawing basic parts

English | [简体中文](plot_primitives.zh-CN.md)

`github.com/dyike/keel/ui/plot` provides custom chart scales, data layout and instant plotting. `kit.Plot` continues to provide finished images with zoom, pan, and picking; this package leaves the application in charge of layout, theme colors, input handling, and data description.

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

The Canvas uses local pixel coordinates, translates to Bounds.Min for each draw, and clips to Bounds. Should be created within Layout and not saved across frames. Sizes need to be converted by the caller by gtx.Metric; Axis.TextSize uses sp. When the source geometric coordinates are non-finite or the absolute value exceeds 10 million pixels, the drawing is skipped to avoid handing it over to the underlying graphics engine.

## Scale and data layout

- ScaleLinear: Supports reverse domain/value domain, Map, Invert, Clamp; constant domain is mapped to the midpoint. Ticks returns uniform ticks including endpoints, the number is limited to 2–1000, and no nice scale rounding is performed. Returns false for illegal input or unrepresentable extrapolation results.
- NewBand: Classification fields are deduplicated and copied, Map returns the low coordinate of the strip, and Bandwidth returns the width. Padding configures the internal and external spacing, and Align configures the remaining space allocation; the modification method returns the new scale. Inverse ranges invert the sorting order, but the width remains positive.
- NewPoint: Classification point position, the default is to align both ends, and single item is centered; Padding leaves space at both ends.
- NewOrdinal: Categories are mapped to discrete value fields used in cycles; copy the input slice, return false for missing categories/null value fields, and do not automatically expand the field. References in range elements are managed by the caller.
- Stack: Inputs 2D values organized by series, output retains series/sample index and Low/High/Valid. Positive and negative values are accumulated from zero respectively, NaN/missing samples retain invalid slots; infinite or cumulative overflow returns an error.
- Pie: The output retains the original index, value and starting and ending radians; it allows up to one revolution in the forward and reverse directions, zero values are omitted, and an error is reported for negative/non-finite values. Normalize by the maximum value to avoid overflow of large summation; the gap angle is limited to the available angles for each item. Zero-corner-width entries preserve the index but do not draw.

## Draw

Bar supports rectangles and rounded corners in any direction; Line supports straight lines, steps, smoothstep curves and data points; Area fills the area between the upper and lower boundaries of equal length, and can draw upper edges. Missing or illegal points break lines/areas and do not connect gaps. Arc draws sectors/rings, using finite segments to approximate arcs; Dot and CrossLine provide point and cross guides.

Axis supports four directions, up, down, left and right, customized scale labels, line width, color and font size; At is the axis position, and From/To are the ends of the baseline. Labels are also cropped by Canvas, and the caller needs to reserve margins. Low-level graphs and axis labels will not automatically generate Agent data nodes and should be surrounded by readable titles, data tables or text; the prompt box can be used with `kit.Tooltip` or a overlay combination. This package does not automatically handle hover animations, nor is it equivalent to upstream full style configuration.

Example combining stacked bars, trend lines, reference lines, axes and donut charts: `go run ./examples/components -section plot_primitives`.
