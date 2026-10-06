# Tag

English | [简体中文](tag.zh-CN.md)

`kit.Tag("紧急").Tone(kit.ToneDanger)` Display dye capsule label. Tone supports ToneNeutral, ToneInfo, ToneSuccess, ToneWarning, ToneDanger, and the default is ToneNeutral; the background is mixed with the current theme Surface and grade color, and SetText updates the copy.

```go
label := kit.Tag("已验证").Tone(kit.ToneSuccess).Outline(true).Size(20).Rounded(4)
```

- `Outline(true)` uses a stroke and a transparent background; uses SelectedBackground to indicate status when selected.
- `Size(dp)` sets the minimum height of the content area, minimum 16dp; 20/28/32 suitable for compact, standard and large labels. Small reduces the font size and padding at the same time; long text still wraps, and strokes can increase the outer size. The original natural height is retained by default.
- `Rounded(dp)` specifies rounded corners, 0 is right angle; the default is `theme.RadiusFull` capsule. Dimensions and rounded corners ignore illegal and non-finite values.
- `Content(el.View)` replaces the visible text, and restores it by passing nil. Transmit the display content without nesting interactive controls; the text during construction is still used as the Agent name and select/remove action name.
- `Appearance(func(TagAppearance) TagAppearance)` modifies Background, Foreground, Border, and SelectedBackground based on the theme color matching of each frame. Transparent colors are valid, non-transparent custom Borders will display borders; passing nil restores the theme. The callback is executed after Outline's default transparent background, so the background can be overridden explicitly.

Agent role tag, the name is the construction copy, and the value retains the Tone name; the custom color does not change the semantic level. The presentation-only version does not respond to the keyboard. `Selectable().OnChange(fn)` enables selection; `Value/SetValue` queries and program assignments. Programmatic assignments do not trigger callbacks. `OnRemove(fn)` adds an independent remove button, Space/Enter or Backspace/Delete is triggered after Tab is focused; the caller is responsible for deleting data, and the component will not hide itself. Both `SetDisabled` and the ancestor Disabled prohibit interaction, and the Agent reports disabled and selected. Removing does not also toggle the selected state.

Multiple tags are combined using el.Wrap. Verification: `go run ./examples/components -section tag -theme dark`, omit theme to see the light theme. Examples show stroke, size, angle, color, rich content, selection, and removal.
