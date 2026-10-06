# Slider

English | [简体中文](slider.zh-CN.md)

Drag or use the keyboard to select a numeric or double-ended range.

```go
volume := kit.Slider("音量", 0, 100).Step(5)
price := kit.RangeSlider("价格区间", 0, 100).Step(5)
price.SetValues(20, 80)
level := kit.RangeSlider("竖向区间", 0, 100).Vertical(180)
```

- Drag the slider, or press and drag anywhere on the track.
- Keyboard: ← ↓ decreases one step, → ↑ increases one step, PageUp / PageDown moves 10 steps, Home / End jumps to both ends.
- `Step(s)` aligns the value to `min + k·s`; when it is 0, the value is taken continuously, and the keyboard moves 1% of the range each time.
- `RangeSlider` initially selects the entire range. `Values()` / `SetValues(low, high)` reads or sets both ends; the program settings will sort, align and clamp, and will not trigger `OnRangeChange`. When the user drags, he selects the closest end and stops when he touches the other end without exchanging endpoint identities.
- Each end supports Tab focus and keystrokes; Home / End are bound by the other end. When the upper and lower limits overlap, press on the overlap to select the upper limit, and you can pull it back to the right or upward; the lower limit can also be operated independently with Tab.
- `Vertical(height)` sets the vertical track height (dp), the bottom is the minimum value, the top is the maximum value; ↑ increases, ↓ decreases.
- `Value()` / `SetValue` (auto-aligned and limited to range), `SetRange`, `SetDisabled`.

Agent: The single-ended role is `slider`; the double-ended role is `slider` named after the localized "lower limit/upper limit + label", and `value` is the value of the end. `FocusID()` points to the lower limit in double-ended mode.

Verify: `go run ./examples/components -section slider`, add `-theme dark` to check the dark theme.

`Scale(kit.SliderLogarithmic)` uses logarithmic scale, single value/double-ended, horizontal/vertical can be used. Valid range requirement `0 < min < max`; if it is not met, it will be displayed on a linear scale, and then setting the valid range will restore the logarithmic scale. Default `SliderLinear`.

```go
frequency := kit.Slider("频率 Hz", 20, 20000).
    Scale(kit.SliderLogarithmic).
    OnRelease(func(value float64) { applyFrequency(value) })
```

For example, on a logarithmic range of 1–1000, the exact middle of the orbit is about 31.62 and one-quarter of the way around is about 5.62. When the positive step is not set, the direction keys move 1% of the track, and PageUp/Down moves 10%; the explicit `Step(s)` still increases, decreases, and aligns according to the original value. Double-ended dragging selects the closest endpoint by distance on the track.

`OnRelease(func(float64))` receives the value at the end of a single-value operation; `OnRangeRelease(func(low, high float64))` receives a double-ended range. Triggered once when the mouse/touch is released; only the value is changed when the navigation key is pressed repeatedly, and once when the last navigation key is released. Valid clicks without numerical changes will also trigger. `OnChange` / `OnRangeChange` still fire continuously on value changes.

Programmatic assignment, drag cancellation, key release after defocusing, and disabling operations do not trigger the end callback; cancellation does not undo the value changes that have occurred. You can use OnChange to update the preview and OnRelease to submit expensive operations. Pass nil to remove the corresponding callback.

`Appearance(func(*kit.SliderAppearance))` configures the track and slider appearance of the current instance; each rendering first reads the current theme default value, then runs the callback, and returns nil to restore the default. Double-ended appearance, the configuration will not change the value or trigger the value callback.

```go
price.Appearance(func(a *kit.SliderAppearance) {
    a.TrackSize, a.ThumbSize = 8, 24
    a.TrackRadius, a.ThumbRadius = 2, 4
    a.FillColor, a.ThumbBorderColor = theme.Success, theme.Success
})
```

Colors include `TrackColor`, `FillColor`, `ThumbColor`, `ThumbBorderColor`; sizes are dp. The default track is 4, the slider is 16, and the border is 2; the track/slider size must be a finite positive number, and invalid values fall back to the default, with a maximum of 1024dp. The border and rounded corners are allowed to be 0, and invalid values will fall back to the default; the border and slider rounded corners can be up to half the size of the slider; the track rounded corners can be up to 1024dp. The horizontal minimum width and vertical minimum height are the slider size, and pointer value mapping uses the distance between the slider centers. When disabled itself, the padding and slider borders change to theme Muted, and keyboard focus still shows the theme focus color.

Verified landscape orientation, double-ended range, double ratio layout, default restoration, disabling and keyboard operation, and overriding window pixel regression with custom colors.
