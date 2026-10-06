# Empty

English | [简体中文](empty.zh-CN.md)

`kit.Empty("暂无订单").Description("新建订单开始").Icon(kit.IconInbox).Action(action)` displays the icon, title, description and operating area vertically centered. Action accepts an el.View and its Render is called every frame. The operation area usually uses `kit.Button`, and a custom view can also be passed in.

SetTitle, SetDescription update copy. Empty does not add new semantic roles, the title and description are used as text respectively, and the operation elements retain their own semantics and keyboard behavior. Narrow container text wraps. Verify: `go run ./examples/components -section empty`, add `-theme dark` to check the dark theme.

```go
kit.Empty("暂无订单").Action(el.ViewFunc(func(cx *el.Context) el.Element {
    return el.Div().OnClick(createOrder).Child(el.Text("新建订单").TextColor(theme.Primary))
}))
```

`Media(view)` Place an avatar, image, avatar group, or any custom content above the title, preserving the content's own size, semantics, and interactivity. The last call replaces the previous content; passing nil restores the `Icon` configuration, and `Icon(kit.IconNone)` hides the default icon. Media, title, and description changes do not re-establish the Action's input state or focus. The media still needs to fit the parent container width; scrolling is provided by the parent container.

```go
kit.Empty("暂无成员").Media(kit.Avatar("Alex").Size(48)).Action(kit.Button("邀请", invite))
```

`Heading(view)` and `DescriptionContent(view)` replace the string display of the title and description, and pass nil to restore the string; at this time, SetTitle / SetDescription updates the fallback copy. `Footer(view)` Add independent auxiliary content below Action, pass nil to remove. The semantics of rich content are provided by the content itself, without repeated declaration of replaced strings.

`PartStyle(part, func(*el.DivEl))` adjusts the partition after the default style every frame, supporting `EmptyPartRoot`, `EmptyPartHeader`, `EmptyPartMedia`, `EmptyPartTitle`, `EmptyPartDescription`, `EmptyPartContent` (Action), `EmptyPartFooter`. Background, border, rounded corners, width, spacing, alignment and inherited font size/color can be set; explicit styles of sub-content take precedence. Pass nil to restore the default style of the partition, and illegal parts are ignored. Callbacks do not retain element references; subpartition IDs are fixed by the component to preserve operation area state.

```go
kit.Empty("没有结果").
    Heading(kit.Tag("没有结果")).
    Footer(kit.Button("查看帮助", help).Variant(kit.ButtonLink)).
    PartStyle(kit.EmptyPartRoot, func(e *el.DivEl) {
        e.Bg(theme.Subtle).Rounded(theme.RadiusLg)
    })
```

The default layout still retains Keel's Surface background and spacing; Empty does not add automatic announcements or focus targets, and the content buttons and input boxes maintain their own behaviors.

## Default style

`Variant(kit.EmptyOutline)` displays a dotted rounded border and a transparent bottom, which is suitable for uploading and dragging areas; `Variant(kit.EmptyMuted)` uses `theme.Subtle` with a shallow bottom and rounded corners, which is suitable for placing in cards and panels. The default `EmptyPlain` maintains the original Surface background color. Further adjustments can still be made later using `PartStyle(kit.EmptyPartRoot, ...)`.
