# ui/internal/editorstyle

供 `ui/widget` 和 `ui/el` 共用的输入框绘制：按字形高度显示光标，保持光标与占位文案对齐。输入事件、选区、滚动、撤销和输入法仍由 Gio 的 `Editor` 处理。

- **依赖**：Gio、Go 标准图像和时间库、`golang.org/x/image/math/fixed`；不依赖其他 Keel 模块。
- **被谁依赖**：`ui/widget`、`ui/el`。

`Caret.Layout` 只负责绘制；调用者须先处理 `Editor.Update`。测试同时覆盖编辑交互和 1× / 2× 的实际渲染像素。
