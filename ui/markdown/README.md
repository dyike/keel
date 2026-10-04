# ui/markdown

把 Markdown 渲染成 `ui/el` 元素，针对 AI 聊天的流式输出优化：普通文档只重新解析正在写的块；含脚注、引用定义或宏时共享整篇解析上下文，临时补全未闭合的语法，写完的块复用元素和布局。

- **依赖**：`el`、`core`、`theme`、`locale`，以及内部的 `ui/internal/imageload`（图片）和经 el 间接依赖的 `ui/internal/editorstyle`；第三方：goldmark（解析）、golang.org/x/net/html（HTML）、chroma（代码高亮）、Gio text.Shaper（字形排版）。
- **被谁依赖**：应用代码。

| 文件 | 内容 |
| --- | --- |
| `markdown.go` | `Doc`：切块、增量解析、流式补全 |
| `plugins.go` | 文档级 Goldmark 扩展、块视图与行内原子控件工厂 |
| `parse.go` | goldmark 语法树 → 中间结构（段落、代码块、列表、表格…） |
| `render.go` | 中间结构 → el 元素；富文本、代码高亮、块缓存 |
| `code_extensions.go` | 代码块操作槽及按语言替换展示，保留源文档 |
| `code.go` | 代码卡片、语言与操作图标、悬停提示、换行切换和横向滚动 |
| `math_parse.go` | 数学分隔符与常用 TeX 子集解析、源码回退 |
| `math_more.go` | 扩展 TeX：更多符号与函数、数学字母表、重音、二项式、括号、颜色、方框、更多环境 |
| `html.go` | 行内 HTML 标签样式、HTML 块转 Markdown 块 |
| `math_macros.go`、`math_structures.go` | 文档宏、嵌套矩阵与配对分隔符 |
| `math_delimiters.go` | 自动伸缩分隔符绘制 |
| `references.go` | 文档级解析、脚注跳转与返回 |
| `images.go` | 异步图片资源共享与布局缓存失效 |
| `math_layout.go` | 公式盒子排版、分式根号、上下标和矩阵 |
| `text.go` | 富文本排版、字形坐标、装饰、链接和选区绘制 |
| `stream_fade.go` | 增量文字及样式淡入、独立片段计时与减少动画 |
| `preview.go` | 整篇行高预算、完整行裁剪及截断状态 |
| `ranges.go` | 渲染文本快照、UTF-8 区间高亮、变更迁移和最小纵向定位 |
| `selection.go` | 文档坐标、跨块选区、边缘自动滚动、整篇选中和纯文本复制 |
| `selection_units.go` | Unicode 选词、三击选段、公式与代码行边界 |

使用和设计：[Markdown](../../docs/markdown.md)。
