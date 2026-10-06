# VirtualList

English | [简体中文](virtual_list.zh-CN.md)

A scrolling list of a large number of equal-sized items, only constructing items near the visual area; vertical rows and horizontal columns are supported.

```go
logs := kit.VirtualList(len(lines), 24, func(cx *el.Context, i int) el.Element {
    return el.Text(lines[i])
}).Height(300)
```

- `VirtualList` Each row has the same height (`rowHeight` in dp); natural height content uses `VariableList` below.
- Build one more screen before and after the viewable area along the scroll axis, so that the scrolling frame will not reveal any blank space.
- `Height(dp)` sets the visual height (default 320), and `Fill()` changes it to fill the space given by the parent container.
- `SetCount(n)` updates the number of rows; `ScrollTo(cx, i)` makes row i visible with minimum scrolling. When the list is not displayed yet (for example, in another tab), it will be scrolled when it is displayed for the first time. `ScrollToAlign(cx, i, align)` specifies the position: `kit.ScrollStart` top (left side when horizontal), `kit.ScrollCenter` center, `kit.ScrollEnd` bottom, `kit.ScrollNearest` is equivalent to ScrollTo; when it is close to the beginning and end, it is intercepted according to the range that the content can be scrolled to.
- Each row is wrapped in an element with a stable ID, and the state is not lost when the row is moved in the window.

Agent: List only the rows in the visible area.

Verify: `go run ./examples/components -section virtual_list`, add `-theme dark` to check the dark theme.

When the amount of data is reduced, the scroll range is restored according to the new content height; after the list is cleared and refilled, it can be displayed normally.


Natural height content uses [VariableList](variable_list.md), which supports stable keys, row height caching and reading position maintenance.

The equal-height list can be set to `ItemKey(func(i int) string)` to retain the element status of the constructed rows according to data identity during insertion and sorting. key must be non-null and unique; the index is used by default and the callback is only called within the build scope.


`Horizontal(true)` is changed to horizontal virtualization, and the rowHeight parameter becomes the width of each item during construction; false returns to vertical. `Width(dp)` sets the viewport width, and the horizontal default is 320dp; `Height(dp)` always sets the height, and the horizontal axis is the cross-axis size. `Fill()` fills the space allocated by the parent layout along the main axis, and is placed in the Row layout when horizontally. Explicit setting of main axis size cancels Fill, setting of cross axis size retains Fill.

Switching direction retains the scroll position and built item identity near the current first item, pending ScrollTo takes precedence; `ScrollToEnd(cx)` exposes the last item. All positioning interfaces follow the current axis, read the precise horizontal scroll status using `cx.ScrollStateX(list.ID())`, and set the offset using `cx.ScrollToX`. Data reduction reclaims out-of-range offsets.

```go
cards := kit.VirtualList(len(items), 120, func(cx *el.Context, i int) el.Element {
    return el.Text(items[i].Title)
}).Horizontal(true).Width(480).Height(80)
```

This is single-axis virtualization without virtualizing the 2D grid at the same time. Project content needs to fit into a fixed slot size; switching orientations does not automatically change the layout of the app content.
