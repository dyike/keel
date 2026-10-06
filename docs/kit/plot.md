# Plot

English | [简体中文](plot.zh-CN.md)

Numerical x / y plots for interactive exploration of data.

```go
p := kit.Plot(
    kit.PlotSeries{Name: "sin(x)", Points: wave},
    kit.PlotSeries{Name: "Measured", Points: samples},
).Lines().Title("Signal")
```

- Mouse: The scroll wheel zooms with the pointer position as the center; drag and pan; double-click or click "Reset" to restore the display of all data.
- Keyboard: After getting focus, + / − zoom, arrow keys pan, 0 reset.
- When hovering, the nearest data point within the 12dp range will be selected, enlarged and displayed, and a prompt box will pop up to display the name and coordinates.
- The default is to draw scattered points; `Lines()` uses 2dp lines to connect the points and only magnifies the selected point. The color of the dot is `theme.Chart`, and there is a stroke of the background color outside, which can be distinguished even when overlapping.
- `View()` / `SetView()` reads or sets the current visible range, `Reset()` adapts to all data, `SetSeries` replaces data, and `Format(fn)` sets the numerical writing method.

Agent: Role `figure`; "Reset" is a button.

Verify: `go run ./examples/components -section plot`, add `-theme dark` to check the dark theme.

Construction and `SetSeries` copy series and point slices; points containing NaN/Inf coordinates do not participate in fitting, picking, or drawing, and the polyline breaks there. Empty data and completely invalid data display an empty status. `SetView` Ignore non-finite or inverted ranges.

Zooming keeps the endpoints limited and orderly, and stops zooming in after reaching the lower limit of relative accuracy (approximately 10⁻¹² of the current coordinate scale); if panning or zooming out will overflow, the original range will be retained. The scroll wheel single magnification limit is 0.01–100. Scientific notation is used by default for very large or very small values.

`SetDisabled(true)` disables the pointer, reset and keyboard; canceling dragging or disabling the area during dragging will return to the view where dragging started. Replacing data will clear old hover selections. Titles and legends wrap in narrow windows so that the tip width does not exceed the plot area.

Custom charts can use the independent `ui/plot` package, see [Public Drawing Basics](plot_primitives.md).
