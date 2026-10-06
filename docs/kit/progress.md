# Progress

English | [简体中文](progress.zh-CN.md)

Labeled progress bar.

```go
p := kit.Progress("Import orders")
p.SetValue(0.42)          // Show 42%
p.SetIndeterminate(true)  // When the total amount is unknown, a progress block sliding back and forth is displayed.
```

- `SetValue` will clamp the value to 0–1 (NaN returns to zero) and exit indeterminate mode. When the reduced motion is turned on, the indeterminate progress is displayed statically.

Agent: role `progressbar`, `value` is percentage or `indeterminate`.

Verify: `go run ./examples/components -section progress`, add `-theme dark` to check the dark theme.

See [ProgressCircle](progress_circle.md) for circle progress and center content.

`Height(dp)` sets the bar height, range 1–128dp, default 8; `Color(c)` overrides the fill color and turns off the theme gradient of this instance; `Rounded(dp)` sets the track and progress block rounded corners, 0 is right angle. Illegal values are ignored. `TrackStyle(func(*el.DivEl))` sets the track background, border, etc. after the default style of each frame; nil restores the default and does not retain element references. Labels and percentages are not affected by track styles; the progress block fills the inner height of the track. Configuration is available in both deterministic and indeterminate modes.

Determine the progress update using a 200ms smooth transition (`kit.ProgressDuration`), and continuous updates from the current display position. Displays the target value immediately upon first display, return from indeterminate mode, and reduced motion. `Value()`, Percent and Agent semantics always return the target value, only graphical interpolation. Animation uses the frame clock and does not create timers or goroutines.
