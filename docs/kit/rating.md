# Rating

English | [简体中文](rating.zh-CN.md)

Star rating.

```go
score := kit.Rating("Rating", 5).OnChange(func(n int) { … })
avg := kit.Rating("Average", 5).ReadOnly()
avg.SetScore(3.7)
```

- Click on the unfilled i-th star to set it to i point; click on the filled star to set it to i-1 point (for example, at 4 minutes, the second star becomes 1 point, and the first star is cleared to zero); after getting the focus, adjust ← →, and Home / End jumps to 0 or full points. Preview the score that will be set on hover.
- `Score()` / `SetScore(float64)` Keep decimal ratings, display half stars or any decimals with a star's horizontal fill ratio. In `ReadOnly()`, you cannot click or use keys to change; in edit mode, you still press the whole star to select. NaN is returned to zero, the infinite value is clamped to 0 or full score, and the program setting does not trigger a callback.
- `Value()` returns the integer part and is compatible with existing integer scores; use `Score()` to read the average score.
- `Value()` / `SetValue` (limited to 0–max), `ReadOnly()`, `SetDisabled`.

Agent: role `slider`, `value` is "score/full score", such as `3/5`.

Verify: `go run ./examples/components -section rating`, add `-theme dark` to check the dark theme.

`Size(dp)` sets the size of each star, range 8–128dp, default 22; illegal values are ignored and the rating is not changed. `Color(color.NRGBA{...})` changes the fill color, defaults to Warning with the theme, and the outline remains Muted. Color is also used for hover previews and decimal fills. The score after the hover preview is clicked. The actual value and callback are only changed during operation; when the size is larger or the number of stars is large, the parent container provides sufficient width.
