# ui/markdown

把 Markdown 渲染成 `ui/el` 元素，针对 AI 聊天的流式输出优化：只重新解析正在写的块，临时补全未闭合的语法，写完的块复用元素和布局。

- **依赖**：`el`、`core`、`theme`；第三方：goldmark（解析）、chroma（代码高亮）、Gio text.Shaper（字形排版）。
- **被谁依赖**：应用代码。

| 文件 | 内容 |
| --- | --- |
| `markdown.go` | `Doc`：切块、增量解析、流式补全 |
| `parse.go` | goldmark 语法树 → 中间结构（段落、代码块、列表、表格…） |
| `render.go` | 中间结构 → el 元素；富文本、代码高亮、块缓存 |
| `code.go` | 代码卡片、语言与操作图标、悬停提示、换行切换和横向滚动 |
| `math_parse.go` | 数学分隔符与常用 TeX 子集解析、源码回退 |
| `math_layout.go` | 公式盒子排版、分式根号、上下标和矩阵 |
| `text.go` | 富文本排版、字形坐标、装饰、链接和选区绘制 |
| `selection.go` | 文档坐标、跨块选区、整篇选中和纯文本复制 |

使用和设计：[Markdown](../../docs/markdown.md)。
