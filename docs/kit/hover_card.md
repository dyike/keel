# HoverCard

English | [简体中文](hover_card.zh-CN.md)

When the pointer stops on a certain view, a preview card pops up, such as personnel information and link summary.

```go
card := kit.HoverCard(nameView, profileView).Width(300).
    OpenDelay(200*time.Millisecond).CloseDelay(500*time.Millisecond).
    Placement(el.Right, el.Start).Offset(8)
```

- By default, the pointer opens after staying for 700ms and closes after leaving for 300ms; `OpenDelay` / `CloseDelay` are configured according to the instance, and negative numbers are treated as zero. The delay for which modifications are pending will be restarted from the time of modification.
- When the pointer moves from the target to the card, the card remains open, so links and buttons can be placed inside the card.
- Pressing Esc or clicking outside will close it immediately; it will be reopened only after the pointer and focus leave, to avoid stopping in place and popping up automatically.
- A focusable target opens immediately when it gains focus; it remains open while the focus is inside the target or card, and the card will not actively move the focus. The plain text target does not add a Tab stop. When keyboard access is required, pass in a Button or other focusable View.

Agent: The role of the card is `dialog`, and the elements inside are listed separately.

Verification: `go run ./examples/components -section hover_card`.

The card width and height are constrained by the window, and long content can be scrolled; non-finite widths are ignored. `SetDisabled(true)` Closes and disables the target area. The content can include overlays such as Menu, and the submenu is closed with a delayed pause during opening, so that the parent card will not be accidentally closed during the operation.

`Placement(side, align)` sets the preferred direction and alignment, `Offset(dp)` sets the anchor spacing (default 4dp, ignore non-finite values). The edges of the window are still automatically avoided; positioning modifications when it is open will take effect in the next frame. The delay only affects mouseover, keyboard focus is still turned on immediately.
