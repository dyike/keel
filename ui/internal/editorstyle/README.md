# ui/internal/editorstyle

`ui/el` 输入框的绘制：按可见字形绘制选区背景和光标，保持光标与占位文案对齐。输入事件、选区范围、滚动、撤销和输入法仍由 Gio 的 `Editor` 处理。

- **依赖**：Gio、Go 标准图像和时间库、`golang.org/x/image/math/fixed`；不依赖其他 Keel 模块。
- **被谁依赖**：`ui/el`（kit 的输入框经 el 使用）。

`Caret.Layout` 只负责绘制；调用者须先处理 `Editor.Update`。测试同时覆盖编辑交互和 1× / 2× 的实际渲染像素。

选区沿用 `Editor.Regions` 的横向范围和基线，再按字形范围计算高度。绘制先记录原编辑器操作，画选区背景，再回放文字；不能靠整体移动文字来修正选区。定位步骤见[文字与选区对齐排查](../../../docs/troubleshooting.md#输入框文字贴着选区上沿选区下方留白过大)。

`Composition` 用编辑器排版后的 Regions 绘制组合下划线，返回视口内的组合范围；`Caret.InputMethodCaret` 提供与可见光标一致的基线和字形高度。自行消费输入法事件的输入适配层在 Layout 后调用，并负责把这些坐标发送给平台。此内部层不维护编辑事务。

InlineFont 为引用块构造仅含私用字符的空白字体。字形的 advance 参与 Gio Editor 的换行、选区和光标计算，UI 可在对应 Regions 中另行绘制内容。字体只加入单个编辑器的 shaper，不全局注册；调用方负责选择不与普通文本冲突的私用字符、测量宽度和布局缓存。已接入 Input/Textarea 的引用排版；无轮廓字形保留非零边界，防止行尾被当作空白忽略。

字体按 OpenType [cmap](https://learn.microsoft.com/en-us/typography/opentype/spec/cmap) format 12 和 [hmtx](https://learn.microsoft.com/en-us/typography/opentype/spec/hmtx) 构造，包含 sfnt 校验和；普通字符必须回退到正常字体。现有 go-text 使用有符号 advance，单字形设计单位限 32767；调用方需按字号和最大宽度选择 UnitsPerEm，不应直接把 dp 当设计单位。
