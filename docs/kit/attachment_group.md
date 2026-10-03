# AttachmentGroup

横向排列附件，超出容器时滚动。每项保留自己的上传状态和操作回调。

```go
files := kit.AttachmentGroup(report, photo).Name("附件").Gap(12)
photo.OnRemove(func() { files.SetItems(report) })
```

默认间距为 SpaceSm，各项顶对齐且不压缩宽度。SetItems 和 Items 隔离切片修改，忽略 nil 条目；元素实例仍共享，同一个实例不要重复渲染。SetDisabled 禁止组内操作，不修改附件自身设置。

Agent：组角色为 group，名称由 Name 设置；内部附件和操作保留各自语义。

运行 `go run ./examples/components -section attachment_group`，横向滚动后移除附件。媒体和上传操作见 [Attachment](attachment.md)。
