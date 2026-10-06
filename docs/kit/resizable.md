# Resizable

English | [简体中文](resizable.zh-CN.md)

Place a draggable divider between the two panels.

```go
split := kit.Resizable(fileTree, editor).Min(160, 320)
split.SetValue(240)                           // The width of the first panel in dp
stack := kit.Resizable(editor, terminal).Vertical() // arranged up and down
```

- Drag the divider to resize; the divider gets focus, the arrow keys move 16dp each time, and Home/End jumps to minimum or maximum.
- When the window size changes, the first panel maintains its size and the second panel takes up the remaining space. `Min(first, second)` limits the minimum size on both sides, defaulting to 80dp each.
- `SetDisabled(true)` disables the divider and content on both sides, removes focus and stops dragging, keystrokes, and user callbacks; `SetValue` remains resizable.
- `Value()` / `SetValue(dp)` (no callback is triggered), `OnChange(fn)` is called when the user drags or adjusts the keys.
- It will fill the space given by the parent container, so it must be placed in a place with a certain size.

Agent: The role of the divider is `separator`, the name is "resize", and `value` is the size of the first panel.

Verify: `go run ./examples/components -section resizable`, add `-theme dark` to check the dark theme.

`Max(first, second)` sets the maximum size on both sides, 0 means there is no limit on this side. Restrictions apply to dragging, arrow keys, Home/End, and SetValue. Negative/non-finite upper bounds are ignored on a per-argument basis; Min and SetValue also ignore non-finite values. When the upper limit is lower than Min on the same side, Min shall prevail.

When the container has enough space, the range on both sides is satisfied at the same time; when the sum of the upper limits on both sides is not enough to fill the container, the space behind the second panel is left empty. When the space is not enough to meet the minimum size, priority is given to retaining the minimum space of the second panel, and the first panel can be reduced to 0; when the container is smaller than the divider, the 6dp handle is still retained. Requesting convergence in the next frame after the window measurement changes, without calling OnChange. After configuring Min/Max to converge at the next Render, SetValue is immediately constrained by the currently known container size.

`Visible(first, second)` controls the visibility of both sides respectively, and both sides are displayed by default. When only one side is visible, the panel fills the container, the divider is hidden, and the split Min/Max is not applied temporarily; when both sides are hidden, the container is retained but the content is not displayed. The content of the panel continues to state that it cannot focus or receive input after hiding, and retains the input status after restoring; the original focus on the hidden side will not be automatically restored.

Window changes during showing and hiding do not overwrite the saved separation size and do not trigger OnChange. Convergence by current container and Min/Max when restoring dual panels. SetValue can still modify the recovery size during hiding. It is limited by the Min/Max of the first panel, and the container constraints are postponed to the double-panel recovery.

The handle's hit area is still 6dp, and the default visual linewidths are 1dp for idle, 2dp for hover/focus, 3dp for press, 4dp for drag. Dragging out the hit area will still maintain the dragging state; after releasing, the hovering or focus state will be restored, and the last effective size will be cancelled. The default line width transition is 150ms, comply with the reduced animation setting.

`HandleAppearance(fn)` The current theme default value is passed in each frame, and the `Idle/Hover/Pressed/Dragging` line width, `Color/ActiveColor` and `Duration` can be modified. Line width is limited to 0–6dp, 0 hides this state, non-finite values fall back 1dp; Duration≤0 switches immediately. nil restores default. Only line width gradient, status color switches immediately. This configuration does not change hit width and panel range.
