# 组件实现对照与待办

[English](TODO.md) | 简体中文

更新日期：2026-10-02，代码基准 `bc54e85`。来源：[GPUI Kit 组件目录](https://gpui-kit.com/component/)（页面版本 v0.7.0），按导航中的独立组件链接去重，共 **77 项**。组件分类参考该站，说明和实现判断根据 Keel 当前工作区重写；源站文档采用 [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)。这是一份能力对照，不要求复制 Rust API。

“已有”表示 Keel 提供可复用的基础组件，并不表示与 GPUI Kit 功能完全一致；“部分”表示已有实现，但仍缺本表列出的关键能力；“未实现”表示缺少通用实现。最后一列列出建议补齐的能力，不是对源站全部配置项的逐项认证。

## 当前实施清单

版本化进度见 [GPUI Kit 实现进度](docs/reports/gpui-progress-2026-10-02.zh-CN.md)。本文件与 `work/` 按仓库规则仅保存在本地。

见 [GPUI 差距实施清单](work/gpui-gap-plan-2026-10-02.md)：原 A–F 共 36 项，35 项已完成，F3 的完整原生场景验收仍未完成。后续新增能力单列记录，不混入原清单分母；清单完成率不等于 GPUI Kit 功能对齐率。

- [x] WebAssembly：`28a0e80`，浏览器 hello 示例已验证中文、输入、复选框和按钮；构建需 `-tags osusergo`，见 [Web](docs/web.zh-CN.md)。
- [x] 视觉基础：`83c50ed`，圆角/字号/阴影刻度、透明度、字重/等宽/行高与 kit 样式迁移；间距刻度仍待统一。
- [x] Dock 最大化：`bf9f69d`，菜单/双击进入、Esc 恢复、布局保存；中心区自由标签组/分割尚未实现。
- [x] 多主题基础：`bf9f69d`，主题注册、JSON 配色、运行时切换、Nord/Paper 示例；目录监听和渐变主题配置尚未实现。
- [x] CodeEditor 基础：`4f22179`，行号、高亮、撤销重做、输入法、自动缩进、诊断/补全/悬停接口；高级能力见下表。
- [x] 编辑器验证与补全修复：`bc54e85`，真机确认补全出现和回车接受；回归测试覆盖 Agent 补全项、回车/点击接受与关闭。20 万行已实际载入并验证滚动、末尾跳转和输入，未采集帧率或输入延迟。
- [ ] 完整原生验收：仍需复核标题栏、系统偏好、多窗口等；代码编辑器真机可运行不代表 F3 全部通过。
- 系统读屏 / VoiceOver：**暂缓**，未接入，不计为已完成。

## 已完成

- [x] 抓取目录并核对 77 项，逐项附来源和仓库位置。
- [x] 第一批（在 `ui/widget` 中）：Badge、Kbd、Toggle / ToggleGroup、Icon；Button、Checkbox、Switch、Progress、Input、TextArea 增强；所有交互组件统一 `Value` / `SetValue` / `OnChange` / `SetDisabled`。
- [x] Review 修复：命名 `XxxView`，Badge 固定占位，`badge` 语义，测量阶段不改状态（`2eea365`）。

## 里程碑（2026-10-01 制定，状态更新至 2026-10-02）

### 原则

1. **组件统一放在 `ui/kit`**，基于 `ui/el` 实现。旧 `ui/widget` / `ui/layout` 已删除，迁移里程碑保留历史名称。
2. **先补基础设施，再做组件**：里程碑按"组件依赖哪项 el 能力"划分，不按外观分类。同一项能力只实现一次。
3. **每个组件的完成标准**：
   - `uitest` 交互测试；
   - `ui/window` Agent 快照测试，新 role 要同步到 `automation.go` 和 `docs/automation.md`；
   - `examples/components` 增加一个 `-section`；
   - `docs/kit/<组件>.md`；
   - 全量 `go test ./...` 通过。
4. **规范**：
   - 构造函数 `Xxx()` 返回 `*XxxView`；
   - 交互组件统一提供 `Value` / `SetValue`（不触发回调）/ `OnChange` / `SetDisabled`；
   - 装饰的显示、隐藏或内容变化不能改变布局；
   - 不能根据输入是否可用（`gtx.Enabled()`、测量阶段）修改跨帧保存的状态。

### el 基础能力（E 系列）

| 编号 | 能力 | 解锁 |
| --- | --- | --- |
| E1 | 焦点与按键：`Focusable`、`OnKey`、`FocusStyle(style)`、`cx.Focused(id)`、`cx.Focus(id)`，Space/Enter 触发 `OnClick` | 所有可键盘操作的组件 |
| E2 | 禁用：`Disabled(bool)` 向子元素传递，`DisabledStyle` | 所有交互组件 |
| E3 | 时间：`cx.After(key, d, fn)`、按帧时间计算的动画工具，并尊重"减少动画"设置 | Spinner、Skeleton/Shimmer、Tooltip 延迟、Notification 超时、展开动画 |
| E4 | 浮层宿主：`cx.Overlay`、`Anchored(id).Placement(...)` 自动翻转、`OnDismiss`（外部点击、Esc），点击区域与主内容隔离 | 所有弹层 |
| E5 | 焦点约束：`TrapFocus()`，关闭后焦点回到原处 | Dialog、Sheet、Menu |
| E6 | 拖动：指针捕获和位移事件 | Slider（kit）、Resizable、列宽调整、ColorPicker、Dock |
| E7 | 虚拟列表：先做等高行，再做可变高度并保持锚点 | List、Tree、DataTable、Command、大型 Select、MessageScroller |
| E8 | 布局补全：wrap、横向滚动、简单 grid | Tag 组、Tabs 溢出、Calendar、DescriptionList、宽表格 |

### 里程碑

**M0 定规则**（小，先做）
- [x] 把《新组件基于 el，ui/widget 冻结》写进 `docs/decisions.md`。
- [x] 新增 `docs/kit.md`，写明上面的规范和完成标准。
- [x] 在 `internal/deps` 中登记 `ui/kit` 的依赖规则：只依赖 `core`、`theme`、`el`。
- [x] theme 补语义色 `Success`、`Warning`、`Info`。加 `theme.Apply(Palette)`，在帧锁内切换浅色/深色并重绘所有窗口（主题作用域暂不做）。

**M0 review 遗留**（不阻塞，排在 M1 中顺手处理）
- [x] Markdown 的 `CodeBg`、`InlineCode`、`InlineCodeBg` 是写死的浅色包级变量，深色模式下代码块会变成浅底浅字，看不清。改成 Render 时从 theme 派生（新增 `CodeBg` 等语义色，或加 `markdown.Palette` 跟随 `theme.Revision()`）。
- [x] 深色配色里 `Primary #2563eb` 叠在 `Highlight #1e3a5f` 上对比度约 2.6:1，选中态 Toggle 的文字不够清楚。拆出 `PrimaryText` 一类的文本用色，或把深色 `Highlight`/`Primary` 调到对比度 ≥ 4.5:1。

**M1 kit 试点与 el 交互基础**：E1、E2、E3
- [x] 先做纯展示组件，验证 kit 的目录、文档和测试模板：Alert、Empty、Avatar、Tag（不可移除）、DescriptionList、GroupBox、StatusBar、Marker。
- [x] E1 焦点与按键实现、测试和示例完成；已按 m1-plan.md 纠偏，补齐 E2；等待整体复核。
- [x] E1、E2 完成并 review 后，做依赖焦点的部分：Tag 可移除、Alert 关闭按钮。
- [x] E3 完成后：Spinner、Skeleton、Shimmer。
- [x] 公共测试工具与 Icon；E3 时间能力、显式 key、减少动画。
- [x] macOS 系统“减少动画”偏好已桥接（`431e854`），保留应用覆盖和恢复跟随接口；实时原生通知验收仍在 F3 中。
- 出口：kit 现有 11 个组件，Shimmer 为 Skeleton 选项；逐项测试和浅深色截图已验证。
- [x] M1 review 修复（R1–R5：API 去重、After 显式 key、插槽改为 View、section 命名、文档结构），提交 `c1784c3`。

**M2 浮层**：E4、E5
- [x] E4 浮层宿主、E5 焦点约束：实现、测试、文档与 overlay 示例及第 1 步 review 修复完成。（执行方案见 `work/m2-plan.md`；kit.Button、kit.Kbd 从 M3 提前到 M2）
- [x] kit 枚举前缀统一（74ccb37）；kit.Button 实现、测试、文档和示例完成（40cb039），Button review B1 / B2 已修复（f6963c0）；Kbd 完成（39f27d1），等待本轮 review。
- [x] Popover → Tooltip（键盘焦点也显示）→ HoverCard → Menu（分隔线、子菜单、方向键）→ DropdownButton。
- [x] kit 版 Dialog / AlertDialog，替代 `widget.Dialog` 和 `window.Options.Overlay` 的写法；再做 Sheet。
- [x] Notification：队列、超时、关闭。
- [x] Clipboard 复制按钮（从 Markdown 代码块提取，带 Tooltip 反馈）。
- 出口：`examples/orders` 的对话框改用 kit，`TestOrders` 通过。
- [x] 规则检查测试（`ui/kit/conventions_test.go`）、Agent 角色默认放行、框架文字集中到 `ui/locale`（多语言）。

**M3 表单与输入**：可能需要先扩展 el.Input（过滤、全选、前后缀）
- [x] 迁移到 kit：Button → Checkbox / Switch / Radio → Select → Slider（依赖 E6）→ Toggle / ToggleGroup。
- [x] Input Group、NumberInput、OtpInput、TimeField。
- [x] Combobox（Input + Popover + 过滤）；Select 补过滤和分组。
- [x] Calendar（依赖 E8 grid）→ DatePicker（Calendar + Popover）→ 日期范围选择。
- [x] 表单模型：字段校验、错误状态、提交。Rating、Stepper。
- 出口：`examples/orders` 的"新建订单"表单全部使用 kit。
- [x] el：拖动事件、输入框 MaxLen/Filter/ReadOnly、禁用子树里的输入框不可编辑；locale 增加日历和表单文字。
- [x] 单行输入框的 `OnKey` 先于编辑器拿到 ↑ ↓ PageUp PageDown：NumberInput、TimeField 用来调整数值，Combobox 移动高亮，Select 搜索框进入列表。

**M4 数据与长内容**：E7，以及 E6 / E8 的剩余部分
- [x] List → VirtualList → Tree。
- [x] kit 版 Table / DataTable：单元格插槽、列宽拖动、Pagination、异步数据。
- [x] Table 横向滚动和左右冻结列：`6589c17`、`749b10a`，含裁剪命中、窄视口、1×/2× 与万行虚拟化回归。
- [x] Command 命令面板（浮层 + 虚拟列表 + 模糊过滤 + Kbd）。
- [x] 从 `examples/chat` 提取 Message、Bubble、MessageScroller（历史加载锚点）、Attachment。
- 出口：`examples/orders` 的表格改用 kit，`ui/widget` 在示例中没有引用。
- [x] 迁移 Tabs、Accordion、Badge、Progress、Link、Image 到 kit；hello、hotkey、multiwindow、components、chat、orders 示例全部不再引用 `ui/widget`。
- [x] el：ScrollState / ScrollIntoView、KeepBottomOn、Flex 权重。
- [x] 可变高度列表与锚点保持（`e13ead8`）、el 横向滚动（`027bb17`）、Table 横向滚动和冻结列已补齐。

**M5 应用外壳**
- [x] Sidebar、Toolbar（溢出菜单）、Tabs（溢出菜单、关闭）、Resizable、Settings、Carousel。
- [x] Dock：可停靠面板、布局保存（DockLayout 可编码为 JSON）。
- [x] TitleBar：`window.Options.Frameless` + `kit.TitleBar`，基于 Gio 的无边框窗口和 ActionMove 拖动，见 docs/decisions.md《自定义标题栏：用 Gio 的无边框窗口实现》。示例 `examples/frameless`。
- [x] TitleBar 的 macOS 双击偏好与窗口失焦外观已实现（`2d006c3`）；原生交互仍待 F3 复核。
- [x] el：`Cursor(...)`；修复模态浮层关闭后第一次点击丢失。

**M6 可视化与专项**
- [x] Chart（折线、柱状、堆叠、坐标轴、图例、悬停提示、数据表视图）→ Plot（缩放、平移、拾取、键盘）。theme 增加 `Chart` 分类色。
- [x] ColorPicker（HSV、透明度、十六进制、预设色）、Questionnaire（五种题型、必填校验、分页、答案模型）。

**收尾：删除 ui/widget**
- [x] 删除 `ui/widget` 和 `ui/layout`；Markdown 图片移到 `ui/internal/imageload`；文档改为指向 kit（76c7427）。

### 暂不做（需要时再单独决策）

- Editor 的搜索替换、代码折叠、多光标/矩形选择、语言级括号配对；行号、语法高亮等基础代码编辑能力已完成。
- TextView 的 HTML 富文本、完整 TeX。
- 主题作用域（局部覆盖主题）。
- Kbd 的动作绑定查询：等快捷键注册表统一后再做。
- 系统读屏 / VoiceOver：按当前决定暂缓。

## 当前组件对照

保留原 77 项目录口径，仓库位置已更新到当前实现。多个目录项可由同一组件或底层能力承接；“—”表示此表没有登记进一步缺口，不表示已认证源站全部 API。原生验证限制见 F3。

| GPUI Kit 组件 | 状态 | Keel 当前能力 / 位置 | 待补齐 |
| --- | --- | --- | --- |
| [Accordion](https://gpui-kit.com/component/accordion/) | 已有 | [单项/多项、自定义标题、动画、键盘、禁用](ui/kit/accordion.go) | — |
| [AlertDialog](https://gpui-kit.com/component/alert-dialog/) | 已有 | [提示/确认/危险对话框、焦点约束与恢复](ui/kit/dialog.go) | — |
| [Alert](https://gpui-kit.com/component/alert/) | 已有 | [行内提示、级别、关闭按钮](ui/kit/alert.go) | — |
| [Attachment](https://gpui-kit.com/component/attachment/) | 已有 | [附件卡片、进度、取消、重试、错误状态](ui/kit/attachment.go) | — |
| [Avatar](https://gpui-kit.com/component/avatar/) | 已有 | [图片/首字母回退、尺寸、状态标记](ui/kit/avatar.go) | — |
| [Badge](https://gpui-kit.com/component/badge/) | 已有 | [数字、圆点、图标、尺寸、颜色、角标](ui/kit/badge.go) | — |
| [Bubble](https://gpui-kit.com/component/bubble/) | 已有 | [可复用聊天气泡、用户操作栏](ui/kit/message.go) | — |
| [Button](https://gpui-kit.com/component/button/) | 已有 | [主/次要/危险、禁用、尺寸、图标、加载](ui/kit/button.go) | — |
| [Calendar](https://gpui-kit.com/component/calendar/) | 已有 | [年月切换、多月、范围、禁用日期、键盘](ui/kit/calendar.go) | — |
| [Carousel](https://gpui-kit.com/component/carousel/) | 已有 | [轮播、指示器、键盘、禁用与定时暂停](ui/kit/carousel.go) | — |
| [Chart](https://gpui-kit.com/component/chart/) | 已有 | [折线/柱状/面积，另有饼图/环图/蜡烛图、图例与数据表](ui/kit/chart.go) | — |
| [Checkbox](https://gpui-kit.com/component/checkbox/) | 已有 | [布尔选择、半选、回调、禁用](ui/kit/checkbox.go) | — |
| [Clipboard](https://gpui-kit.com/component/clipboard/) | 已有 | [通用复制按钮、提示与连续复制反馈](ui/kit/copy_button.go) | — |
| [Collapsible](https://gpui-kit.com/component/collapsible/) | 已有 | [独立 Trigger/Content、动画、焦点恢复](ui/kit/collapsible.go) | — |
| [ColorPicker](https://gpui-kit.com/component/color-picker/) | 已有 | [HSV、透明度、HEX、预设、键盘、禁用](ui/kit/color_picker.go) | — |
| [Combobox](https://gpui-kit.com/component/combobox/) | 已有 | [过滤、多选标签、异步结果、重试、虚拟化](ui/kit/combobox.go) | — |
| [Command](https://gpui-kit.com/component/command/) | 已有 | [模糊过滤、分组、快捷键、异步结果、虚拟化](ui/kit/command.go) | — |
| [DataTable](https://gpui-kit.com/component/data-table/) | 已有 | [横向滚动、冻结列、列管理、多选/单元格选择、复制、筛选、分页加载](ui/kit/table.go) | — |
| [DatePicker](https://gpui-kit.com/component/date-picker/) | 已有 | [日期输入、日历弹层、范围、取消草稿、键盘](ui/kit/date_picker.go) | — |
| [DescriptionList](https://gpui-kit.com/component/description-list/) | 已有 | [标签/值布局、响应式列数](ui/kit/description_list.go) | — |
| [Dialog](https://gpui-kit.com/component/dialog/) | 已有 | [可组合内容、嵌套浮层、长内容、焦点约束与恢复](ui/kit/dialog.go) | — |
| [Dock](https://gpui-kit.com/component/dock/) | 部分 | [边缘标签组拖放、嵌套分割、布局保存、最大化](ui/kit/dock.go) | 中心区自由标签组/分割；跨窗口 Dock 未纳入本轮 |
| [DropdownButton](https://gpui-kit.com/component/dropdown_button/) | 已有 | [按钮菜单、分体按钮、键盘与焦点恢复](ui/kit/dropdown_button.go) | — |
| [Editor](https://gpui-kit.com/component/editor/) | 部分 | [行号、后台高亮、选择/撤销/输入法、自动缩进、诊断/补全/悬停接口](ui/kit/code_editor.go) | 搜索替换、折叠、多光标/矩形选择、语言级括号配对；LSP 客户端由应用提供 |
| [Empty](https://gpui-kit.com/component/empty/) | 已有 | [空状态标题、说明与操作](ui/kit/empty.go) | — |
| [Focus Trap](https://gpui-kit.com/component/focus-trap/) | 已有 | [弹层焦点循环、关闭后返回焦点](ui/el/overlay.go) | — |
| [Form](https://gpui-kit.com/component/form/) | 已有 | [字段组织、校验、错误聚焦、异步提交/取消](ui/kit/form.go) | — |
| [GroupBox](https://gpui-kit.com/component/group-box/) | 已有 | [标题、描述与内容分组](ui/kit/group_box.go) | — |
| [HoverCard](https://gpui-kit.com/component/hover-card/) | 已有 | [悬停卡片、延迟、定位、跨目标与取消](ui/kit/hover_card.go) | — |
| [Icon](https://gpui-kit.com/component/icon/) | 已有 | [内置矢量图标、自定义图标、尺寸与颜色](ui/kit/icon.go) | 按应用需要扩充图标 |
| [Image](https://gpui-kit.com/component/image/) | 已有 | [异步缓存、适配/裁剪/拉伸、圆角、预览、失败重试](ui/kit/image.go) | — |
| [Input Group](https://gpui-kit.com/component/input-group/) | 已有 | [独立组合容器、前后内容、统一边框、标签聚焦](ui/kit/input_group.go) | — |
| [Input](https://gpui-kit.com/component/input/) | 已有 | [单行、密码、长度、前后缀、清空、校验、禁用、标签聚焦](ui/kit/input.go) | — |
| [Kbd](https://gpui-kit.com/component/kbd/) | 部分 | [平台键帽、动态文案、尺寸、无边框](ui/kit/kbd.go) | 动作绑定查询暂缓 |
| [Label](https://gpui-kit.com/component/label/) | 已有 | [Text、字号/颜色、For 标签关联聚焦](ui/el/element.go) | — |
| [List](https://gpui-kit.com/component/list/) | 已有 | [列表项、稳定 ID、单项禁用、多选/范围、键盘、拖动](ui/kit/list.go) | — |
| [Marker](https://gpui-kit.com/component/marker/) | 已有 | [标记形状、大小和颜色](ui/kit/marker.go) | — |
| [Menu](https://gpui-kit.com/component/menu/) | 已有 | [菜单、子菜单、分隔线、长内容、方向键、焦点恢复](ui/kit/menu.go) | — |
| [MessageScroller](https://gpui-kit.com/component/message-scroller/) | 已有 | [可变高度虚拟化、跟随尾部、流式增高、历史加载锚点](ui/kit/message_scroller.go) | — |
| [Message](https://gpui-kit.com/component/message/) | 已有 | [消息内容、状态、操作栏、反应、失败重试](ui/kit/message.go) | — |
| [Notification](https://gpui-kit.com/component/notification/) | 已有 | [通知队列、超时、关闭、暂停与原位更新](ui/kit/notification.go) | — |
| [NumberInput](https://gpui-kit.com/component/number-input/) | 已有 | [数值解析、范围/步长/精度、草稿提交与取消](ui/kit/number_input.go) | — |
| [OtpInput](https://gpui-kit.com/component/otp-input/) | 已有 | [分格输入、粘贴、退格、焦点移动、窄布局](ui/kit/otp_input.go) | — |
| [Pagination](https://gpui-kit.com/component/pagination/) | 已有 | [页码、前后翻页、总数、窄布局换行](ui/kit/pagination.go) | — |
| [Plot](https://gpui-kit.com/component/plot/) | 已有 | [缩放、平移、数据拾取、键盘、异常数据保护](ui/kit/plot.go) | — |
| [Popover](https://gpui-kit.com/component/popover/) | 已有 | [锚点定位、避让、长内容、外部点击/Esc、焦点恢复](ui/kit/popover.go) | — |
| [Progress](https://gpui-kit.com/component/progress/) | 已有 | [确定进度、不确定动画、模式切换](ui/kit/progress.go) | — |
| [Questionnaire](https://gpui-kit.com/component/questionnaire/) | 已有 | [题型、答案模型、校验、分页、禁用与提交快照](ui/kit/questionnaire.go) | — |
| [Radio](https://gpui-kit.com/component/radio/) | 已有 | [单选组、横纵布局、独立 Item、单项禁用、键盘](ui/kit/radio_group.go) | — |
| [Rating](https://gpui-kit.com/component/rating/) | 已有 | [评分、半星/小数展示、只读、键盘](ui/kit/rating.go) | — |
| [Resizable](https://gpui-kit.com/component/resizable/) | 已有 | [横纵分割、最小尺寸、拖动、键盘、取消与禁用](ui/kit/resizable.go) | — |
| [Root View](https://gpui-kit.com/component/root/) | 已有 | [根布局、统一浮层宿主、窗口快捷键](ui/el/root.go) | — |
| [Scrollable](https://gpui-kit.com/component/scrollable/) | 已有 | [ScrollX/ScrollY、滚动条拖动/轨道点击、定位与尾部跟随](ui/el/viewport.go) | 通过 el 组合，没有独立 kit.Scrollable 类型 |
| [Select](https://gpui-kit.com/component/select/) | 已有 | [过滤、分组、多选、禁用项、万条虚拟化](ui/kit/select.go) | — |
| [Settings](https://gpui-kit.com/component/settings/) | 已有 | [设置分组、导航、搜索、窄布局](ui/kit/settings.go) | — |
| [Sheet](https://gpui-kit.com/component/sheet/) | 已有 | [侧边抽屉、遮罩、长内容、焦点与禁用继承](ui/kit/sheet.go) | — |
| [Shimmer](https://gpui-kit.com/component/shimmer/) | 已有 | [Skeleton 的扫光选项、减少动画](ui/kit/skeleton.go) | — |
| [Sidebar](https://gpui-kit.com/component/sidebar/) | 已有 | [嵌套分组、收起、选中、固定头尾、键盘滚动](ui/kit/sidebar.go) | — |
| [Skeleton](https://gpui-kit.com/component/skeleton/) | 已有 | [占位形状、尺寸、加载展示](ui/kit/skeleton.go) | — |
| [Slider](https://gpui-kit.com/component/slider/) | 已有 | [单值/双端范围、横向/竖向、步长、拖动与键盘](ui/kit/slider.go) | — |
| [Spinner](https://gpui-kit.com/component/spinner/) | 已有 | [不确定动画、减少动画、可访问名称](ui/kit/spinner.go) | — |
| [StatusBar](https://gpui-kit.com/component/status-bar/) | 部分 | [固定状态栏、左右内容组、单行文本](ui/kit/status_bar.go) | 独立溢出菜单尚未实现 |
| [Stepper](https://gpui-kit.com/component/stepper/) | 已有 | [步骤状态、导航、键盘、横向滚动与禁用](ui/kit/stepper.go) | — |
| [Switch](https://gpui-kit.com/component/switch/) | 已有 | [布尔开关、尺寸、加载、禁用](ui/kit/switch.go) | — |
| [Table](https://gpui-kit.com/component/table/) | 已有 | [排序、行选择、单元格插槽、列宽调整；高级能力同 DataTable](ui/kit/table.go) | — |
| [Tabs](https://gpui-kit.com/component/tabs/) | 已有 | [页面状态、溢出、关闭与焦点恢复、拖动排序](ui/kit/tabs.go) | — |
| [Tag](https://gpui-kit.com/component/tag/) | 已有 | [颜色、移除、选中](ui/kit/tag.go) | — |
| [TextView](https://gpui-kit.com/component/text-view/) | 部分 | [Markdown、公式子集、图片、选择复制、代码块、流式渲染](ui/markdown/README.zh-CN.md) | HTML 富文本、完整 TeX 暂缓 |
| [Textarea](https://gpui-kit.com/component/textarea/) | 已有 | [多行、只读、自动高度、最大可见行数、错误恢复](ui/kit/input.go) | — |
| [Theme](https://gpui-kit.com/component/theme/) | 部分 | [语义配色、字号/圆角/阴影、浅深切换、注册与 JSON 主题](ui/theme/registry.go) | 局部作用域、目录监听、渐变主题配置、统一间距刻度 |
| [TimeField](https://gpui-kit.com/component/time-field/) | 已有 | [时分秒、步进/进位、12/24 小时、Tab 与本地化](ui/kit/time_field.go) | — |
| [TitleBar](https://gpui-kit.com/component/title-bar/) | 已有 | [自定义标题栏、窗口控制、macOS 双击偏好与失焦外观](ui/kit/title_bar.go) | 实现已有，原生行为验收仍待 F3 |
| [Toggle](https://gpui-kit.com/component/toggle/) | 已有 | [状态按钮、图标、尺寸、单选/多选组](ui/kit/toggle_group.go) | — |
| [Toolbar](https://gpui-kit.com/component/toolbar/) | 已有 | [左右区域、尺寸、工具分组、溢出与键盘](ui/kit/toolbar.go) | — |
| [Tooltip](https://gpui-kit.com/component/tooltip/) | 已有 | [通用提示、键盘焦点、延迟与取消](ui/kit/tooltip.go) | — |
| [Tree](https://gpui-kit.com/component/tree/) | 已有 | [虚拟化、展开、多选、单项禁用、键盘、动态数据与拖动](ui/kit/tree.go) | 拖动时自动滚动/展开未实现 |
| [VirtualList](https://gpui-kit.com/component/virtual-list/) | 已有 | [等高及可变高度实现、稳定 key、尺寸缓存、插入保持锚点](ui/kit/variable_list.go) | — |


## 2026-10-02 查漏补缺

- [x] VirtualList 缩减数据后空白视口，d893c51。
- [x] Calendar 月末 PageUp / PageDown 跳错月份，50eee10。
- [x] Form 静默跳过非 Validatable 控件校验、Required 漏判 Unicode 空白，c7878d1。
- 审查范围、证据和剩余缺口见 work/audit-2026-10-02.md。
