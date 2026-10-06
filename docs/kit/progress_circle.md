# ProgressCircle

English | [简体中文](progress_circle.zh-CN.md)

Circular progress display, supporting real progress, uncertain animation and central content.

```go
p := kit.ProgressCircle("Import orders").Size(64)
p.SetValue(.42)
p.Child(el.ViewFunc(func(*el.Context) el.Element {
    return el.Text(fmt.Sprintf("%.0f%%", p.Value()*100))
}))
```

`Value/SetValue` Progress of reading and writing 0–1; assignment exits indeterminate mode. NaN zeroes, infinity and out-of-bounds values are limited to endpoints. `SetIndeterminate(true)` displays the arc of rotation and stops when the animation is reduced; switch back to determination mode to retain the previous value.

`Size` sets the diameter (default 48dp, ignores non-positive numbers and non-finite numbers); `Color` explicitly overrides the foreground color and switches with the theme by default. `SetLabel` updates the semantic name, `Child` sets the center view, and passes nil to clear it. Center content should be short and drawn clipped within the scope of the component. Under narrow constraints, the ring is drawn centered on the smaller value of available width and height.

Pure display component, does not receive keyboard focus. The Agent role is `progressbar`, the name is label, and the value is percentage or `indeterminate`. The application is responsible for updating state; background updates should go through `core.Update`.

Validation: `go run ./examples/components -section progress_circle`, with `-theme dark` to check dark colors, `-width 320 -scale 2 -screenshot /tmp/progress-circle.png` to check narrow layouts. Unit tests cover state and constraints, and window tests cover Agent values and animated pixels.

Determine the progress update using a 200ms smooth transition (`kit.ProgressDuration`), and continuous updates from the current display position. Displays the target value immediately upon first display, return from indeterminate mode, and reduced motion. `Value()`, Percent, and Agent semantics always return the target value, with graphical interpolation only; center custom content is managed by the caller. Animation uses the frame clock and does not create timers or goroutines.
