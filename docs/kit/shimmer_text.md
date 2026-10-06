# ShimmerText

English | [简体中文](shimmer_text.zh-CN.md)

Readable text is preserved, allowing the highlight to sweep across the glyphs; the background and white space between characters are unaffected.

```go
loading := kit.ShimmerText("正在生成内容……").Duration(2*time.Second).Spread(.3)
loading.Reverse(true).Once(true)
loading.Restart()
```

The default cycle is from left to right, with a cycle of 2 seconds, and the highlight half-width is 30% of the width of the text box. `Duration` accepts positive duration; `Spread` accepts (0,1], illegal values are ignored. `Reverse` reverses; `Once` restores normal text after one cycle and stops requesting animation frames, `Restart` starts again. Adjust the duration without resetting the starting point, and explicitly Restart when needed.

`Enabled(false)` displays normal text, and then enables Start from Scratch; it also displays normal text under the reduced animation setting. `SetText` updates the text, `Size` sets the font size (0 is inherited), `MaxLines` limits the number of lines (0 is not limited); the font, font weight, line height and color are inherited from the el container by default. `Color` covers the text background color, `Highlight` covers the highlight color, and the default highlight uses the theme PrimaryText.

Colored bitmap glyphs (such as some emoji) retain their original colors and do not participate in highlight shading. The component maintains text semantics and layout size, does not undertake loading tasks, and does not provide text selection. The underlying `el.Text(...).Shimmer(phase, spread, color)` is only responsible for drawing, making it easy to apply custom time control.

Run the example: `go run ./examples/components -section shimmer_text`.

`ShimmerStyle` is a reusable animation configuration value, including Duration, Spread, Reverse, and Once; `Style(config)` replaces four configurations at a time, and the text, font size, color, etc. remain unchanged. A value of zero reverts to default; non-positive Duration, beyond (0,1], or non-finite Spread use default. Valid configuration changes are replayed from the next frame, and repeated application of the same configuration does not restart the animation. Legacy methods such as itemized Duration/Spread still retain their original behavior.

```go
motion := kit.ShimmerStyle{Duration: 3*time.Second, Spread: .45, Reverse: true, Once: false}
text.Style(motion)
attachment.TitleShimmer(motion)
```
