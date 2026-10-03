# Bubble

聊天气泡。

```go
kit.Bubble(text).Mine()  // 自己发的：靠右、主色
kit.Bubble(text)         // 别人发的：靠左、浅色底
```

- 宽度最多占父容器的 75%。内容是任意 View。
- 只要气泡本身时用 Bubble；要带头像和操作栏的完整一条消息，用 Message。

验证：`go run ./examples/components -section bubble`，加 `-theme dark` 检查深色。

`Variant` 将外观与对齐分开：BubbleFilled（主色）、BubbleSecondary（次级底色）、BubbleMuted（弱化文字）、BubbleTinted（主题染色）、BubbleOutline（描边）、BubbleGhost（无框整行）、BubbleDestructive（错误色）。默认 BubbleAuto 保留原用法：Mine 使用主色，其余使用次级底色。Alignment(el.Start/el.End) 可切换对齐，Mine 等同 End；显式 Variant 不受 Mine 影响，非法枚举值忽略。

普通气泡的内容和反应区整体限制为父宽度的 75%，保留 Keel 原口径；Ghost 使用整行宽度并去掉默认背景、内边距、边框和圆角。普通内容内边距改用 SpaceLg/SpaceMd。Ghost 不添加裁剪区域，内容仍可使用自己的裁剪或阴影。

`Reactions(view)` 接收独立 View，nil 清除此槽；`ReactionSide(BubbleReactionTop/Bottom)` 设置位置，默认底部；`ReactionAlignment(el.Start/el.End)` 设置反应区对齐，默认靠右。反应区使用 Surface、Border 和胶囊圆角，内容自行管理计数、选中和回调，可放按钮或弹层触发器。反应区采用紧邻内容边缘的流式布局。

`PartStyle(BubblePartRoot/Content/Reactions, fn)` 在默认样式后分别调整整行、内容表面和反应表面；nil 恢复默认。回调接收每帧新建元素，不应保留引用。按钮与输入框的禁用、键盘和语义由子组件处理；通过 Root 的 Disabled 可禁用整组。切换外观、对齐和反应位置保留子控件身份。应在帧外创建并复用 Bubble 实例。

```go
b := kit.Bubble(content).Variant(kit.BubbleOutline).
    Reactions(kit.Button("赞同", like).Variant(kit.ButtonGhost).Size(24)).
    ReactionSide(kit.BubbleReactionTop)
```

连续气泡可用 [BubbleGroup](bubble_group.md) 组合，支持间距、样式、更新和组级禁用。

`ReactionActions(buttons ...*ButtonView)` 设置直接按钮列表，复制列表、忽略 nil，空参数清空。按钮仍由应用持有，后续 SetText/SetLoading/SetDisabled 会在下一帧生效；变体、大小、图标、自定义内容和回调保持原配置。气泡只在本次绘制中把这些按钮圆角设为 RadiusFull，不修改按钮实例，独立绘制时仍用 Button 的默认圆角。

只要有一个直接按钮，反应区就去掉默认内边距并按行排列，窄宽度时换行。`PartStyle(BubblePartReactions, ...)` 仍在默认配置之后执行，可显式恢复内边距或调整间距。`Reactions(view)` 的普通内容先于直接按钮显示，两者可共存，分别清空；普通槽中的按钮、Popover 等不会自动改圆角。同一个直接按钮实例在列表中只出现一次，复用实例可在重排时保留焦点。

```go
like := kit.Button("赞同 · 2", onLike).Variant(kit.ButtonGhost).Size(24)
copy := kit.Button("复制", onCopy).Variant(kit.ButtonGhost).Size(24)
b.ReactionActions(like, copy)
```
