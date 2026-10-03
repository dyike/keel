# Attachment

附件卡片：文件名、大小、上传进度或错误，可以打开、移除。

```go
a := kit.Attachment("报价单.pdf", size).OnRemove(remove).OnOpen(open)
a.SetProgress(0.6)          // 上传中；负数表示已完成
a.SetError("超过 10 MB 上限")
```

- 大小用 `kit.FileSize` 格式化为 B / KB / MB / GB。上传中显示进度条和"上传中 60%"，出错时用危险色显示原因。
- "上传中"文字来自 locale。

Agent：角色 `attachment`，名字是文件名；`value` 为空、"上传中 60%"或 `error`；移除按钮名为"移除 文件名"。

验证：`go run ./examples/components -section attachment`，加 `-theme dark` 检查深色。

`OnCancel(fn)` 在上传中显示取消按钮；点击先标记已取消，再通知业务停止上传。`OnRetry(fn)` 在错误或取消后显示重试，点击先清除错误并将进度重置为 0，再调用业务回调。回调负责启动/停止真实传输；后台任务通过 `core.Update` 更新组件，并丢弃已取消任务的迟到结果。

`SetProgress` 清除此前错误和取消状态。进度大于 1 截为 1，NaN/Inf 忽略，负数标记完成。负文件大小显示为 0 B。只有完成且无错误的附件可以打开，取消和移除不会触发打开回调。`SetDisabled(true)` 禁止卡片内全部操作。取消、重试按钮的可访问名称包含文件名；取消状态的 Agent 值为 `canceled`。

`Media(view)` 用展示型 View 替换默认文件图标；nil 恢复图标。可传入 `kit.Image(pixels, alt).Size(width, height).Fit(kit.ImageCover)` 显示图片，图片加载仍由应用负责。媒体使用内容自身尺寸，最大宽度受卡片约束；横向预览宜用小缩略图，给文件名和操作留出空间。

`Vertical(true)` 将媒体放在文字上方、操作放在底部，`Vertical(false)` 恢复默认横排。打开期间切换布局保持打开区域的键盘身份；上传/失败时预览不会触发 OnOpen，取消、重试、移除保持独立。Media 用于展示，打开交互交给附件 OnOpen，避免在预览内嵌套按钮或另一个可点击图片。此接口不自动添加图片状态遮罩、扫光或尺寸档。
