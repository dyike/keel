# ui/el

GPUI 风格的元素与视图：视图是普通 struct，每帧 `Render` 返回一棵链式样式搭起来的元素树；flexbox 布局；元素状态（悬停、滚动、输入框内容）按元素路径或 ID 自动保存；Agent 语义自动生成。

- **依赖**：`core`、`theme`，以及内部的 `ui/internal/loop` 和 `ui/internal/editorstyle`。不依赖 `kit`、`window`。
- **被谁依赖**：应用代码。`window` 通过 `FillsWindow` 接口认出 `el.Root`，不引用本包。

| 文件 | 内容 |
| --- | --- |
| `element.go` | `Element`、`Node`、`Styled[T]` 的全部链式方法，`Div`、`Text`、`Widget`、`Map`；`Decorate` 包裹绘制，`VisitWidgets` 读取布局后的部件坐标 |
| `time.go` | 帧时间、声明式定时器、显式 key 与减少动画 |
| `overlay.go` | 声明式浮层、锚定定位、模态输入隔离、焦点约束与悬停查询 |
| `focus.go` | 原生焦点顺序、程序焦点、按键冒泡与默认激活 |
| `input.go` | `Input`、`TextArea` |
| `style.go` | `Style`、长度（`Dp`、`Frac`、`Full`）、对齐常量 |
| `layout.go`、`flow.go` | flexbox、换行与简单网格布局 |
| `paint.go` | 绘制、点击区域、滚动、输入框、语义信息 |
| `viewport.go` | 绘制坐标、可视区域与最近滚动容器的程序滚动 |
| `state.go` | 元素状态存储与回收 |
| `root.go` | `View`、`ViewFunc`、`Context`、`Root`、`Embed`，每帧的执行顺序 |

使用指南：[元素与视图](../../docs/el.md)。

`input_paste.go` 在默认文本插入前处理 OnPaste；可接入 core.ClipboardReader，原生读取失败回退 Gio 纯文本通路。异步完成通过 core.Update，编辑内容或选区变化后丢弃旧结果。

原子输入引用由 `el.InputDocument` 接入 `ui/internal/inputcontent`，后者只保存文本、引用范围、选区和编辑事务，不依赖 Gio 或其他 Keel 模块。kit 和 markdown 仅经 el 间接依赖它。
