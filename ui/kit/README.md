# ui/kit

基于 el 的新组件模块。M1 包含展示、选择/关闭交互和加载动画。焦点、禁用与时间能力由 el 提供。

- 直接依赖：`core`、`theme`、`el`。
- 不依赖：`widget`、`layout`、`window`。
- 调用者：后续迁移的应用视图和示例。

组件 API、文件模板和验收要求见[kit 组件规范](../../docs/kit.md)，阶段安排见[设计决策](../../docs/decisions.md#新组件基于-eluiwidget-冻结)。

| 组件 | 说明 |
| --- | --- |
| [Kbd](../../docs/kit/kbd.md) | 快捷键键帽，继承字号和平台格式 |
| [Button](../../docs/kit/button.md) | 操作按钮、图标、键盘与加载状态 |
| [Alert](../../docs/kit/alert.md) | 行内状态提示，支持浅深色 |
| [Empty](../../docs/kit/empty.md) | 空状态说明 |
| [Avatar](../../docs/kit/avatar.md) | 固定尺寸头像和姓名回退 |
| [Tag](../../docs/kit/tag.md) | 可选择、可移除标签 |
| [DescriptionList](../../docs/kit/description_list.md) | 固定标签列与自定义值 |
| [GroupBox](../../docs/kit/group_box.md) | 带标题的视图分组 |
| [StatusBar](../../docs/kit/status_bar.md) | 固定 24dp 左右状态栏 |
| [Marker](../../docs/kit/marker.md) | 纯图形标记 |
| [Icon](../../docs/kit/icon.md) | 矢量图标 |
| [Spinner](../../docs/kit/spinner.md) | 不确定进度与减少动画 |
| [Skeleton](../../docs/kit/skeleton.md) | 占位、圆形、Shimmer 扫光 |
