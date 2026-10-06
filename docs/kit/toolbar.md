# Toolbar

English | [简体中文](toolbar.zh-CN.md)

A row of command buttons, buttons that cannot fit will be automatically moved into the "More" menu.

```go
bar := kit.Toolbar(
    kit.ToolbarItem{Label: "新建", Icon: kit.IconPlus, Action: create},
    kit.ToolbarItem{Separator: true},
    kit.ToolbarItem{Label: "导出", Action: export},
)
```

- Each item can be a button or a divider. When `Icon` is set, the icon is displayed, and when `IconOnly` is set, only the icon is displayed with the name as a prompt; disabled buttons do not respond.
- Overflow judgment uses the width given by the parent container, so the toolbar must be placed in a position with width constraints, such as the entire row, or a container with a set width. When the parent container is width-limited by content, all buttons will be displayed.
- The command area only occupies one Tab stop, ← → moves between available buttons and "More", loops from beginning to end, Home / End jumps to beginning and end, press Enter or space to execute. Return to trigger buttons after the More menu is closed; left and right append actions retain their respective tab stops.
- `Leading(el.View)` / `Trailing(el.View)` fix the left and right areas, and the middle command overflows according to the remaining width. `Size(dp)` sets the uniform height, default is 32, minimum is 24.
- `SetDisabled` disables the entire toolbar and additional actions; `SetItemDisabled(index, bool)` disables individual items. Icon buttons are visible via hover or keyboard focus.
- `SetItems` replaces the button and closes the old menu; passing in a slice makes a copy, and `Items()` also returns a copy. The disabled state of the command with the same name is independent.

Agent: container role `toolbar`, the button is listed separately; the "More" button opens the ordinary `menu`.

Verify: `go run ./examples/components -section toolbar`, add `-theme dark` to check the dark theme.

When the width of the command area has not yet been measured in the first frame, the operation is to enter "More" first, and after the measurement, the buttons that can be accommodated are automatically expanded. The command area crops its own drawing and hits to avoid covering the right slot when it is first displayed or when the window is zoomed; there is a regression test for the first frame menu operation.

## Custom group anywhere

`ToolbarItem{Label: "缩放", Content: zoomSelect, Width: 150}` Place an interactive view at this location, and multiple controls can be combined inside the view. `Content` takes precedence over Action/Icon, Separator still takes precedence; Label is used for group semantics and overflow menus. `Width` refers to the group width. The legal range is greater than 0 and no more than 4096dp. If it is not set or illegal, 160dp is used.

For groups that cannot be placed, use Label to enter "More". After clicking, Content will be displayed in the modal overlay below the toolbar. `OverflowContent` can be used to provide another layout; the overlay is limited to the window, scrolled when the content is too high, Esc or click outside to close. Closes the overlay when the group is resized, disabled, or replaced by SetItems. The state of the custom view is held by the application; switching between the toolbar and the overlay will change the element tree path, and the internal state of the frame such as input selection is not guaranteed to be retained.

The custom group's child controls retain their respective tab stops and arrow key behaviors, and the toolbar's left and right/Home/End navigation still only manages command buttons. Disabling entire groups and entire toolbars is passed to child controls. The example adds a zoom selector between search and copy; automatic testing covers double overflow, nested Select, Esc, disabling, width change and replacement cleanup; the native vision is not accepted.
