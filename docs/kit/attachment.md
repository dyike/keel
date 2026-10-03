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

`Media(view)` 用展示型 View 替换默认文件图标；nil 恢复图标。可传入 `kit.Image(pixels, alt).Size(width, height).Fit(kit.ImageCover)` 显示图片，图片加载仍由应用负责。横向媒体使用内容自身尺寸，最大宽度受卡片约束；横向预览宜用小缩略图，给文件名和操作留出空间。竖排默认使用方形预览，见下文。

`Vertical(true)` 将媒体放在文字上方、操作叠加在卡片右上角，`Vertical(false)` 恢复默认横排。打开期间切换布局保持打开区域的键盘身份；上传/失败时预览不会触发 OnOpen，取消、重试、移除保持独立。Media 用于展示，打开交互交给附件 OnOpen，避免在预览内嵌套按钮或另一个可点击图片。此接口自动添加媒体状态遮罩；上传/处理中标题显示文字扫光，尺寸档见下文。

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

默认 Medium 宽度从原来的 280dp 调整为 232dp。`PartStyle` 在尺寸默认值后应用，可覆盖宽度和媒体尺寸；自定义 Media/Content/Actions 中显式设置的尺寸保持不变。竖排默认方形预览和右上角操作；自定义比例见 MediaAspectRatio。

默认状态外观：Pending 使用虚线边框，Failed 使用 DangerText 边框；完成后恢复普通边框。未提供自定义 Media 时，失败显示危险色背景和图标：有 OnRetry 用错误图标，无 OnRetry 用禁止图标。自定义媒体上传时覆盖暗色遮罩和白色进度环，处理中显示不确定进度环；失败时遮罩加深，有 OnRetry 显示圆形重试按钮，否则显示禁止图标。完成、待上传和取消状态恢复原预览。遮罩不改变媒体尺寸，裁剪跟随 Media 分区圆角。媒体重试与操作区重试共用状态校验，先进入 0% 上传再通知应用，遵守卡片及祖先禁用；仅完成状态可打开。

分区样式在状态默认值之后应用，可通过 Root 的 `Border` 覆盖颜色/宽度，`BorderDashed(false)` 恢复实线；Media 可覆盖背景。底层 el 的 BorderDashed 同样适用于其他元素和状态样式，保持原有边框宽度与圆角，虚线为 4dp 实段和 3dp 间隔。

`MediaOverlay(view)` 在媒体区域居中叠加自定义 View，绘制于预览和内置生命周期遮罩上方，不参与媒体尺寸计算；nil 移除。可放播放按钮、徽标或自定义进度，内容应适配媒体大小，超出部分按媒体边界和圆角裁剪。无自定义 Media 时同样可叠加在默认图标上。

叠加层按钮有独立点击和键盘焦点，不触发附件 OnOpen，遵守附件和祖先禁用。展示内容或叠加层的空白处仍可打开已完成附件；上传、处理和失败期间不会打开附件。状态切换保留叠加层身份，应用可在 ViewFunc 中按 Status 自行决定显示内容。

```go
a.MediaOverlay(kit.Button("播放", play).Size(24))
a.MediaOverlay(nil) // 清除叠加层
```

默认标题在上传和处理中显示 ShimmerText 扫光，其他状态或减少动画时恢复普通文字；保留 Title 分区继承的字号、字重和行高。自定义 Content 替换默认标题，需要时可组合 kit.ShimmerText。

`PartStatus(part, status)` 为默认 Title 或 Description 设置独立展示状态；`ClearPartStatus(part)` 恢复继承附件当前有效状态。非法状态、其他分区忽略。Title 的上传/处理状态控制扫光；Description 的状态控制自动文案及失败配色。覆盖在父状态更新后继续保留，不修改附件本身的状态、媒体、进度条、打开/取消/重试逻辑或 Agent 生命周期值。

`Description(text)` 替换默认描述文字，同时保留有效描述状态的颜色；空字符串显示空文案。`ClearDescription()` 恢复自动大小/状态文案。覆盖为 Uploading 但附件没有有效上传进度时显示 0%。PartStyle 仍在状态配色之后应用。自定义 Content 替换整个默认元信息，此时标题/描述配置暂不显示；清除 Content 后恢复。

```go
a.SetError("当前版本上传失败")
a.Description("上一版本已上传").
    PartStatus(kit.AttachmentPartDescription, kit.AttachmentStatusComplete)
// 卡片仍处于失败状态，描述采用普通颜色；重试操作仍可用。
a.ClearPartStatus(kit.AttachmentPartDescription).ClearDescription()
```

`MediaAspectRatio(width/height)` 配置竖排预览比例，默认 1；例如 2 为宽度两倍于高度。0 恢复媒体自然尺寸，负值与非有限值忽略。竖排预览填满卡片内宽，已加载的 kit.Image 默认居中裁剪覆盖预览，不修改原 Image 实例；其他自定义 View 保持自身尺寸并居中。PartStyle(Media) 的显式高度优先于比例，也可用 AspectRatio 覆盖比例。

竖排操作区叠加在卡片右上角，默认偏移随尺寸档内边距变化；自定义 Root 内边距后，可用 PartStyle(Actions) 调整 Top/Right。操作背景取 Surface，按钮遵守禁用且不触发打开，横竖切换保留按钮焦点。操作区按内容宽度收缩，过多操作可能遮挡预览，应用应限制数量或用 PartStyle(Actions) 配置换行。

布局变化：原竖排自然预览及底部操作改为上述默认值。需要保留自然预览时设置 MediaAspectRatio(0)。独立示例：`go run ./examples/components -section attachment_vertical`。

底层 el 的 `AspectRatio(ratio)` 在宽度已解析、高度自动时按宽高比推导高度；显式高度和最大/最小高度限制优先，0 清除比例。它不为两个轴均自动的内容推导尺寸。

`ShowMedia`、`ShowContent`、`ShowActions` 独立控制媒体、元信息和操作区，默认均显示。隐藏后不占布局、不出现在 Agent 元素中，也不可聚焦/操作；再次显示使用原来的内容、样式、回调和状态配置。附件的文件名与生命周期语义始终保留。ShowActions 控制操作区内的自定义和内置按钮，媒体遮罩的重试按钮及 MediaOverlay 仍由 ShowMedia 控制。

竖排且显示媒体、隐藏元信息时，卡片成为纯图片 tile：去掉内边距与最小高度，预览铺满边框内侧，默认内圆角比卡片小 1dp。默认仍为方形，MediaAspectRatio 可改比例。自定义 Root 的边框/圆角后，可用 PartStyle(Media) 同步内圆角。只有操作区时采用普通流布局，避免操作悬在空预览上；三个区域全部隐藏时保留卡片外壳及语义。

```go
photo.Vertical(true).ShowContent(false) // 纯图片卡片
file.ShowMedia(false)                  // 只有元信息和操作
file.ShowMedia(false).ShowContent(false) // 只有操作区
photo.ShowContent(true)                // 恢复元信息
```
