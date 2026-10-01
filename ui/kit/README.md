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
