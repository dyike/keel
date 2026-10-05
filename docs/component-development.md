# 组件开发规范

新增或修改 `ui/kit` 组件时遵守这份规范。实现骨架见 [扩展指南](extending.md#新增组件)，外观约定见 [组件视觉规范](visual-guidelines.md)，用户用法和分类入口见 [组件参考](kit.md)。

## 模块与 API

kit 的枚举常量一律以类型名作前缀，去掉类型名中的 Name / Shape / Status / Variant 后缀：ToneNeutral / ToneInfo / ToneSuccess / ToneWarning / ToneDanger，AvatarOnline / AvatarBusy / AvatarOffline，ButtonPrimary / ButtonSecondary / ButtonGhost / ButtonDanger；IconCheck、MarkerDot 沿用现有命名。不保留旧名称的兼容别名。el 是底层布局库，el.Bottom、el.Start 等布局短名不受这条规则限制。

kit 只直接依赖 Keel 的 `core`、`theme`、`locale`、`el`、`base`；不引用 `window`。el 缺少的基础能力先在 el 中实现，不在各组件里复制 Gio 输入路由、定时或浮层机制。依赖测试按传递依赖登记 `internal/loop`、`internal/editorstyle` 和 `internal/inputcontent`，它们不是 kit 的直接依赖。

构造函数 `Xxx(...)` 返回 `*XxxView`。组件以 `Render(*el.Context) el.Element` 接入 el，实例保留业务状态，Render 根据当前状态生成元素树。动态列表使用稳定 ID，不用数组位置代表可移动项目。

有值的交互组件提供 `Value()`、`SetValue(...)`、链式 `OnChange(...)`、`SetDisabled(bool)`。程序赋值不触发回调，只有用户操作触发。按钮等动作组件使用点击回调和 `SetDisabled`，纯展示组件不强加值与回调接口。返回切片等可变值时不能暴露内部存储。

## 状态与布局

加载、忙碌状态不改变焦点和 Tab 顺序，只忽略激活。

回调在帧锁内改状态；后台任务通过 `core.Update` 返回 UI。组件不能根据 `gtx.Enabled()` 为 false 重置跨帧状态：它也可能表示 el 测量或离屏布局没有输入源。关闭菜单、停止拖动等重置放在 `SetDisabled(true)` 中；事件、外观与语义遵守父级禁用状态。

装饰的显示、隐藏和内容变化不应改变宿主尺寸或基线，例如 Badge 计数变化不能让按钮跳动。普通内容变化可以重新排版。窄容器中遵守约束；文本检查中文、拉丁字母、数字、混排、1×/2×缩放，输入光标和选择区使用统一字形度量，不按某个截图硬补偏移。

圆角、字号、阴影只用 `theme` 的刻度（`scale.go`），不写数字：同一种角色在各组件里看起来一样，改设计只改刻度。卡片用 `surface()`，浮在页面上的层（菜单、弹层、下拉、对话框、通知）用 `floating(层级)`，带阴影。

所有文本类字段（Input、NumberInput、Select、Combobox、TimeField、DatePicker、InputGroup、ColorPicker 的十六进制框，以及 Select、Command、Settings 的搜索框）都用 `ui/kit/field.go` 里同一个 `fieldFrame` 画外框：高 `theme.ControlHeight`（36dp）、左右内边距 10dp、圆角 6dp，边框按"错误 → 聚焦 → 常态"取色，禁用和只读用 Subtle 底色。搜索框统一用 `searchField`，前面带搜索图标。包着文本输入的外框都调用 `FocusOnPress`，点外框任何空白处都聚焦文字。改字段外观只改这一处，不在组件里另写边框和内边距；`TestFieldsShareControlHeight` 检查各字段等高。

框架自己的文字（按钮文案、无障碍名称、占位文字、计数）一律在 Render 时从 `locale.Current()` 读取，不在构造时保存，也不写死中文；`internal/deps` 的测试会拦下写死的中文字符串。拼接"动作 + 对象"形式的名称时用 `locale.Current().Name(action, target)`。应用传进来的文字（标题、菜单项）原样使用。

颜色在 Render 时读取 theme 语义色，不在构造函数中保存主题快照。自定义固定色属于显式覆盖；主题切换不会替应用推断其含义。缓存必须包含主题版本，或使用支持主题失效的 `cx.Cache`。局部主题使用 `cx.Themed`，见 [元素与视图](el.md#局部主题)。

交互组件支持鼠标和键盘；禁用时不能激活或获得焦点，Agent 快照报告 disabled。焦点、按键、定时和浮层使用 [el 提供的 API](el.md)。

浮层组件必须先创建面板并调用 `cx.Overlay`，再渲染 Body / Footer 的内容并追加到面板。内容本身可能登记子浮层；顺序反过来会让子菜单早于父层登记，因锚点尚不可用而被关闭。模态组件用 `Layer.Owner` 绑定所属元素，继承外层的禁用和隐藏。验收必须包含真实嵌套打开、逐层 Esc、焦点返回和祖先禁用，不能只测试独立弹层。

## 文件模板与验收

一个组件对应 `ui/kit/<name>.go`、`<name>_test.go`、`docs/kit/<name>.md`、`examples/components/<name>.go`。在 [组件参考](kit.md) 的对应分类登记；站点侧栏从这个索引生成。示例注册独立 `-section <name>`，展示常用状态、边界和浅深色。

组件文档包含用途、最小用法、公开 API、键盘操作（适用时）、语义、边界和验证入口；测试验证用户可观察行为，不复制实现算法。

`ui/kit/conventions_test.go` 自动检查其中可以机器判断的部分：每个组件都有文档、同名示例 section 和 `ui/window` 中的 Agent 测试；枚举常量带类型名前缀；kit 和 el 的公开 API 没有兼容入口或别名。新增的 Agent 角色不需要在自动化代码里登记，只有需要单独列出子元素的容器角色才加入 `containerRoles`。

提交前逐项检查：

- `uitest` 驱动布局与交互；纯展示组件检查尺寸、约束、状态和颜色。
- `ui/window` Agent 快照覆盖名称、角色、值和状态；新增 role 同步 `automation.go` 与 `docs/automation.md`。
- 有交互时覆盖键盘、禁用、恢复、程序赋值不触发回调；嵌入 el 后连续多帧不丢状态。
- 示例可运行，浅深色和文字布局经过截图检查。
- 公开 API、组件文档、README、示例同时更新。
- `go build ./... && go vet ./ui/... && go test ./... -count=1` 全部通过，包括 `cmd/keel-mcp` 端到端测试。

一个变更只处理一个组件或一项共同能力，验证要求见 [测试](testing.md)。

## 维护文档

- 完整入门示例放在 [快速开始](getting-started.md)，各模块 README 记录职责、依赖和实现文件。
- 组件分类只维护 [组件参考](kit.md)，新增组件同时补文档、示例和 Agent 测试。
- API 变更同步使用文档；影响全局的取舍记录在 [设计决策](decisions.md)。
- 开发进度和历史验收结果放在 `docs/reports/`，不加入文档导航。
- 修改索引后运行 `go test ./internal/site`，检查分类覆盖和站内链接。
