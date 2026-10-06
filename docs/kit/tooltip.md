# Tooltip

English | [简体中文](tooltip.zh-CN.md)

Add a short reminder to any view.

```go
copy := kit.WithTooltip(kit.Button("", doCopy).Icon(kit.IconCopy), "Copy (⌘C)")
```

- When to display: Displayed after the pointer stays at `kit.TooltipDelay` (500ms); displayed immediately when the keyboard focus enters the wrapped view.
- When to hide: pointer moves away, focus leaves, press Esc. After pressing Esc to hide it, it will not be displayed again until the pointer and focus leave once.
- Display position: Centered at the top by default, flip to the bottom if it cannot fit.
- The prompt itself cannot receive focus and does not respond to clicks.
- `SetText` is used to update the prompt text, such as changing it to "Copied" after copying.

Agent: The role is `tooltip`, the name is the prompt text; rich content and action keys are listed as child elements.

Verification: `go run ./examples/components -section tooltip`.

The hover delay is bound to the actual visible and enabled target, and will stop waiting when removed, disabled, or modally blocked; resume waiting for the full delay again. `SetDisabled(true)` Disables both the prompt and the target area. The prompt width will be limited by the current window. The delays of multiple Tooltips are calculated independently and the waiting time of the previous target is not used.

`Content(view)` replaces the display content and supports multi-line text, icons, etc.; the original text is retained as a semantic name and must be non-empty. Pass nil to restore plain text. Controls in rich content are disabled and will not gain focus or perform clicks; please use HoverCard for previews that require interaction.

`Action(name)` displays the first shortcut key for the action in `core.Bind`, and is automatically bound accordingly; it is not displayed when it is not bound, and is not responsible for registering or executing actions. `Placement(side, align)` sets direction and alignment, `Offset(dp)` sets spacing (default 4dp, ignore non-finite values); positioning is still constrained by window edge avoidance.
