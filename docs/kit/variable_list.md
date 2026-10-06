# VariableList

English | [简体中文](variable_list.zh-CN.md)

```go
list := kit.VariableList(messageIDs, 72, func(cx *el.Context, i int) el.Element {
    return el.Text(messages[i].Body) // Automatic line wrapping, natural height
}).Height(400)
```

The second parameter of `VariableList` is the estimated height (dp) of the row that has not yet been measured. Each frame only builds rows near the visible area and one screen above and below. The actual height is cached after layout, and the row position is queried using prefixes and indexes; normal scrolling does not scan all data.

- key must be non-empty and unique. `SetKeys(ids)` copies the new sequence, retaining the size cache of still existing keys; duplicate or empty keys will panic before modification.
- When the header is inserted or the height of the upper row changes, the screen position of the first visible key is maintained; when the row is deleted, the row near the original index is selected. Window width or zoom changes automatically clear the size cache and remeasure.
- Height changes of visible rows are automatically detected; `Invalidate(keys...)` is called after modifying off-screen data or fonts. If the key is not passed, all size caches will be invalidated.
- `ScrollTo(cx, i)` is positioned according to the zero-based index of the current list; `ScrollToKey(cx, key)` is positioned according to a stable key, which is suitable for message jump after insertion, deletion or rearrangement. Non-existing keys do not change the current positioning request. Both only scroll until the target row is visible, and continue to correct after initially displaying and measuring the row height. `ScrollToAlign(cx, i, align)` / `ScrollToKeyAlign(cx, key, align)` Put the target row at the top, center or bottom (`kit.ScrollStart/ScrollCenter/ScrollEnd`). The unmeasured rows are first positioned according to the estimated height and automatically corrected after measurement. `Height`, `Fill`, `Count`, `ID`, and `SetDisabled` are consistent with equal height list usage.
- Stable keys maintain the identity of elements within the build scope; business state such as input values that have long left the build scope should still be preserved in the application model.
- Unvisited rows use an estimated height, so the scrollbar length adjusts to the measurement. In scenarios where table row heights are known, continue to use equal height lists.

Verification: `go run ./examples/components -section variable_list`. Click "Locate Message 50000", then continuously perform "Header Insert 10" and reposition, you should always see "Message 50000". Also checks for expanded line content and narrow window wrapping.


`Horizontal(true)` Enables horizontally widening virtual lists; constructed with the estimate parameter as the unmeasured width. `Width(dp)` sets the viewport width (horizontal default is 320dp), `Height(dp)` sets the height; horizontal `Fill()` fills the main axis in the Row parent layout. Item renderers can return different natural or explicit widths.

Switching the axis retains the first visible stable key, resets the offset within the project to zero, and clears the old axis measurements; row heights are not treated as column widths. The pending ScrollTo/ScrollToKey is retained and will be modified after layout. `ScrollToEnd(cx)` locates the last item. Horizontal insertion, deletion, reordering, and size changes also maintain the reading anchor; cross-axis height and scaling changes in horizontal mode invalidate cache. Visible items are re-measured every frame, and Invalidate is still called when modifying off-screen content.

```go
strip := kit.VariableList(ids, 160, func(cx *el.Context, i int) el.Element {
    return el.Div().W(el.Dp(widths[i])).Child(el.Text(labels[i]))
}).Horizontal(true).Width(640).Height(100)
```

The same list is virtualized along only one axis. Keel uses actual measurements of visible items and estimates of unvisited items, and provides all sizes upstream through the caller; the length of the scroll bar may be adjusted with the measurement, and the positioning adopts the minimum exposure strategy.
