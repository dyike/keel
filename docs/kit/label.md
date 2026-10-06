# Label

English | [简体中文](label.zh-CN.md)

`kit.Label(text)` displays wrappable labels. `Secondary(text)` appends the secondary color copy after the same text stream and removes the empty string; `SetText` updates the main copy.

```go
kit.Label("公司名称").Secondary("（可选）").Highlights("公司")
kit.Label("账户余额").Masked(true)
kit.Label("Hello World").HighlightPrefix("Hello").Style(func(t *el.TextEl) {
    t.TextSize(20).Bold().LineHeight(1.5)
})
```

`Highlights` is case-sensitive and marks all non-overlapping exact matches; `HighlightPrefix` only matches the beginning, and the call after the two will take effect, and the empty string will be cleared. `HighlightColor` Overrides the theme PrimaryText. The secondary copywriter uses the current theme Muted and does not participate in the search.

`Masked(true)` replaces each Unicode rune of the main copy with a dot and turns off the main copy highlighting; both rendering and Agent semantics only contain dots. Secondary copy is still visible. Combining characters and emoji sequences may correspond to multiple dots and are not counted by glyphs.

`Style` receives a new TextEl for each frame, configurable font, weight, line height, alignment, width, MaxLines and FocusOnPress field focus associations. Do not retain a reference to the element or re-expose the masked text via a custom Name. nil restores the default style.

The underlying `el.Text(...).Ranges(...)` is colored with half-open rune intervals and does not split the entire section for typesetting. When matching involves ligatures/combining characters, the entire glyph cluster is colored; colored bitmap glyphs remain in their original color, and truncated ellipses use the base color. The latter takes precedence over overlapping intervals, and Shimmer takes precedence over Ranges.
