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
| `table.go` | `Table`、`Col`：排序、选中、键盘导航、禁用、只渲染可见行 |
| `select.go` | `Select`：下拉选择、禁用与恢复 |
| `tabs.go` | `Tabs`：页面切换、静默赋值与禁用 |
| `radio.go` | `RadioGroup`：单选、静默赋值、回调与禁用 |
| `switch.go` | `Switch`：尺寸、禁用、加载、键盘操作 |
| `slider.go` | `Slider`：范围、步长、拖动和键盘调整 |
| `accordion.go` | `Accordion`：单项 / 多项展开、统一值接口、键盘导航、整体禁用 |
| `progress.go` | `Progress`：确定进度、不确定动画和模式切换 |
| `dialog.go` | `Dialog`：确认、提示，放在 `window.Options.Overlay` |
| `icon.go` | `Icon`、`VectorIcon`：矢量图标、尺寸与颜色 |
| `kbd.go` | `Kbd` / `KbdView`：平台快捷键、尺寸、无边框样式 |
| `badge.go` | `Badge`：数字、上限、圆点、图标、颜色和固定占位角标 |
| `toggle.go` | `Toggle` / `ToggleView`：尺寸、图标、Ghost、选中、禁用与键盘 |
| `toggle_group.go` | `ToggleGroup`：单选/多选、禁用项与方向键导航 |
| `draw.go` | 内部绘图小工具 |

- **依赖**：`core`、`theme`、`layout`；输入框绘制共用内部的 `ui/internal/editorstyle`。
- **被谁依赖**：应用代码。

新组件就是这个目录下的新文件，步骤见 [扩展指南 · 新增组件](../../docs/extending.md#新增组件)。API 详见 [组件与布局](../../docs/widgets.md)。
