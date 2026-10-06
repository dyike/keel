# DescriptionList

English | [简体中文](description_list.zh-CN.md)

`kit.DescriptionList().Item("Order number", "SO-1001").ItemView("Status", view).LabelWidth(96)` displays labels and values in rows, the default label column is 96dp, the value column takes up the remaining width and wraps, and the top is aligned. Narrow containers allow label columns to shrink.

Plain text entries expose a text to the Agent named "tag:value". ItemView accepts an el.View and calls its Render every frame, retaining the semantics and interactions of the subviews and making the labels individually readable. SetItems replaces all text items, and the parameter is a Description list.

The component itself has no keyboard operation. Verification: `go run ./examples/components -section description_list -theme dark`, omit theme to see the light theme.

`Columns(n)` sets the number of columns per row of entries, at least one column. `Span(n)` sets the span of the recently added entries, which is limited to 1 to the current number of columns during layout; it will wrap when the remaining columns cannot fit, and the vacancies in the previous row will not be backfilled. `Separator()` inserts a full line separator, with subsequent entries starting on a new line. `SetItems` will clear previous rich content, spans and dividers.

```go
kit.DescriptionList().Columns(2).Vertical().Bordered(true).
    Item("Order number", "SO-123").Item("Customer", "Alex Chen").
    Separator().Item("Notes", "Description spans two columns").Span(2)
```

`Vertical()` Places each entry's label above the value; column number and span still work. The default is horizontal layout, `LabelWidth` only takes effect in horizontal layout.

`Bordered(true)` adds padding and theme borders to each entry, with no borders by default. `Size(sp)` adjusts the text size and spacing file. `theme.TextSm`, `theme.TextBody`, and `theme.TextLg` are recommended; the text size is inherited by default. Size ignores non-positive and non-finite values, label width is allowed 0, but negative and non-finite values are ignored.

The number of columns is set by the caller and will not automatically switch according to the window width; text in a narrow container is wrapped. `Columns(1)` is available for small screens, rich content's own fixed minimum width must still fit within the container. Textual semantics are preserved across multiple columns and across columns, and rich content buttons preserve keyboard, click, and disabled inheritance.
