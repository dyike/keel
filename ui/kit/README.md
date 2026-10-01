# ui/kit

基于 el 的新组件模块。M0 只登记目录和边界，尚无组件；M1 开始增加展示组件。

- 直接依赖：`core`、`theme`、`el`。
- 不依赖：`widget`、`layout`、`window`。
- 调用者：后续迁移的应用视图和示例。

组件 API、文件模板和验收要求见[kit 组件规范](../../docs/kit.md)，阶段安排见[设计决策](../../docs/decisions.md#新组件基于-eluiwidget-冻结)。
