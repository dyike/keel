# kit 组件规范

组件放在 `ui/kit`，用 `ui/el` 组织元素、布局和交互。演进过程见[设计决策](decisions.md#新组件基于-eluiwidget-冻结)。

## 模块与 API

kit 的枚举常量一律以类型名作前缀，去掉类型名中的 Name / Shape / Status / Variant 后缀：ToneNeutral / ToneInfo / ToneSuccess / ToneWarning / ToneDanger，AvatarOnline / AvatarBusy / AvatarOffline，ButtonPrimary / ButtonSecondary / ButtonGhost / ButtonDanger；IconCheck、MarkerDot 沿用现有命名。不保留旧名称的兼容别名。el 是底层布局库，el.Bottom、el.Start 等布局短名不受这条规则限制。

kit 只直接依赖 Keel 的 `core`、`theme`、`locale`、`el`；不引用 `window`。el 缺少的基础能力先在 el 中实现，不在各组件里复制 Gio 输入路由、定时或浮层机制。依赖测试按传递依赖登记 `internal/loop` 和 `internal/editorstyle`，它们不是 kit 的直接依赖。

构造函数 `Xxx(...)` 返回 `*XxxView`。组件以 `Render(*el.Context) el.Element` 接入 el，实例保留业务状态，Render 根据当前状态生成元素树。动态列表使用稳定 ID，不用数组位置代表可移动项目。

有值的交互组件提供 `Value()`、`SetValue(...)`、链式 `OnChange(...)`、`SetDisabled(bool)`。程序赋值不触发回调，只有用户操作触发。按钮等动作组件使用点击回调和 `SetDisabled`，纯展示组件不强加值与回调接口。返回切片等可变值时不能暴露内部存储。

## 状态与布局

加载、忙碌状态不改变焦点和 Tab 顺序，只忽略激活。

回调在帧锁内改状态；后台任务通过 `core.Update` 返回 UI。组件不能根据 `gtx.Enabled()` 为 false 重置跨帧状态：它也可能表示 el 测量或离屏布局没有输入源。关闭菜单、停止拖动等重置放在 `SetDisabled(true)` 中；事件、外观与语义遵守父级禁用状态。

装饰的显示、隐藏和内容变化不应改变宿主尺寸或基线，例如 Badge 计数变化不能让按钮跳动。普通内容变化可以重新排版。窄容器中遵守约束；文本检查中文、拉丁字母、数字、混排、1×/2×缩放，输入光标和选择区使用统一字形度量，不按某个截图硬补偏移。

框架自己的文字（按钮文案、无障碍名称、占位文字、计数）一律在 Render 时从 `locale.Current()` 读取，不在构造时保存，也不写死中文；`internal/deps` 的测试会拦下写死的中文字符串。拼接"动作 + 对象"形式的名称时用 `locale.Current().Name(action, target)`。应用传进来的文字（标题、菜单项）原样使用。

颜色在 Render 时读取 theme 语义色，不在构造函数中保存主题快照。自定义固定色属于显式覆盖；主题切换不会替应用推断其含义。缓存必须包含主题版本，或使用支持主题失效的 `cx.Cache`。M0 不提供局部主题作用域。

交互组件支持鼠标和键盘；禁用时不能激活或获得焦点，Agent 快照报告 disabled。焦点、按键、禁用、定时、浮层的具体新 API 在对应阶段 review，不在本规范预先定型。

浮层组件必须先创建面板并调用 `cx.Overlay`，再渲染 Body / Footer 的内容并追加到面板。内容本身可能登记子浮层；顺序反过来会让子菜单早于父层登记，因锚点尚不可用而被关闭。模态组件用 `Layer.Owner` 绑定所属元素，继承外层的禁用和隐藏。验收必须包含真实嵌套打开、逐层 Esc、焦点返回和祖先禁用，不能只测试独立弹层。

## 文件模板与验收

一个组件对应 `ui/kit/<name>.go`、`<name>_test.go`、`docs/kit/<name>.md`、`examples/components/<name>.go`。在 kit README 和文档索引登记。示例注册独立 `-section <name>`，展示常用状态、边界和浅深色。

组件文档包含用途、最小用法、公开 API、键盘操作（适用时）、语义、边界和验证入口；测试验证用户可观察行为，不复制实现算法。

`ui/kit/conventions_test.go` 自动检查其中可以机器判断的部分：每个组件都有文档、同名示例 section 和 `ui/window` 中的 Agent 测试；枚举常量带类型名前缀；kit 和 el 的公开 API 没有兼容入口或别名。新增的 Agent 角色不需要在自动化代码里登记，只有需要单独列出子元素的容器角色才加入 `containerRoles`。

提交前逐项检查：

- `uitest` 驱动布局与交互；纯展示组件检查尺寸、约束、状态和颜色。
- `ui/window` Agent 快照覆盖名称、角色、值和状态；新增 role 同步 `automation.go` 与 `docs/automation.md`。
- 有交互时覆盖键盘、禁用、恢复、程序赋值不触发回调；嵌入 el 后连续多帧不丢状态。
- 示例可运行，浅深色和文字布局经过截图检查。
- 公开 API、组件文档、README、示例同时更新。
- `go build ./... && go vet ./ui/... && go test ./... -count=1` 全部通过，包括 `cmd/keel-mcp` 端到端测试。

每个组件完成后单独提交，再开始下一个。

## 已实现组件

- [Kbd](kit/kbd.md)：继承字号的快捷键键帽，支持平台格式和 Plain。

- [Button](kit/button.md)：操作按钮，支持焦点、禁用、图标和固定尺寸的加载状态。

- [Alert](kit/alert.md)：行内状态提示。
- [Empty](kit/empty.md)：空状态说明。
- [Avatar](kit/avatar.md)：图片与姓名回退头像。
- [Tag](kit/tag.md)：可选择、可移除标签。
- [DescriptionList](kit/description_list.md)：字段说明列表。
- [GroupBox](kit/group_box.md)：带标题的视图分组。
- [StatusBar](kit/status_bar.md)：状态与详情栏。
- [Marker](kit/marker.md)：纯图形标记。

[Icon](kit/icon.md)：矢量图标，默认颜色随主题切换。

主题文本使用场景：选中底色 Highlight 上使用 PrimaryText；Primary 保留为按钮背景和描边。Markdown 默认使用 CodeBg/CodeText，旧包级颜色变量的零值表示跟随主题，CodeStyle 为空时按背景自动选择 github/github-dark。

[Spinner](kit/spinner.md)：支持减少动画的不确定进度。

[Skeleton](kit/skeleton.md)：占位、圆形和 Shimmer 扫光。

表单控件：

- [Checkbox](kit/checkbox.md)、[Switch](kit/switch.md)、[RadioGroup](kit/radio_group.md)、[Toggle](kit/toggle.md)、[ToggleGroup](kit/toggle_group.md)：选择与开关。
- [Input / TextArea](kit/input.md)：文本框，支持前后缀、清空、错误提示。
- [Input Group](kit/input_group.md)：统一标签和边框的输入、图标与按钮组合。
- [Select](kit/select.md)、[Combobox](kit/combobox.md)：下拉选择、可筛选输入。
- [NumberInput](kit/number_input.md)、[OtpInput](kit/otp_input.md)、[TimeField](kit/time_field.md)：数字、验证码、时间。
- [Calendar](kit/calendar.md)、[DatePicker](kit/date_picker.md)：日期与日期范围。
- [Slider](kit/slider.md)、[Rating](kit/rating.md)、[Stepper](kit/stepper.md)：数值、评分、步骤。
- [Form](kit/form.md)：两列表单与统一校验。

数据与长内容：

- [VirtualList](kit/virtual_list.md)、[List](kit/list.md)、[Tree](kit/tree.md)：只构建可见行的列表和树。
- [Table](kit/table.md)、[Pagination](kit/pagination.md)：表格与分页。
- [Command](kit/command.md)：命令面板。
- [Message](kit/message.md)、[Bubble](kit/bubble.md)、[MessageScroller](kit/message_scroller.md)、[Attachment](kit/attachment.md)：对话界面。

基础组件：[Tabs](kit/tabs.md)、[Accordion](kit/accordion.md)、[Collapsible](kit/collapsible.md)、[Badge](kit/badge.md)、[Progress](kit/progress.md)、[Link](kit/link.md)、[Image](kit/image.md)。

应用外壳：[TitleBar](kit/title_bar.md)、[Sidebar](kit/sidebar.md)、[Toolbar](kit/toolbar.md)、[Resizable](kit/resizable.md)、[Dock](kit/dock.md)、[Settings](kit/settings.md)、[Carousel](kit/carousel.md)。

可视化与专项：[Chart](kit/chart.md)、[Plot](kit/plot.md)、[ColorPicker](kit/color_picker.md)、[Questionnaire](kit/questionnaire.md)。

浮层组件（需要 `el.Root`）：

- [Popover](kit/popover.md)：触发元素旁的非模态面板。
- [Tooltip](kit/tooltip.md)：悬停或键盘聚焦时的简短提示。
- [HoverCard](kit/hover_card.md)：悬停预览卡片。
- [Menu](kit/menu.md)：命令菜单，支持子菜单和键盘导航。
- [DropdownButton](kit/dropdown_button.md)：带菜单的按钮和分体按钮。
- [Dialog](kit/dialog.md)：模态对话框和标准确认框。
- [Sheet](kit/sheet.md)：贴边滑入的模态面板。
- [Notifier](kit/notifier.md)：右上角通知栈。

其他：[CopyButton](kit/copy_button.md)，复制到剪贴板并显示反馈。

DangerText 用于 Alert/Tag 等表面上的危险状态文字；Danger 仍用于实心危险按钮。两套默认配色的状态文字对 Surface 均以 4.5:1 为最低对比度验收。
