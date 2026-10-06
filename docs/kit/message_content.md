# MessageContent

English | [简体中文](message_content.zh-CN.md)

Mix multiple bubbles, attachments, and normal views in the same message body. Bubbles passed directly will inherit the message alignment; when any direct bubble is a Ghost, the head and tail will be automatically indented. Explicit setting of HeaderInset/FooterInset still takes precedence.

```go
content := kit.MessageContent(
    kit.Bubble(kit.Label("Export complete")),
    kit.Attachment("orders.csv", 2048),
    kit.Bubble(kit.Label("The link expires in 24 hours")).Variant(kit.BubbleGhost),
)
msg := kit.Message("Assistant", content).Header(kit.Label("Just now"))
```

`SetItems` copies the new sequence and ignores nil, empty parameters are cleared; `Items` returns a copy. Children should be reused and appear only once per instance; custom views need to provide stable IDs. Persistent bubbles are rendered using an internal stable copy. Dynamic rearrangement retains its input and focus. The style and alignment of the original bubble will not be modified. Changing the source bubble's variant updates the style and Ghost inheritance on the next frame.

`Gap` sets non-negative finite dp, the default is `theme.SpaceMd`; `Style` configures the text stack, nil returns to default. Style callbacks are executed after default values and should not retain elements or add children. `SetDisabled` disables the entire body, and the disabling of the ancestor Message is also effective.

When installed directly through the Message constructor parameter or `Content(content)`, the component is responsible for mixing and will not include another layer of User bubbles. The independent Render defaults to the left and uses the receiver's default bubble color. Bubbles wrapped in ordinary Views will not be recursively detected; automatic Ghost inheritance is only for direct Bubble children.

Run: `go run ./examples/components -section message_content`, add `-theme dark` to check the dark theme.
