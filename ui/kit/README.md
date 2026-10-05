# ui/kit

现成组件层，基于 el 组织外观与交互，base 提供键盘导航、首字母跳转和多选行为。

- 使用和分类索引：[组件参考](../../docs/kit.md)。
- 实现要求：[组件开发规范](../../docs/component-development.md)。
- 配套示例：[组件库](../../examples/components/README.md)。

直接依赖 `core`、`theme`、`locale`、`el`、`base`，不依赖 `window`。SVG 解析和栅格化使用 `oksvg`、`rasterx`，支持范围见 [Icon](../../docs/kit/icon.md) 和 [Image](../../docs/kit/image.md)。

一个组件对应一个 `<name>.go`，共享字段外框在 `field.go`，共享表面和浮层样式由 `surface()`、`floating()` 提供。`conventions_test.go` 检查配套文档、示例、Agent 测试和公开 API 约定。

原子输入引用经 `el.InputDocument` 接入 `ui/internal/inputcontent`；kit 不直接依赖这个内部包。
