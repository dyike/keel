# Chat 示例

预设回答模拟 AI 流式输出，支持停止、消息区跟随到底和 Markdown 交互验证。无需模型 API。

从仓库根目录启动，图片样例使用仓库内的相对路径：

```sh
go run ./examples/chat
```

空白页列出六项样例，点击任意一项即可流式展示。也可以在输入框发送下面的关键词。

| 样例名 | 关键词 | 检查方法 |
| --- | --- | --- |
| `selection` | 选择、跨段、拖选 | 跨标题、段落、引用、列表、代码和表格正反向拖选；Cmd/Ctrl+A 全选当前回答，Cmd/Ctrl+C 复制；输出期间向上滚动后选择文字，检查追加和结束后选区是否保留 |
| `math` | 公式、数学、LaTeX | 查看行内与独立公式的分式、根号、下标、求和上下限、嵌套矩阵、跨段宏及自动伸缩括号；代码里的美元符号应保持原样 |
| `code-scroll` | 长代码、长行、横向 | 查看“纯文本”和语言标题、图标悬停提示与复制反馈；用触控板横滑，或在底部滚动条上拖动、点击、滚轮滚到 END 标记，切换自动换行；各块设置独立，复制后长行不应被插入换行 |
| `click-selection` | 双击、三击、选词、选段 | 双击中英文词、标点旁的词及跨粗体边界的 `rendering`；三击自动折行的段落，复制检查边界 |
| `references` | 脚注、引用、链接 | 点击跨块完整、折叠、缩略引用；查看重复引用、多段脚注和返回位置；同块链接作为对照 |
| `images` | 图片 | 查看本地 PNG 的三个色块、缩放比例、带链接的图片、空替代文字和加载失败占位 |
| `table` | 表格 | 查看列对齐、跨单元格选择及复制 |
| `downloader` | 其他消息 | 查看原有的并发下载器回答，包含代码、表格、引用和任务列表 |
| `all` | 全部、TODO、验证示例 | 在一篇回答内查看前六项样例 |

跨段落选择、双击选词、三击选段、边缘拖选自动滚动、常用数学公式、代码块横向滚动和换行切换已实现。脚注跳转与返回、跨块引用、异步图片、数学宏、嵌套矩阵和伸缩括号也已实现，按表中动作检查。

可以直接显示完整样例，从顶部开始浏览，跳过流式等待：

```sh
go run ./examples/chat -sample=all
go run ./examples/chat -sample=math
go run ./examples/chat -sample=images
```

`-sample` 使用表格第一列的样例名。查看快速或慢速的流式过程：

```sh
go run ./examples/chat -delay=0
go run ./examples/chat -delay=80ms
```

“复制全文”复制 Markdown 源文；选中文字后按 Cmd/Ctrl+C 复制纯文本；代码块按钮只复制该块代码。单击链接会调用 macOS 的 `open` 命令。

```sh
go test ./examples/chat -count=1
```

测试检查关键词路由、全部样例的覆盖范围、同块引用链接对照和 PNG 文件有效性。渲染器的选择与复制回归测试在 `ui/markdown/selection_test.go` 和 `ui/markdown/selection_units_test.go`，代码块的换行、横向滚动、滚动后的选择和流式状态测试在 `ui/markdown/code_test.go`；公式解析、基线、行高、流式回退和源码复制测试在 `ui/markdown/math_test.go` 和 `math_extensions_test.go`；脚注与引用测试在 `references_test.go`，图片加载与布局测试在 `images_test.go`、`ui/internal/imageload/decode_test.go`。
