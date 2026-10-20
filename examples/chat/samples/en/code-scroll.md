## Horizontal code scrolling

The code block shows its language on the left and wrap and copy buttons on the right. Hover over the icons for tooltips; after copying, the icon changes to a check mark.

```text
Queued 2 items · send after this run · Backspace to retrieve
  › update the documentation too
  › run go test ./pkg/example/ when done

running · Enter queues · Esc interrupts
────────────────────────────────────────
›
```

Narrow the window, then swipe horizontally with a trackpad or drag the bottom scrollbar until the following block reaches its right edge. With a mouse, place the pointer over the bottom scrollbar and use the wheel, or click the track to jump. The first line must remain on one line and reveal END_OF_LONG_LINE. The second line contains spaces and checks that automatic wrapping stays off.

```go
const unbroken = "ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz_0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz_0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz_0123456789_END_OF_LONG_LINE"
fmt.Printf("this line has spaces and a 中文 comment; it should stay on one visual line inside the code block, even when the window is narrow; the final marker is END_OF_SPACED_LINE")
short := "short-line control"
```

The second code block should keep an independent horizontal position. Its language label and copy button remain visible.

```text
path=/workspace/projects/markdown/examples/chat/fixtures/deeply/nested/directories/with/a/very/long/file/name/that/must/remain/on/one/line/independent_horizontal_scroll_position_END_OF_SECOND_BLOCK
```

Click the code copy button and paste. No newline should have been inserted into the long line. Ordinary prose should still wrap with the window, and vertical message scrolling should continue to work.

Click "Wrap lines" to wrap long lines inside the code block and remove horizontal scrolling. Click "Do not wrap" to restore the original lines. Each block keeps its own setting, including while streaming appends content.
