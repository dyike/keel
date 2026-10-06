# Accordion

English | [简体中文](accordion.zh-CN.md)

Expandable sectioned panels. By default, only one section is expanded at the same time, and multiple sections can be expanded after `Multiple()`.

```go
faq := kit.Accordion().Add("How do I request a refund?", answer1).Add("When will the payment arrive?", answer2).Multiple()
```

- Titles can get focus: ↑ ↓ Home End Move between titles and skip disabled sections; press Enter or space to expand or collapse.
- `Value()` returns the sequence number of the expanded section; `SetValue(i...)` does not trigger a callback, and only the first one is retained in single section mode; `SetItemDisabled(i, bool)`.

Agent: container role `group`; each title is `disclosure`, `value` is expanded / collapsed; expanded content is listed separately.

Verify: `go run ./examples/components -section accordion`, add `-theme dark` to check the dark theme.

`Heading(i, view)` Customize the display title of item i. The original title is still used as the accessibility name. `Trigger(i)` and `Content(i)` can be entered into custom layouts separately, each rendered once; do not render the entire Accordion at the same time. Arrow keys will skip titles that are not rendered, disabled, or in a disabled container.

`SetDisabled` disables the entire group, `SetItemDisabled` disables both the item's title and expanded content, but retains the expanded state. Expand/collapse has 180ms animation, continuous switching reverses from the current height; reduced motion switches immediately. When folding, the input status is retained. If the focus is in the text, it will return to the title. The text cannot be continued during the folding process.

`Bordered(false)` hides the outer frame and section dividers, retains the background, rounded corners, and expanded state; it has borders by default. `Size` supports `AccordionSizeXSmall`, `AccordionSizeSmall`, `AccordionSizeMedium` (default), `AccordionSizeLarge`, and uniformly adjusts the inherited font size of titles, arrows, spacing, and text. The default file retains the original font size inheritance; the font size explicitly set for the custom title or body takes precedence.
