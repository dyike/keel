# Tag

`kit.Tag("紧急").Tone(kit.ToneDanger)` 显示染色胶囊标签。Tone 支持 ToneNeutral、ToneInfo、ToneSuccess、ToneWarning、ToneDanger，默认 ToneNeutral；背景由当前主题 Surface 与等级色混合，SetText 更新文案。

```go
label := kit.Tag("已验证").Tone(kit.ToneSuccess).Outline(true).Size(20).Rounded(4)
```

- `Outline(true)` 使用描边和透明背景；选中时用 SelectedBackground 表示状态。
- `Size(dp)` 设置内容区最小高度，最小 16dp；20/28/32 适合紧凑、标准和大号标签。小号同时缩小字号与内边距；长文本仍换行，描边可增加外部尺寸。默认保留原自然高度。
- `Rounded(dp)` 指定圆角，0 为直角；默认 `theme.RadiusFull` 胶囊。尺寸和圆角忽略非法及非有限值。
- `Content(el.View)` 替换可见文案，传 nil 恢复。传展示内容，不嵌套交互控件；构造时的文字仍作为 Agent 名称及选择/移除动作名称。
- `Appearance(func(TagAppearance) TagAppearance)` 在每帧主题配色基础上修改 Background、Foreground、Border、SelectedBackground。透明颜色有效，非透明自定义 Border 会显示边框；传 nil 恢复主题。回调在 Outline 默认透明背景之后执行，因此可以显式覆盖背景。

Agent 角色 tag，名字是构造文案，value 保留 Tone 名称；自定义颜色不改语义级别。纯展示版本不响应键盘。`Selectable().OnChange(fn)` 启用选择；`Value/SetValue` 查询和程序赋值，程序赋值不触发回调。`OnRemove(fn)` 添加独立移除按钮，Tab 聚焦后 Space/Enter 或 Backspace/Delete 触发；调用方负责删除数据，组件不会自行隐藏。`SetDisabled` 和祖先 Disabled 均禁止交互，Agent 报告 disabled 与 selected。移除操作不会同时切换选中状态。

多个标签用 el.Wrap 组合。验证：`go run ./examples/components -section tag -theme dark`，省略 theme 查看浅色。示例展示描边、尺寸、直角、配色、富内容、选择和移除。
