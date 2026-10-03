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

`AttachmentGroup(items ...el.View)` 将附件排列为可横向滚动的一行，不压缩卡片宽度。`Gap(dp)` 设置非负间距，默认 SpaceSm；`Name` 设置组的可访问名称。组宽度填满父容器，各项顶对齐。

`SetItems` 替换列表，`Items` 返回副本，两者隔离切片修改并忽略 nil 条目；附件实例仍共享，保持自身上传状态和回调。同一实例不要在组中重复渲染。`SetDisabled` 禁止组内操作，不修改附件自身禁用设置。移除由应用调用 SetItems 完成；组只负责排列与滚动，不接管文件选择或上传任务。

```go
files := kit.AttachmentGroup(report, photo).Name("附件").Gap(12)
photo.OnRemove(func() { files.SetItems(report) })
```

`SetStatus(AttachmentStatus...)` 设置显式生命周期，`Status()` 读取当前有效状态。默认 Complete；可用 Pending、Uploading、Processing、Failed、Complete，以及 Keel 保留的 Canceled。状态值提供 IsPending/IsUploading/IsProcessing/IsFailed/IsComplete/IsInProgress 查询，IsInProgress 包含上传和处理。

Pending 显示“待上传”，Processing 显示“处理中”；上传和处理中默认媒体图标替换为转圈，可取消。Failed 无错误说明时显示“上传失败”，可重试；仅 Complete 可打开。文字随 locale 切换，Agent 新增 pending/processing 值，旧有上传、error、canceled 和空值保持兼容。

显式 SetStatus 清除错误和取消状态；进入 Uploading 保留有效进度或从 0 开始，其他状态清除进度。SetProgress 非负值进入 Uploading（包括 1），负值进入 Complete；完成传输后还需处理时显式设 Processing。SetError 临时覆盖当前状态，清空错误恢复此前状态；显式 Failed 需 SetStatus 或重试退出。取消状态继续优先于错误显示。所有程序状态更新不调用操作回调，重试先进入 0% 上传再通知应用。

`Content(view)` 替换默认文件名、状态描述和进度条；nil 恢复默认。启用 OnOpen 时这里应使用展示内容，交互控件放进 `Actions(views...)`。自定义元信息自行读取 Status 并显示所需状态，卡片本身的 Agent 名称和生命周期值保持不变。

`Actions` 复制传入切片，忽略 nil，在内置取消/重试/移除之前添加控件。空参数清除自定义控件；清除 OnCancel/OnRetry/OnRemove 回调可去掉对应内置按钮。自定义操作不触发 OnOpen，遵守卡片和祖先禁用状态。

`PartStyle(part, func(*el.DivEl))` 调整 Root、Media、Content、Title、Description、Actions 六个分区，常量统一以 AttachmentPart 开头。可设置背景、边框、圆角、间距、字号和颜色，也可用 Hidden 隐藏可选区域。样式在默认值之后应用，nil 恢复默认；元素每帧重建，不应保存引用或在样式回调里添加子内容。Root 的 ID、角色、名称、生命周期值及窗口最大宽度由组件保持。Title/Description 只作用于默认元信息，Content 自定义时由应用控制内部样式。

`Size(AttachmentSize...)` 选择 XSmall、Small、Medium、Large 四档，默认 Medium。卡片宽度分别为 176/200/232/272dp，默认媒体边长 28/32/38/44dp，标题字号 11/12/13/14sp；最小高度为 40/48/56/64dp，内容较多时继续增高。内边距、间距和内置操作按钮随档位调整。

默认 Medium 宽度从原来的 280dp 调整为 232dp。`PartStyle` 在尺寸默认值后应用，可覆盖宽度和媒体尺寸；自定义 Media/Content/Actions 中显式设置的尺寸保持不变。竖排仍使用内容自身的预览比例，操作区仍位于底部，与上游默认方形预览及右上角操作布局不同。
