# 组件能力差距盘点

[English](gallery-gaps.md) | [简体中文](gallery-gaps.zh-CN.md)

对照 [NSH Gallery 源码](https://github.com/longbridge/nsh/blob/main/examples/gallery/App.tsx) 的 41 个 base 组件演示（本地获取于 2026-11-14）。下表记录对应入口，不代表行为、性能或移动体验等价。

| 参考组件 | Keel 示例入口 | 状态 |
| --- | --- | --- |
| Accordion | [accordion](../examples/components/accordion.go) | 已有入口，细项待验收 |
| Avatar | [avatar](../examples/components/avatar.go) | 已有入口，细项待验收 |
| Button | [button](../examples/components/button.go) | 已有入口，细项待验收 |
| Carousel | [carousel](../examples/components/carousel.go) | 已有入口，细项待验收 |
| Checkbox | [checkbox](../examples/components/checkbox.go) | 已有入口，细项待验收 |
| Collapsible | [collapsible](../examples/components/collapsible.go) | 已有入口，细项待验收 |
| ColorPicker | [color_picker](../examples/components/color_picker.go) | 已有入口，细项待验收 |
| Combobox | [combobox](../examples/components/combobox.go) | 已有入口，细项待验收 |
| Command | [command](../examples/components/command.go) | 已有入口，细项待验收 |
| DataTable | [table](../examples/components/table.go) | 已有入口，细项待验收 |
| DatePicker | [date_picker](../examples/components/date_picker.go) | 已有入口，细项待验收 |
| DiffView | `kit.Diff` | 已有统一／并排审阅与补丁；已有同步滚动和字素级行内高亮；单文件补丁输入已实现；并排上下文折叠已实现；行号回调已实现；已实现嵌入式差异块自定义操作；完整设备／视觉一致性未验收 |
| Dock | [dock](../examples/components/dock.go) | 已有入口，细项待验收 |
| Editor | [code_editor](../examples/components/code_editor.go) | 已有入口，细项待验收 |
| EmojiPicker | [emoji_picker](../examples/components/emoji_picker.go) | 已有 Unicode 目录、搜索、肤色、最近使用、虚拟网格及自适应弹层／底部面板；真机／字体一致性未验收 |
| HoverCard | [hover_card](../examples/components/hover_card.go) | 已有入口，细项待验收 |
| Input | [input](../examples/components/input.go) | 已有入口，细项待验收 |
| Link | [link](../examples/components/link.go) | 已有入口，细项待验收 |
| Menu | [menu](../examples/components/menu.go) | 已有入口，细项待验收 |
| NavStack | [nav_stack](../examples/components/nav_stack.go) | 已实现导航栈与输入／滚动保留；模态／滑动返回仍待补齐 |
| NumberInput | [number_input](../examples/components/number_input.go) | 已有入口，细项待验收 |
| OtpInput | [otp_input](../examples/components/otp_input.go) | 已有入口，细项待验收 |
| Pagination | [pagination](../examples/components/pagination.go) | 已有入口，细项待验收 |
| Plot | [plot](../examples/components/plot.go) | 已有入口，细项待验收 |
| Popover | [popover](../examples/components/popover.go) | 已有入口，细项待验收 |
| Progress | [progress](../examples/components/progress.go) | 已有入口，细项待验收 |
| Radio | [radio_group](../examples/components/radio_group.go) | 已有入口，细项待验收 |
| Resizable | [resizable_group](../examples/components/resizable_group.go) | 已有入口，细项待验收 |
| RichEditor | — | 缺少对应组件 |
| Select | [select](../examples/components/select.go) | 已有入口，细项待验收 |
| Sheet | [sheet](../examples/components/sheet.go) | 已有入口，细项待验收 |
| Slider | [slider](../examples/components/slider.go) | 已有入口，细项待验收 |
| Switch | [switch](../examples/components/switch.go) | 已有入口，细项待验收 |
| Tabs | [tabs](../examples/components/tabs.go) | 已有入口，细项待验收 |
| Terminal | [terminal](../examples/components/terminal.go) | 已有 ANSI 显示、应用接管输入输出、选择复制与滚动；完整 VT 兼容、真实 PTY／输入法和设备验收仍未完成 |
| TextView | [text_view](../examples/components/text_view.go) | 已有入口，细项待验收 |
| Toast | [notifier](../examples/components/notifier.go) | 已有入口，细项待验收 |
| Toggle | [toggle](../examples/components/toggle.go) | 已有入口，细项待验收 |
| Toolbar | [toolbar](../examples/components/toolbar.go) | 已有入口，细项待验收 |
| Tooltip | [tooltip](../examples/components/tooltip.go) | 已有入口，细项待验收 |
| Tree | [tree](../examples/components/tree.go) | 已有入口，细项待验收 |

NavStack 已提供页面状态保留与返回行为，模态和滑动返回仍待补齐。Diff 已提供审阅动作和补丁导出，同步滚动及字素级行内高亮已实现，单文件补丁输入已实现；并排上下文折叠已实现；行号回调已实现；已实现嵌入式差异块自定义操作；完整设备／视觉一致性未验收。EmojiPicker 已有 Unicode 目录、搜索、肤色、最近使用及自适应展示，真机／字体一致性仍待验收；RichEditor 仍缺少对应组件。Terminal 已有 ANSI 显示与交互测试，尚未建立完整 VT 兼容及真实 PTY／输入法／设备验收结论。新增组件必须包含双语文档、组件库入口、交互测试和窗口快照测试。

原语差距另行跟踪：滚动分页／缩放、列表吸顶、Canvas／Shader 扩展、Video。参考库的 WebView 依赖不是本项目纯 Go 绘制架构中的组件替代方案；不能通过引入 HTML/WebView 来宣称补齐。

Diff 范围已对照[参考 DiffDemo](https://github.com/longbridge/nsh/blob/main/examples/gallery/demos/DiffDemo.tsx)核实。演示包含补丁输入、上下文展开／折叠、行统计及差异块／行号动作，没有展示可编辑合并面板；此前工作记录把可编辑合并作为完成门槛扩大了范围。Keel 已有行统计和统一视图上下文展开／折叠，上表剩余项按实际参考列出。
