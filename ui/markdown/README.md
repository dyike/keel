# ui/markdown

把 Markdown 渲染成 `ui/el` 元素，针对 AI 聊天的流式输出优化：只重新解析正在写的块，临时补全未闭合的语法，写完的块复用元素和布局。

- **依赖**：`el`、`core`、`theme`；第三方：goldmark（解析）、chroma（代码高亮）、gioui.org/x/richtext（行内混排）。
- **被谁依赖**：应用代码。

| 文件 | 内容 |
| --- | --- |
| `markdown.go` | `Doc`：切块、增量解析、流式补全 |
| `parse.go` | goldmark 语法树 → 中间结构（段落、代码块、列表、表格…） |
| `render.go` | 中间结构 → el 元素；富文本、代码高亮、复制按钮 |

使用和设计：[Markdown](../../docs/markdown.md)。
