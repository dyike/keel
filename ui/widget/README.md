# ui/widget

交互组件，一个组件一个文件：

| 文件 | 组件 |
| --- | --- |
| `text.go` | `Text`、`Heading`、`Muted` |
| `button.go` | `Button` |
| `link.go` | `Link` |
| `input.go` | `Input`、`TextArea` |
| `checkbox.go` | `Checkbox` |
| `table.go` | `Table`、`Col`：排序、选中、键盘导航、只渲染可见行 |
| `select.go` | `Select`：下拉选择 |
| `tabs.go` | `Tabs` |
| `radio.go` | `RadioGroup` |
| `switch.go` | `Switch` |
| `progress.go` | `Progress` |
| `dialog.go` | `Dialog`：确认、提示，放在 `window.Options.Overlay` |
| `draw.go` | 内部绘图小工具 |

- **依赖**：`core`、`theme`、`layout`。
- **被谁依赖**：应用代码。

新组件就是这个目录下的新文件，步骤见 [扩展指南 · 新增组件](../../docs/extending.md#新增组件)。API 详见 [组件与布局](../../docs/widgets.md)。
