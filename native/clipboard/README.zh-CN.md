# clipboard

[English](README.md) | 简体中文

独立读取剪贴板的一次快照，不引用 UI 或 Gio。实现 macOS 14+ / cgo、Windows、Linux X11，以及（应用提供连接时的）Linux Wayland；其他平台和 iOS 返回 `native.ErrUnsupported`。

```go
clipboard.Read(func(data clipboard.Data, err error) {
    // 后台 goroutine；修改界面需经 core.Update。
    // data.Text 为文本，Images 为编码图片，Files 为文件路径。
})
```

- `Read` 立即返回，在主队列读取后通过后台 goroutine 调用完成回调；nil 回调不读取。
- 图片优先 PNG，否则 TIFF，保持编码数据，不解码或转换像素。文件 URL 转为路径；文件自带的图片预览不重复计作图片。
- 最多 128 个 pasteboard item，文本/路径 UTF-8 字节和图片编码数据合计不超过 16MiB。超限、序列化失败或读取期间剪贴板版本变化返回 `native.ErrFailed`，不返回部分结果。
- 不写剪贴板、不打开引用文件、不读取文件内容；空剪贴板成功返回空 Data。
- Linux Wayland：只有键盘焦点所在的客户端能看到剪贴板，所以由应用传入聚焦窗口的连接：`clipboard.UseWaylandDisplay(w.WaylandDisplay())`，焦点换窗口时重新传。没有连接或合成器不支持时退回 X11/XWayland。这一路径链接 libwayland-client，`-tags nowayland` 可去掉。
- **依赖**：native、native/internal/sys、native/internal/wlclip。应用可把结果适配为 core.ClipboardData，接到 Input/TextArea.PasteReader。

macOS 已实际读取图片快照；Windows、Linux 的读取已通过格式和协议测试，尚未在真机上验收。
