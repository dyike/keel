# 组件参考

[English](kit.md) | 简体中文

`ui/kit` 提供现成的 el 视图。按用途从下面的分类找组件，每篇文档包含用法、API 和交互说明；在线文档中的示例可以直接操作。

## 基础展示

| 组件 | 用途 |
| --- | --- |
| [Label](kit/label.zh-CN.md) | 可换行标签、次级文案与关键词高亮 |
| [Icon](kit/icon.zh-CN.md) | 矢量图标 |
| [Image](kit/image.zh-CN.md) | 图片：异步加载、SVG、GIF 动图、内存与磁盘缓存、预览 |
| [Avatar](kit/avatar.zh-CN.md) | 固定尺寸头像和姓名回退 |
| [AvatarGroup](kit/avatar_group.zh-CN.md) | 叠放头像组、人数上限与溢出标记 |
| [Badge](kit/badge.zh-CN.md) | 数字/圆点/图标角标、尺寸与自定义颜色 |
| [Tag](kit/tag.zh-CN.md) | 可选择/移除标签，描边、尺寸、圆角、自定义内容与配色 |
| [Marker](kit/marker.zh-CN.md) | 纯图形标记 |
| [StatusMarker](kit/status_marker.zh-CN.md) | 消息状态、时间线边界和系统提示 |
| [Kbd](kit/kbd.zh-CN.md) | 快捷键键帽、平台格式、独立/继承字号与自定义样式 |
| [DescriptionList](kit/description_list.zh-CN.md) | 多列/跨列、横纵标签、自定义值与分隔线 |
| [GroupBox](kit/group_box.zh-CN.md) | 标题/描述分组、外观、框外 footer 与样式配置 |

## 操作与导航

| 组件 | 用途 |
| --- | --- |
| [Button](kit/button.zh-CN.md) | 操作按钮、图标、键盘与加载状态、选中状态 |
| [ButtonGroup](kit/button_group.zh-CN.md) | 几个按钮连成一体，横向或竖向 |
| [Link](kit/link.zh-CN.md) | 链接 |
| [CopyButton](kit/copy_button.zh-CN.md) | 复制并显示反馈 |
| [Tabs](kit/tabs.zh-CN.md) | 四种外观、图标/富标签、单项禁用、宽度上限、滚动/溢出与拖动重排 |
| [Accordion](kit/accordion.zh-CN.md) | 折叠面板、四档尺寸、边框开关 |
| [Collapsible](kit/collapsible.zh-CN.md) | 独立触发器与内容、可中断展开动画 |
| [Pagination](kit/pagination.zh-CN.md) | 分页 |
| [Stepper](kit/stepper.zh-CN.md) | 横纵步骤进度、图标、尺寸、单步禁用 |
| [Command](kit/command.zh-CN.md) | 命令面板 |

## 输入与选择

| 组件 | 用途 |
| --- | --- |
| [Input / TextArea](kit/input.zh-CN.md) | 文本框、前后缀、清空、错误 |
| [Input Group](kit/input_group.zh-CN.md) | 输入/多行编辑器与四方向附加内容 |
| [Checkbox](kit/checkbox.zh-CN.md) | 复选框、半选 |
| [Switch](kit/switch.zh-CN.md) | 开关 |
| [Radio](kit/radio.zh-CN.md) | 可以放在任意位置的单个单选按钮 |
| [RadioGroup](kit/radio_group.zh-CN.md) | 单选组 |
| [Toggle](kit/toggle.zh-CN.md) | 保持按下的按钮 |
| [ToggleGroup](kit/toggle_group.zh-CN.md) | 单选或多选按钮组 |
| [Select](kit/select.zh-CN.md) | 下拉选择、可搜索 |
| [Combobox](kit/combobox.zh-CN.md) | 可筛选输入 |
| [NumberInput](kit/number_input.zh-CN.md) | 数字输入 |
| [OtpInput](kit/otp_input.zh-CN.md) | 验证码、密码遮罩、分组与尺寸 |
| [TimeField](kit/time_field.zh-CN.md) | 时间输入 |
| [Calendar](kit/calendar.zh-CN.md) | 日历、范围 |
| [DatePicker](kit/date_picker.zh-CN.md) | 日期字段 |
| [Slider](kit/slider.zh-CN.md) | 线性/对数滑块、范围选择与结束回调 |
| [Rating](kit/rating.zh-CN.md) | 星级评分、已填星减分、自定义尺寸与颜色 |
| [ColorPicker](kit/color_picker.zh-CN.md) | 颜色选择 |
| [Form](kit/form.zh-CN.md) | 表单与校验 |
| [Questionnaire](kit/questionnaire.zh-CN.md) | 分页问卷与答案模型 |

## 浮层与反馈

