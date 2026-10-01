# ui/internal/editorstyle

`ui/el` 输入框的绘制：按可见字形绘制选区背景和光标，保持光标与占位文案对齐。输入事件、选区范围、滚动、撤销和输入法仍由 Gio 的 `Editor` 处理。

- **依赖**：Gio、Go 标准图像和时间库、`golang.org/x/image/math/fixed`；不依赖其他 Keel 模块。
- **被谁依赖**：`ui/el`（kit 的输入框经 el 使用）。

`Caret.Layout` 只负责绘制；调用者须先处理 `Editor.Update`。测试同时覆盖编辑交互和 1× / 2× 的实际渲染像素。

选区沿用 `Editor.Regions` 的横向范围和基线，再按字形范围计算高度。绘制先记录原编辑器操作，画选区背景，再回放文字；不能靠整体移动文字来修正选区。定位步骤见[文字与选区对齐排查](../../../docs/troubleshooting.md#输入框文字贴着选区上沿选区下方留白过大)。
