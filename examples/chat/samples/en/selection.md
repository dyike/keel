## Cross-block selection

Start dragging in the middle of this paragraph, continue across the next paragraph, the list, and the code, and finish inside the table. Dragging in reverse should select the same text.

The second paragraph contains **bold text**, *italic text*, `inline code`, and the [Go documentation](https://go.dev/doc/). Starting a drag on the link should select text without opening the page; clicking should open it.

> Quoted text should also be included in a cross-block selection.

1. First list item with `context.Context`.
2. Second list item with a **bold fragment**.

```go
func greet(name string) string {
    return "Hello, " + name
}
```

| Name | Value | Notes |
|:--|--:|:--:|
| Alpha | 12 | First row |
| Beta | 34 | Second row |

After selecting this answer, press Cmd/Ctrl+A and then Cmd/Ctrl+C. Paste to verify the complete plain text. Table columns should be separated by tabs, rows by newlines, and list markers and code-copy buttons should be excluded.

During streaming, scroll upward and select text that has already arrived. Continued output and completion should preserve the selection.
