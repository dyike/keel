# AttachmentGroup

English | [简体中文](attachment_group.zh-CN.md)

Arrange attachments horizontally and roll when out of container. Each item retains its own upload status and operation callbacks.

```go
files := kit.AttachmentGroup(report, photo).Name("附件").Gap(12)
photo.OnRemove(func() { files.SetItems(report) })
```

The default spacing is SpaceSm, items are top aligned and width is not compressed. SetItems and Items isolate slice modifications and ignore nil entries; element instances are still shared, and the same instance should not be rendered repeatedly. SetDisabled prohibits operations within the group and does not modify the attachment's own settings.

Agent: The group role is group, and the name is set by Name; internal attachments and operations retain their own semantics.

Run `go run ./examples/components -section attachment_group` and remove the attachment after scrolling sideways. See [Attachment](attachment.md) for media and upload operations.

`ScrollTo(dp)` requests an absolute horizontal offset, which can be called before the first display; negative values return to the starting point, NaN/Inf are ignored. The request is limited to the valid range according to the current content after actual drawing, and is retained during the hiding period. Program scrolling is available even if in-group operations are disabled and does not trigger attachment selection or opening; background calls should be scheduled to the UI thread via core.Update.

`ScrollState(cx)` Returns the horizontal offset, visual width, and content width in dp; zero before first drawing. The Previous/Next buttons in the example use the current offset to add or subtract the visual width. The interface reads the status of the current root, and the same group of instances should only be rendered in one location.

```go
next := kit.Button("后一屏", func() {
    offset, width, _ := files.ScrollState(cx)
    files.ScrollTo(offset + width)
})
```

`EdgeFade(color)` Draws a 24dp fade on the side with hidden content, the color should match the background around the group; does not show when the content is fully visible, hides the left/right fade when reaching the start/end point respectively. Narrow containers can take up to half the width on each side, leaving the bottom scroll bar uncovered. Fade only draws, does not add hit areas, and does not block attachment operations. `ClearEdgeFade()` is closed and not enabled by default; when the theme changes, the current theme background color can be re-introduced in Render.
