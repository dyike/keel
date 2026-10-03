# AttachmentGroup

横向排列附件，超出容器时滚动。每项保留自己的上传状态和操作回调。

```go
files := kit.AttachmentGroup(report, photo).Name("附件").Gap(12)
photo.OnRemove(func() { files.SetItems(report) })
```

默认间距为 SpaceSm，各项顶对齐且不压缩宽度。SetItems 和 Items 隔离切片修改，忽略 nil 条目；元素实例仍共享，同一个实例不要重复渲染。SetDisabled 禁止组内操作，不修改附件自身设置。

Agent：组角色为 group，名称由 Name 设置；内部附件和操作保留各自语义。

运行 `go run ./examples/components -section attachment_group`，横向滚动后移除附件。媒体和上传操作见 [Attachment](attachment.md)。

`ScrollTo(dp)` 请求绝对水平偏移，可在首次显示前调用；负值回到起点，NaN/Inf 忽略。请求在实际绘制后按当前内容限制到有效范围，隐藏期间保留。程序滚动即使组内操作禁用也可用，不触发附件选择或打开；后台调用应通过 core.Update 安排到 UI 线程。

`ScrollState(cx)` 返回水平偏移、可视宽度和内容宽度，单位 dp；首次绘制前为零。示例中的“前一屏/后一屏”按钮用当前偏移加减可视宽度。接口读取当前 root 的状态，同一组实例应只在一个位置渲染。

```go
next := kit.Button("后一屏", func() {
    offset, width, _ := files.ScrollState(cx)
    files.ScrollTo(offset + width)
})
```
