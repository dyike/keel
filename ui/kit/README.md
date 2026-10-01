# ui/kit

基于 el 的新组件模块。M1 从展示组件开始，交互部分等待 el 基础接口 review。

- 直接依赖：`core`、`theme`、`el`。
- 不依赖：`widget`、`layout`、`window`。
- 调用者：后续迁移的应用视图和示例。

组件 API、文件模板和验收要求见[kit 组件规范](../../docs/kit.md)，阶段安排见[设计决策](../../docs/decisions.md#新组件基于-eluiwidget-冻结)。

| 组件 | 说明 |
| --- | --- |
| [Alert](../../docs/kit/alert.md) | 行内状态提示，支持浅深色 |
| [Empty](../../docs/kit/empty.md) | 空状态说明 |
| [Avatar](../../docs/kit/avatar.md) | 固定尺寸头像和姓名回退 |
| [Tag](../../docs/kit/tag.md) | 不可移除标签 |
| [DescriptionList](../../docs/kit/description_list.md) | 单列字段说明 |
| [GroupBox](../../docs/kit/group_box.md) | 带标题的视图分组 |
| [StatusBar](../../docs/kit/status_bar.md) | 状态与详情栏 |
| [Marker](../../docs/kit/marker.md) | 纯图形标记 |

Icon：矢量图标，默认颜色随主题切换。

[Spinner](../../docs/kit/spinner.md)：支持减少动画的不确定进度。

[Skeleton](../../docs/kit/skeleton.md)：占位、圆形和 Shimmer 扫光。
