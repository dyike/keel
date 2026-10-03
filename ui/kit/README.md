# ui/kit

Keel 的组件，全部基于 el：展示、浮层、表单、数据、应用外壳和可视化。焦点、禁用、定时和浮层能力由 el 提供。

- 直接依赖：`core`、`theme`、`locale`、`el`。
- 不依赖：`window`。
- 调用者：应用视图和示例。

组件 API、文件模板和验收要求见[kit 组件规范](../../docs/kit.md)，阶段安排见[设计决策](../../docs/decisions.md#新组件基于-eluiwidget-冻结)。

| 组件 | 说明 |
| --- | --- |
| [CodeEditor](../../docs/kit/code_editor.md) | 代码编辑器：高亮、补全、诊断 |
| [Kbd](../../docs/kit/kbd.md) | 快捷键键帽，继承字号和平台格式 |
| [Button](../../docs/kit/button.md) | 操作按钮、图标、键盘与加载状态 |
| [Alert](../../docs/kit/alert.md) | 行内/横幅提示、尺寸、图标与富内容，支持浅深色 |
| [Empty](../../docs/kit/empty.md) | 空状态说明 |
| [Avatar](../../docs/kit/avatar.md) | 固定尺寸头像和姓名回退 |
| [AvatarGroup](../../docs/kit/avatar_group.md) | 叠放头像组、人数上限与溢出标记 |
| [Tag](../../docs/kit/tag.md) | 可选择/移除标签，描边、尺寸、圆角、自定义内容与配色 |
| [DescriptionList](../../docs/kit/description_list.md) | 多列/跨列、横纵标签、自定义值与分隔线 |
| [GroupBox](../../docs/kit/group_box.md) | 标题/描述分组、外观、框外 footer 与样式配置 |
| [StatusBar](../../docs/kit/status_bar.md) | 固定 24dp 左右状态栏 |
| [Marker](../../docs/kit/marker.md) | 纯图形标记 |
| [Icon](../../docs/kit/icon.md) | 矢量图标 |
| [Spinner](../../docs/kit/spinner.md) | 不确定进度、减少动画、自定义图标与颜色 |
| [Skeleton](../../docs/kit/skeleton.md) | 占位、圆形、Shimmer 扫光 |
| [Popover](../../docs/kit/popover.md) | 触发元素旁的非模态面板 |
| [Tooltip](../../docs/kit/tooltip.md) | 富内容提示、动作键位与定位 |
| [HoverCard](../../docs/kit/hover_card.md) | 悬停预览、实例延时与锚点定位 |
| [Menu](../../docs/kit/menu.md) | 命令菜单与子菜单 |
| [DropdownButton](../../docs/kit/dropdown_button.md) | 带菜单的按钮、分体按钮 |
| [Dialog](../../docs/kit/dialog.md) | 模态对话框、确认框 |
| [Sheet](../../docs/kit/sheet.md) | 贴边滑入的模态面板 |
| [Notifier](../../docs/kit/notifier.md) | 右上角通知栈 |
| [CopyButton](../../docs/kit/copy_button.md) | 复制并显示反馈 |
| [Checkbox](../../docs/kit/checkbox.md) | 复选框、半选 |
| [Switch](../../docs/kit/switch.md) | 开关 |
| [RadioGroup](../../docs/kit/radio_group.md) | 单选组 |
| [Toggle](../../docs/kit/toggle.md) | 保持按下的按钮 |
| [ToggleGroup](../../docs/kit/toggle_group.md) | 单选或多选按钮组 |
| [Input / TextArea](../../docs/kit/input.md) | 文本框、前后缀、清空、错误 |
| [Input Group](../../docs/kit/input_group.md) | 输入/多行编辑器与四方向附加内容 |
| [Select](../../docs/kit/select.md) | 下拉选择、可搜索 |
| [Combobox](../../docs/kit/combobox.md) | 可筛选输入 |
| [Slider](../../docs/kit/slider.md) | 线性/对数滑块、范围选择与结束回调 |
| [NumberInput](../../docs/kit/number_input.md) | 数字输入 |
| [OtpInput](../../docs/kit/otp_input.md) | 验证码、密码遮罩、分组与尺寸 |
| [TimeField](../../docs/kit/time_field.md) | 时间输入 |
| [Calendar](../../docs/kit/calendar.md) | 日历、范围 |
| [DatePicker](../../docs/kit/date_picker.md) | 日期字段 |
| [Rating](../../docs/kit/rating.md) | 星级评分 |
| [Stepper](../../docs/kit/stepper.md) | 横纵步骤进度、图标、尺寸、单步禁用 |
| [Form](../../docs/kit/form.md) | 表单与校验 |
| [VirtualList](../../docs/kit/virtual_list.md) | 等高虚拟列表 |
| [VariableList](../../docs/kit/variable_list.md) | 自然高度虚拟列表、稳定 key 与阅读位置保持 |
| [List](../../docs/kit/list.md) | 单选长列表 |
| [Tree](../../docs/kit/tree.md) | 树 |
| [Table](../../docs/kit/table.md) | 表格：排序、列宽、单元格插槽 |
| [Pagination](../../docs/kit/pagination.md) | 分页 |
| [Command](../../docs/kit/command.md) | 命令面板 |
| [Message](../../docs/kit/message.md) | 对话消息 |
| [Bubble](../../docs/kit/bubble.md) | 聊天气泡 |
| [MessageScroller](../../docs/kit/message_scroller.md) | 对话滚动区 |
| [Attachment](../../docs/kit/attachment.md) | 附件卡片 |
| [Tabs](../../docs/kit/tabs.md) | 四种外观、图标/富标签、单项禁用、宽度上限、滚动/溢出与拖动重排 |
| [Collapsible](../../docs/kit/collapsible.md) | 独立触发器与内容、可中断展开动画 |
| [Accordion](../../docs/kit/accordion.md) | 折叠面板、四档尺寸、边框开关 |
| [Badge](../../docs/kit/badge.md) | 数字/圆点/图标角标、尺寸与自定义颜色 |
| [ProgressCircle](../../docs/kit/progress_circle.md) | 圆形进度、中心内容与不确定状态 |
| [Progress](../../docs/kit/progress.md) | 进度条 |
| [Link](../../docs/kit/link.md) | 链接 |
| [Image](../../docs/kit/image.md) | 图片 |
| [Sidebar](../../docs/kit/sidebar.md) | 导航侧栏 |
| [Toolbar](../../docs/kit/toolbar.md) | 工具栏与溢出菜单 |
| [Resizable](../../docs/kit/resizable.md) | 可拖动分隔的两栏 |
| [Dock](../../docs/kit/dock.md) | 可停靠面板与布局保存 |
| [Settings](../../docs/kit/settings.md) | 设置页 |
| [Carousel](../../docs/kit/carousel.md) | 轮播 |
| [TitleBar](../../docs/kit/title_bar.md) | 无边框窗口的标题栏 |
| [Chart](../../docs/kit/chart.md) | 折线图、面积图、柱状图、堆叠柱状图 |
| [CandlestickChart](../../docs/kit/candlestick_chart.md) | 开高低收、密集聚合、数据表 |
| [PieChart](../../docs/kit/pie_chart.md) | 饼图、环形图、交互图例 |
| [Plot](../../docs/kit/plot.md) | 可缩放平移的 x/y 绘图 |
| [ColorPicker](../../docs/kit/color_picker.md) | 颜色选择 |
| [Questionnaire](../../docs/kit/questionnaire.md) | 分页问卷与答案模型 |