| 组件 | 用途 |
| --- | --- |
| [Popover](kit/popover.zh-CN.md) | 触发元素旁的非模态面板 |
| [Tooltip](kit/tooltip.zh-CN.md) | 富内容提示、动作键位与定位 |
| [HoverCard](kit/hover_card.zh-CN.md) | 悬停预览、实例延时与锚点定位 |
| [Menu](kit/menu.zh-CN.md) | 命令菜单与子菜单 |
| [DropdownButton](kit/dropdown_button.zh-CN.md) | 带菜单的按钮、分体按钮 |
| [Dialog](kit/dialog.zh-CN.md) | 模态对话框、确认框；`Show(cx)` 不用放进视图树 |
| [Sheet](kit/sheet.zh-CN.md) | 贴边滑入的模态面板，可拖动调整尺寸；`Show(cx)` 不用放进视图树 |
| [Notifier](kit/notifier.zh-CN.md) | 八个方位的通知栈、富内容与操作、系统通知投递；`WindowNotifier(cx)` 是窗口自带的一个 |
| [Alert](kit/alert.zh-CN.md) | 行内/横幅提示、尺寸、图标与富内容，支持浅深色 |
| [Empty](kit/empty.zh-CN.md) | 空状态富内容、媒体、操作/尾部与分区样式 |
| [Spinner](kit/spinner.zh-CN.md) | 不确定进度、减少动画、自定义图标/颜色/周期 |
| [Skeleton](kit/skeleton.zh-CN.md) | 占位、圆形、自定义圆角、次级色阶与 Shimmer 扫光 |
| [Progress](kit/progress.zh-CN.md) | 进度条、自定义高度/颜色/圆角与轨道样式 |
| [ProgressCircle](kit/progress_circle.zh-CN.md) | 圆形进度、中心内容与不确定状态 |
| [ShimmerText](kit/shimmer_text.zh-CN.md) | 文字扫光 |

## 列表与表格

| 组件 | 用途 |
| --- | --- |
| [List](kit/list.zh-CN.md) | 长列表：单选/多选、搜索、分组、加载更多、状态插槽、拖动重排 |
| [VirtualList](kit/virtual_list.zh-CN.md) | 等高虚拟列表 |
| [VariableList](kit/variable_list.zh-CN.md) | 自然高度虚拟列表、稳定 key 与阅读位置保持 |
| [Tree](kit/tree.zh-CN.md) | 树：懒加载、多选、拖动重排（自动展开与滚动） |
| [Table](kit/table.zh-CN.md) | 表格：排序、列宽、单元格插槽 |

## 聊天与内容

| 组件 | 用途 |
| --- | --- |
| [Message](kit/message.zh-CN.md) | 对话消息 |
| [MessageContent](kit/message_content.zh-CN.md) | 一条消息里混排多个气泡、附件和视图 |
| [MessageGroup](kit/message_group.zh-CN.md) | 连续消息分组 |
| [Bubble](kit/bubble.zh-CN.md) | 聊天气泡 |
| [BubbleGroup](kit/bubble_group.zh-CN.md) | 连续气泡分组 |
| [MessageScroller](kit/message_scroller.zh-CN.md) | 对话滚动区 |
| [Attachment](kit/attachment.zh-CN.md) | 附件卡片 |
| [AttachmentGroup](kit/attachment_group.zh-CN.md) | 横向排列、可滚动的附件组 |
| [CodeEditor](kit/code_editor.zh-CN.md) | 代码编辑器：高亮、补全、诊断、多光标、折叠、软换行、查找替换 |

## 图表与绘图

| 组件 | 用途 |
| --- | --- |
| [Chart](kit/chart.zh-CN.md) | 折线图、面积图、柱状图、堆叠柱状图 |
| [PieChart](kit/pie_chart.zh-CN.md) | 饼图、环形图、交互图例 |
| [CandlestickChart](kit/candlestick_chart.zh-CN.md) | 开高低收、密集聚合、数据表 |
| [RadarChart](kit/radar_chart.zh-CN.md) | 雷达图 |
| [SankeyChart](kit/sankey_chart.zh-CN.md) | 桑基图 |
| [Plot](kit/plot.zh-CN.md) | 可缩放平移的 x/y 绘图 |
| [公共绘图基础件](kit/plot_primitives.zh-CN.md) | `ui/plot`：自定义图表的比例尺与即时绘制 |

## 应用布局

| 组件 | 用途 |
| --- | --- |
| [TitleBar](kit/title_bar.zh-CN.md) | 无边框窗口的标题栏 |
| [Sidebar](kit/sidebar.zh-CN.md) | 导航侧栏 |
| [Toolbar](kit/toolbar.zh-CN.md) | 工具栏与溢出菜单 |
| [StatusBar](kit/status_bar.zh-CN.md) | 固定 24dp 左右状态栏 |
| [Resizable](kit/resizable.zh-CN.md) | 可拖动分隔的两栏 |
| [ResizableGroup](kit/resizable_group.zh-CN.md) | 多面板分割，横纵和嵌套 |
| [Dock](kit/dock.zh-CN.md) | 可停靠面板、拆分、最大化、收起侧栏、跨窗口与布局保存 |
| [Settings](kit/settings.zh-CN.md) | 设置页 |
| [Carousel](kit/carousel.zh-CN.md) | 轮播：多项视口、拖动、触控板与滚轮、循环轨道 |

## 公共用法

组件实例保存状态，应在构造视图时创建一次，保存在字段里，再在 `Render(cx)` 中使用。`SetValue` 是程序赋值，不触发 `OnChange`；用户操作才触发回调。后台更新遵守 [线程规则](architecture.zh-CN.md#线程规则)。

浮层组件需要放在 `el.Root` 中。布局和组合见 [元素与视图](el.zh-CN.md)，颜色和尺寸见 [主题](theme.zh-CN.md)，新增框架组件见 [组件开发规范](component-development.zh-CN.md)。

## 可选功能与体积

代码高亮和网络图片按需引入：

| 功能 | 引入 | 不引入时 |
| --- | --- | --- |
| 代码高亮：CodeEditor、TextView、Markdown 代码块 | `_ "github.com/dyike/keel/ui/highlight"` | 显示纯文本 |
| 网络图片：Image、Avatar、Attachment、Markdown 的 http(s) 地址 | `_ "github.com/dyike/keel/ui/netimage"` | 本地文件和 data URL 照常，网络地址返回 `core.ErrNoImageFetcher` |

两项各增加约 4 MB。`keel build` 默认去掉调试信息，只使用 kit 的应用约 10 MB。
