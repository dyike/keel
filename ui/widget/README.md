# ui/widget

交互组件，一个组件一个文件：

| 文件 | 组件 |
| --- | --- |
| `text.go` | `Text`、`Heading`、`Muted`：换行、文字对齐、`For` 关联聚焦 |
| `button.go`、`style.go` | `Button`：尺寸、图标、加载、禁用；共享尺寸和加载绘制 |
| `link.go` | `Link` |
| `image.go`、`image_load.go` | `Image`、`LoadImage`、`ImageData`：异步加载、缩放和失败占位 |
| `input.go` | `Input`、`TextArea`：前后缀、清空、禁用、标签聚焦、自动高度 |
| `checkbox.go` | `Checkbox`：禁用、半选、回调和键盘操作 |
| `table.go` | `Table`、`Col`：排序、选中、键盘导航、只渲染可见行 |
| `select.go` | `Select`：下拉选择 |
| `tabs.go` | `Tabs` |
| `radio.go` | `RadioGroup` |
| `switch.go` | `Switch`：尺寸、禁用、加载、键盘操作 |
| `slider.go` | `Slider`：范围、步长、拖动和键盘调整 |
| `accordion.go` | `Accordion`：单项 / 多项展开、标题键盘导航 |
| `progress.go` | `Progress`：确定进度、不确定动画和模式切换 |
| `dialog.go` | `Dialog`：确认、提示，放在 `window.Options.Overlay` |
| `icon.go` | `Icon`、`VectorIcon`：矢量图标、尺寸与颜色 |
| `kbd.go` | `Kbd`：平台快捷键、尺寸、无边框样式 |
| `badge.go` | `Badge`：数字、上限、圆点、图标、颜色和角标 |
| `draw.go` | 内部绘图小工具 |

- **依赖**：`core`、`theme`、`layout`。
- **被谁依赖**：应用代码。

新组件就是这个目录下的新文件，步骤见 [扩展指南 · 新增组件](../../docs/extending.md#新增组件)。API 详见 [组件与布局](../../docs/widgets.md)。
