# inputcontent

原子引用输入的纯内容层，由 el.InputDocument 使用。保存 UTF-8 提交文本、引用 ID/显示名称/范围，以及选区和撤销事务；不依赖 Gio 或其他 Keel 包。

Content 是独立草稿，引用范围必须匹配原文并落在字素边界。Presentation 映射提交文本与显示名称的字节坐标。Session 按明确编辑区间处理替换，避免文本 diff 漏掉“相同文字替换但移除引用”的操作；输入法组合事务只记一次撤销。

UI 层负责焦点、只读/禁用、平台事件、复制粘贴、命中与绘制。这里的组合输入测试只证明事务行为，不证明原生候选窗口或组合下划线显示正确。

LayoutPresentation 可为引用指定编辑器内部的显示替代串，例如用于占位字形的单个私用字符。它不修改提交文本和引用元数据；用 SourceRange 将显示编辑范围映射到原文，再调用 Session.ReplaceSource，可沿用原有组合输入和撤销事务。复制仍从 Session.SelectedText 取原文，不能复制布局占位串。已接入 Input/Textarea 的对象布局事件通路；平台 IME 使用可读标签投影，内部编辑器使用对象投影，两者分别映射到原文。
