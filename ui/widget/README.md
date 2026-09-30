# ui/widget

交互组件，一个组件一个文件：

| 文件 | 组件 |
| --- | --- |
| `text.go` | `Text`、`Heading`、`Muted` |
| `button.go` | `Button` |
| `link.go` | `Link` |
| `input.go` | `Input`、`TextArea` |
| `checkbox.go` | `Checkbox` |

- **依赖**：`core`、`theme`、`layout`。
- **被谁依赖**：应用代码。

新组件就是这个目录下的新文件，步骤见 [扩展指南 · 新增组件](../../docs/extending.md#新增组件)。API 详见 [组件与布局](../../docs/widgets.md)。
