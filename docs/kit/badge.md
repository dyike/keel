# Badge

English | [简体中文](badge.zh-CN.md)

A number, dot or icon logo that can be displayed alone or hung on a subcomponent corner.

```go
unread := kit.Badge(3).Child(kit.Button("通知", open))
kit.Badge(1).Dot().Tone(kit.ToneSuccess).Child(avatar)
```

- In number and dot modes, hide when the count is less than or equal to 0. Displays "99+" when exceeding `Max` (default 99).
- When hanging on a sub-component, the corner mark is drawn on the corner without changing the size and position of the sub-component, and the count changes will not cause the layout to jump.
- `Icon(IconName)` switches to icon mode regardless of count and is displayed even if count is 0. `Icon(IconNone)` Restores numeric mode; Dot switches to dot and clears icon. Icon mode is in the lower right corner with a Surface-colored border; numbers and dots are in the upper right corner.
- `Size(dp)` Set number/icon height, default 18, accepts 12–128; recommended 12/18/24. The dots are scaled 8/18, and the width grows with the text when the numbers are too long. Illegal values are ignored.
- `Color(color.NRGBA)` specifies the background color, and the foreground automatically selects black and white according to the background color. `Tone` Sets the theme semantic color (default ToneDanger) while clearing the Color override. Fixed custom colors not automatically changing with the theme.
- `Name(string)` Sets the accessible name, appropriate for the icon state, such as "verified"; an empty string restores the original count name. `Value()` / `SetValue` operation count, does not change icon mode.

Agent: role `badge`, the default name is the original count, which can be overridden by Name; `value` is the displayed number ("99+"), `dot` or `icon`.

Verify: `go run ./examples/components -section badge`, add `-theme dark` to check the dark theme.
