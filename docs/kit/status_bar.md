# StatusBar

English | [简体中文](status_bar.zh-CN.md)

`kit.StatusBar().Left(status).Right(position)` Creates a fixed 24dp status bar with a top 1dp border and default 12sp Muted text. The left and right content are grouped, the middle space is opened by Grow, and long text is limited to a single line. The parent view determines the status bar position.

Left/Right accepts el.View, replaces the respective view group, and calls the subview's Render every frame. Simple content can be wrapped with el.ViewFunc:

```go
status := el.ViewFunc(func(cx *el.Context) el.Element {
    return el.Text("Ready").TextColor(theme.Muted)
})
```

## Overflow menu

When all the content cannot fit in the narrow window, use `Add` (left group) and `AddRight` (right group) to add items that can be collapsed:

```go
bar := kit.StatusBar().Left(ready).
    Add(kit.StatusItem{Label: "main branch", Action: switchBranch, Priority: 3}).
    AddRight(
        kit.StatusItem{Label: "Line 12, column 4", Action: gotoLine, Priority: 2},
        kit.StatusItem{Label: "UTF-8", Action: pickEncoding},
    )
```

- When it cannot be put down, the one with the lowest `Priority` will be put into the "..." menu on the right first; if it has the same priority, the one with the lower `Priority` will be put first. The space vacated after folding a wide item will bring back the narrow item that was previously folded.
- After the window is widened, the collapsed items automatically return to the status bar.
- `Left`, `Right` The incoming view is always displayed and will not be collapsed. If the text is too long, it will be truncated.
- When `View` is empty, the status bar displays `Label`; when `Action` is present, it is a clickable button. `Label` is always displayed in the menu, click to run `Action`.
- The status bar must remember the size of each item according to its width, so it must be saved in the view for reuse instead of creating a new one every frame.
- `Hidden()` returns the `Label` of the currently collapsed item.

Agent role status, child elements are listed separately; "..." is a button named "More".

Verification: `go run ./examples/components -section status_bar -theme dark`, omit theme to see the light theme.
