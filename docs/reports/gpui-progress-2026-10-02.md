# GPUI Kit 实现进度 · 2026-10-02

更新日期：2026-10-03（原报告 2026-10-02，代码基准 `2fe8d1d`，本轮逐页复核 77 项公开文档及 Keel 公共接口/相关实现）。来源：[GPUI Kit 组件目录](https://gpui-kit.com/component/)（页面版本 v0.7.0），按导航中的独立组件链接去重，共 **77 项**。组件分类参考该站，说明和实现判断根据 Keel 当前工作区重写；源站文档采用 [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)。这是一份能力对照，不要求复制 Rust API。

补齐后状态：**46 项主体已有、28 项部分覆盖、3 项用途不同**（初次复核为 38/36/3）。这是按文档列出的功能判断，不是功能完成百分比，也不是视觉成熟度评分。

- **主体已有**：核心用途覆盖；末列仍列出配置、交互或组合方式差异，不能读成全部功能相同。
- **部分**：已有可运行的主体，但缺源站明确提供的扩展功能或组合能力；已实现部分继续记为完成。
- **用途不同**：旧表拿同名或近似效果代替了不同用途的组件，需重新建立对应关系。

本轮是文档与源码静态复核，检查了跨文件扩展接口；没有重跑 77 项真机交互，也没有逐项证明视觉一致。缺口指组件当前未提供的接口/行为，应用通过 `el` 自行组合不自动算作 kit 已封装。GPUI 文档示例不等于上游源码和所有平台均已验证。

## 本轮更正与重点差距

1. Marker、Plot、Shimmer 三项不能继续算“同等组件已有”：分别是消息标记行 vs 几何图形、底层绘图工具集 vs 成品图、文字扫光 vs 骨架屏扫光。
2. 数据与输入组件仍有实质差距：Chart 缺雷达/桑基图；DatePicker 缺时间联动和预设；Input/Textarea 缺原子 token，Input 另缺格式 mask；Progress 的圆形进度缺口已在第一批补齐。
3. Editor 已有多光标、查找替换、折叠和括号配对，但没有编辑跟踪装饰集合、开放语言规则与完整搜索会话。`OnComplete`/`OnHover`/`OnDefinition` 是应用接口，LSP 客户端仍由应用提供；本轮不把它当作已证实的上游内置能力差距。
4. TextView 已有 Markdown/HTML/扩展 TeX，但富文本折叠预览、流式逐段淡入、区间高亮/定位和插件仍缺。完整 TeX/CSS 是 Keel 的边界，不能无依据当作 GPUI 已有功能。
5. Dock、主题、状态栏和 Kbd 的近期补齐继续保留完成记录。Dock 分离由应用开窗、恢复布局不会重开分离窗口；主题机制已有，预设数量和 token 格式仍不同。
6. 旧表夸大了 Badge 图标/尺寸、DescriptionList 响应式列数、Kbd 尺寸、Switch 尺寸/加载、Toggle 尺寸、Textarea 最大行数；均按当前接口改正。

## 分批补齐记录

- [x] 第一批（`945b090`）：ProgressCircle，支持 0–100% 圆环、不确定动画、减少动画、中心内容、尺寸/颜色和 Agent 语义；含非有限值/约束测试、动画像素测试及浅色 1×/深色 2× 截图检查。文档和示例已登记。
- [x] 第二批（`11fd62f`）：OtpInput 的遮罩、分组与尺寸配置；遮罩同时覆盖可见数字和 Agent 语义值，编辑回调保留原始数字。测试覆盖分组边界、1×/2× 窄布局、程序赋值不触发回调；浅深色截图已检查。
- [x] 第三批（`edac67d`）：HoverCard 的 OpenDelay/CloseDelay、Placement 和 Offset；自定义及零延时、等待中改延时、四方向定位、Esc 和 Agent 快照通过测试，已打开卡片的浅深色截图已检查。
- [x] 第四批（`7a25711`）：Tooltip 的富内容、动作键位和定位/间距配置；快捷键随改绑更新，内容不接受点击，Agent 展开富内容与键位；相关测试和浅深色弹层截图已检查。
- [x] 第五批（`a118ee5`）：Clipboard 的 OnCopied、Content 和 Copied；回调接收实际提交的原文，自定义内容保留键盘操作、禁用继承与连续复制计时。含回调/空取值/计时测试、Agent 快照及浅深色截图检查。
- [x] 第六批（`204309e`）：Stepper 的竖向布局、步骤图标、单步禁用与尺寸；另支持富内容条目。覆盖切片隔离、更新索引收敛、键盘跳过禁用、祖先禁用、窄窗口纵向滚动及 Agent 状态；浅深色截图已检查。
- [x] 第七批（`ee6ef65`）：AvatarGroup 叠放、组尺寸、人数上限、+N/省略号、窄窗口滚动；Avatar.Source 后台加载、取消、失败回退与重试。共享解码器迁至 core.DecodeImage，Markdown 继续复用。含网络/旧请求覆盖、布局、Agent 和浅深色截图验证。
- [x] 第八批（`40f241c`）：DescriptionList 多列/跨列、纵向标签、满行分隔线、边框与字号；el.Grid 增加 ColSpan。覆盖跨列换行/最小宽度、1×/2× 窄布局、富内容交互及 Agent 重排；浅深色截图已检查。
- [x] 第九批（`a6c142c`）：InputGroup 四方向 Addon、多附加内容、稳定 ID 替换/移除与 TextArea 组合。按钮沿用 kit.Button 的样式/尺寸/加载/禁用配置。覆盖布局顺序、文字聚焦、按钮焦点、Tab、只读、禁用继承、动态内容及 Agent；浅深色截图已检查。
- [x] 第十批（`ad5ea79`）：Slider 对数刻度、OnRelease/OnRangeRelease；覆盖单值/双端和横纵向映射、键盘重复、取消、禁用、非法/极端范围、Agent 回调及浅深色截图。正步长仍按数值对齐，无正步长的对数键盘操作按轨道百分比移动。
- [x] 第十一批（`5043522`）：Button 新增 Link/Text/Success/Warning/Info，支持叠加 Outline/Compact、自定义 Content 与 Appearance、正方形图标按钮。自定义内容加载时保持尺寸并显示进度环；语义底色自动选择黑白文字。覆盖键盘/加载焦点、1×/2× 布局和像素测试，浅深色截图已检查。
- [x] 第十二批（`8607d0d`）：Tabs 增加 Underline/Pill/Outline/Segmented 外观、TabItem 图标/自定义标签、SetItem 与单项禁用。点击、关闭、键盘、拖动和溢出菜单遵守禁用状态；禁用当前项自动选择可用项，程序更新不触发回调。覆盖 1×/2× 布局、状态保留、Agent 和浅深色截图。新发现的最大标签宽度与滚动标签栏接口仍记为未完成。
- [x] 第十三批（`cb981a6`）：Tabs 的 MaxWidth、Scrollable、ScrollTo 与 ScrollState；键盘/程序切换自动定位，先滚动再转移焦点，单独滚动不改变选中页。覆盖首次显示前请求、身份重排、禁用项跳过、1×/2×、溢出模式切换和 Agent 连续方向键；浅深色滚动栏截图已检查。
- [x] 第十四批（`61b303f`）：Tag 的 Outline、Size、Rounded、Content、Appearance；支持主题派生及自定义背景/前景/边框/选中背景。保留选择/移除行为，覆盖窄布局、1×/2×、键盘、祖先禁用、程序赋值、Agent 与颜色/圆角像素测试；浅深色截图已检查。
- [x] 第十五批（`7097862`）：Alert 的 Banner、四档 Size、Icon/IconNone 与 Content；支持 Markdown/操作按钮，横幅省略独立标题行并保留可访问名称。覆盖尺寸与窄布局、图标间距、正文操作/关闭隔离、祖先禁用、恢复与 Agent；浅深色截图已检查。
- [x] 第十六批（`43206c1`）：GroupBox 增加 Normal/Fill/Outline 外观、框外 Footer、TitleStyle 与 ContentStyle；保留原 Surface 默认。正文保持稳定身份，覆盖外观切换、标题增删后的输入与焦点、1×/2× 窄布局、footer 交互及 Agent 样式隔离；浅深色截图已检查。
- [x] 第十七批（`a715aad`）：Badge 的 Icon、Size、Color 与 Name；图标放右下角并加 Surface 边框，数字/圆点保持右上角且不改变子组件布局。自定义底色自动选择黑白前景。覆盖模式切换、零值、数字上限、1×/2×、子组件点击、Agent 与像素测试；浅深色截图已检查。
- [x] 第十八批（`5c83f02`）：Accordion 的 Bordered 与四档 Size；边框开关同时控制外框和分节线，尺寸统一调整间距、箭头和继承字号。保留默认字号继承及 Collapsible 原行为；覆盖 1×/2× 布局、状态保留、自定义标题、键盘跳过禁用项与 Agent 快照。全量构建、vet、测试通过，浅色 1×/深色 2× 截图已检查。
- [x] 第十九批（`72b8472`）：Spinner 的 Icon、VectorIcon 与 Color；圆环和自定义图标共享帧时钟与减少动画策略，可恢复默认圆环。像素测试覆盖旋转、静止、自定义颜色及恢复，Agent 语义保持不变；全量构建、vet、测试及浅色 1×/深色 2× 截图检查通过。
- [x] 第二十批（`8a59f77`）：Spinner 增加 Period 旋转周期；0 恢复一秒、负值忽略，减少动画保持优先。注入帧时间的像素测试验证圆环/自定义图标的两秒周期、整周重复、默认恢复与静止；全量构建、vet、测试通过。
- [x] 第二十一批（`2f6f1c2`）：Empty 增加 Media，支持头像、图片与任意 View；nil 恢复图标，IconNone 隐藏回退。媒体、标题、说明和操作区使用稳定身份，测试覆盖媒体替换后的输入与焦点、1×/2× 窄布局、主题切换、Agent 语义与操作按钮；全量构建、vet、测试及浅深色截图检查通过。
- [x] 第二十二批（`5b8fbce`）：Empty 增加 Heading、DescriptionContent、Footer 与七个分区的 PartStyle；富内容可恢复原字符串，尾部独立于 Action。覆盖窄布局、替换/恢复后的输入焦点、尾部操作与禁用继承、富内容主题切换；全量构建、vet、测试及浅深色截图检查通过。
- [x] 第二十三批：Skeleton 增加 Secondary 与 Rounded，整体透明度减半、任意有限非负圆角并按短边限制；Circle/Rounded 后调用者生效。像素测试覆盖浅深主题、直角/圆角/极大圆角、普通/次级的脉冲与扫光、减少动画及装饰语义；全量构建、vet、测试和浅色 1×/深色 2× 截图通过。
- 后续差异继续以 77 项表中末列为准。

## 当前实施清单

历史实施清单（沿用原验收口径，本轮未重新执行）：原 A–F 共 36 项，35 项已完成，F3 的完整原生场景验收仍未完成。后续新增能力单列记录，不混入原清单分母；清单完成率不等于 GPUI Kit 功能对齐率。

- [x] WebAssembly：`28a0e80`，浏览器 hello 示例已验证中文、输入、复选框和按钮；构建需 `-tags osusergo`，见 [Web](../../docs/web.md)。
- [x] 视觉基础：`83c50ed`，圆角/字号/阴影刻度、透明度、字重/等宽/行高与 kit 样式迁移；间距刻度 `08daebf` 已统一。
- [x] Dock 最大化：`bf9f69d`，菜单/双击进入、Esc 恢复、布局保存；中心区文档标签组与拆分、跨窗口分离见下表。
- [x] 多主题基础：`bf9f69d`，主题注册、JSON 配色、运行时切换、Nord/Paper 示例；目录监听、渐变配置 `08daebf` 已完成。
- [x] CodeEditor 基础：`4f22179`，行号、高亮、撤销重做、输入法、自动缩进、诊断/补全/悬停接口；高级能力见下表。
- [x] 编辑器验证与补全修复：`bc54e85`，真机确认补全出现和回车接受；回归测试覆盖 Agent 补全项、回车/点击接受与关闭。20 万行已实际载入并验证滚动、末尾跳转和输入，未采集帧率或输入延迟。
- [x] 后续补齐（2026-10-03）：编辑器多光标/查找替换/折叠/括号 `56fa4f6`；局部主题与内置主题 `cc8e536`；Windows、Linux 原生能力 `643ceb6`（仅交叉编译验证）；无样式基础层 `d4d38b5`；状态栏溢出与动作键位 `3c4e511`；主题渐变/间距/目录监听 `08daebf`；HTML 与扩展 TeX `2bd4c61`；Dock 中心文档与跨窗口接口 `b66875e`。
- [x] 新窗口初始居中：`2fe8d1d`，macOS 标准/无边框窗口与后续不强制回中已做原生验证。
- [ ] 完整原生验收：标题栏、系统偏好与跨平台多窗口仍未全部通过；窗口居中局部验收不代表 F3 完成。
- 系统读屏 / VoiceOver：**暂缓**，未接入，不计为已完成。

## 当前组件对照

保留源站 77 项目录口径，每行链接对应官方文档与 Keel 主实现文件。跨文件实现的审计线索见表后；Table/DataTable 等不能仅按组件数量判断对齐。

| GPUI Kit 组件 | 本轮状态 | Keel 已完成能力 / 主实现 | 已确认的差异与边界 |
| --- | --- | --- | --- |
| [Accordion](https://gpui-kit.com/component/accordion/) | 主体已有 | [单项/多项、自定义标题、动画、键盘、禁用、边框开关与四档尺寸](../../ui/kit/accordion.go) | 第十八批已关闭登记缺口；无边框保留背景和圆角，默认 Medium 保留字号继承。自定义标题和正文的显式字号优先。 |
| [AlertDialog](https://gpui-kit.com/component/alert-dialog/) | 主体已有 | [提示/确认/危险对话框、焦点约束与恢复](../../ui/kit/dialog.go) | 行为/配置差异：Persistent 只禁止点击遮罩关闭，Esc 仍关闭；没有独立 keyboard 开关。内置确认按钮先关闭再执行回调，不能用返回值阻止关闭；可自组 Footer。 |
| [Alert](https://gpui-kit.com/component/alert/) | 主体已有 | [行内/横幅提示、级别、四档尺寸、可替换图标、富正文、关闭按钮](../../ui/kit/alert.go) | 第十五批已关闭登记缺口；Content 可组合 Markdown 与操作按钮。横幅没有独立标题行，无正文时使用标题作为消息；自定义内容的内部样式由内容自身控制。 |
| [Attachment](https://gpui-kit.com/component/attachment/) | 部分 | [附件卡片、进度、取消、重试、错误状态](../../ui/kit/attachment.go) | 缺媒体/图片预览槽、横纵布局、附件组；当前是文件名/大小卡片，已有上传进度与失败操作。 |
| [Avatar](https://gpui-kit.com/component/avatar/) | 主体已有 | [图片/首字母回退、URL 加载与重试、尺寸、状态标记](../../ui/kit/avatar.go)、[叠放头像组/上限/+N/省略号](../../ui/kit/avatar_group.go) | 第七批已关闭原登记缺口；加载不跨实例缓存。外观仍为圆形和主题色回退，GPUI 的自定义占位图标、边框/圆角等样式接口及配色算法不同。 |
| [Badge](https://gpui-kit.com/component/badge/) | 主体已有 | [数字/圆点/图标、上限、尺寸、自定义颜色/名称与角标容器](../../ui/kit/badge.go) | 第十七批已关闭登记缺口；图标模式不依赖计数，数字/圆点仍在 count≤0 时隐藏。Size 为 dp，圆点按比例缩放；Tone 清除固定颜色覆盖。 |
| [Bubble](https://gpui-kit.com/component/bubble/) | 部分 | [可复用内容气泡、Mine 对齐](../../ui/kit/message.go) | 缺 ghost 等外观变体、气泡组和独立反应槽；当前只接 content + Mine，操作/反应在 Message 层。 |
| [Button](https://gpui-kit.com/component/button/) | 主体已有 | [九种变体、描边/紧凑、自定义内容/配色、禁用、尺寸、图标、加载](../../ui/kit/button.go) | 第十一批已关闭登记的变体、样式和内容缺口。Outline/Compact 为叠加配置；自定义内容限展示元素。Tooltip 可外部组合，本项不表示与上游所有组合接口完全相同。 |
| [Calendar](https://gpui-kit.com/component/calendar/) | 主体已有 | [年月切换、多月、范围、禁用日期、键盘](../../ui/kit/calendar.go) | 主体覆盖；禁用日期用函数、年份限制可用 Bounds 表达。缺组件尺寸档，API 组织不同。 |
| [Carousel](https://gpui-kit.com/component/carousel/) | 部分 | [轮播、指示器、键盘、禁用与定时暂停](../../ui/kit/carousel.go) | 缺竖向轨道、同屏多项、可组合前后控件；Keel 每次只显示一张，另有自动播放。 |
| [Chart](https://gpui-kit.com/component/chart/) | 部分 | [折线/柱状/面积/饼图/环图/蜡烛图、图例与数据表](../../ui/kit/chart.go) | 缺 RadarChart、SankeyChart；已有折线/柱/面积/饼环/蜡烛图。轴域、刻度数量、参考线、线型和 tooltip 内容的公共配置较少。 |
| [Checkbox](https://gpui-kit.com/component/checkbox/) | 主体已有 | [布尔选择、半选、回调、禁用](../../ui/kit/checkbox.go) | 主体覆盖，另有半选。缺组件级尺寸配置和 GPUI 的 tab_index/tab_stop 配置入口。 |
| [Clipboard](https://gpui-kit.com/component/clipboard/) | 主体已有 | [通用复制按钮、提示与连续复制反馈](../../ui/kit/copy_button.go) | 第五批已补齐 OnCopied、Content 与反馈状态查询；回调表示已提交写入请求，非操作系统成功确认。此表登记缺口已关闭。 |
| [Collapsible](https://gpui-kit.com/component/collapsible/) | 主体已有 | [独立 Trigger/Content、动画、焦点恢复](../../ui/kit/collapsible.go) | 主体覆盖：拆分 Trigger/Content、状态控制与动画；本轮未发现新的主要功能缺口。 |
| [ColorPicker](https://gpui-kit.com/component/color-picker/) | 主体已有 | [HSV、透明度、HEX、预设、键盘、禁用](../../ui/kit/color_picker.go) | 颜色编辑主体已有；GPUI 自带触发器/弹层，Keel 是内联选择器，弹层需组合 Popover；缺触发图标、标签与尺寸配置。 |
| [Combobox](https://gpui-kit.com/component/combobox/) | 部分 | [过滤、多选标签、异步结果、重试、虚拟化](../../ui/kit/combobox.go) | 缺分组、单项禁用、自定义行/触发器、footer；目前候选数据是 string 列表。多选与异步搜索已完成。 |
| [Command](https://gpui-kit.com/component/command/) | 部分 | [模糊过滤、分组、快捷键、异步结果、虚拟化](../../ui/kit/command.go) | 缺内联模式、关闭搜索的模式、自定义行/header/footer；当前固定为带搜索的模态命令面板。 |
| [DataTable](https://gpui-kit.com/component/data-table/) | 部分 | [横向滚动、冻结列、列管理、多选/单元格选择、复制、筛选、分页加载](../../ui/kit/table.go) | 主要数据表能力已有；缺独立整列选择模式、列级 selectable/resizable/movable 限制，以及 stripe/密度等公开配置。 |
| [DatePicker](https://gpui-kit.com/component/date-picker/) | 部分 | [日历弹层、范围、多月、取消草稿、键盘](../../ui/kit/date_picker.go) | 缺日期+时间联动、快捷日期/范围预设、组件级 date_format 与清空按钮；已有独立 TimeField 不等于 DatePicker 已集成。 |
| [DescriptionList](https://gpui-kit.com/component/description-list/) | 主体已有 | [多列/跨列、横纵标签、富值插槽、分隔线、边框、字号与标签宽度](../../ui/kit/description_list.go) | 第八批已关闭登记缺口。Columns 由调用方设置，不按窗口宽度自动切换；默认仍为无边框单列，保留原用法。 |
| [Dialog](https://gpui-kit.com/component/dialog/) | 主体已有 | [可组合内容、嵌套浮层、长内容、焦点约束与恢复](../../ui/kit/dialog.go) | 主体覆盖；缺独立遮罩显示/Esc/关闭按钮开关。Body/Footer 可组合，但非 GPUI 的完整 compound parts API。 |
| [Dock](https://gpui-kit.com/component/dock/) | 部分 | [边缘与中心区标签组、嵌套分割、拖放、布局保存、最大化、跨窗口分离](../../ui/kit/dock.go) | 中心/边缘嵌套分割、拖放、最大化已完成；缺 GPUI 的面板工厂注册/面板自有状态恢复和独立 DockSkin。分离由 OnDetach 交给应用开窗，恢复布局不会重开分离窗口。 |
| [DropdownButton](https://gpui-kit.com/component/dropdown_button/) | 主体已有 | [按钮菜单、分体按钮、键盘与焦点恢复](../../ui/kit/dropdown_button.go) | 主体覆盖，另有分体动作；缺公开 anchor、loading 和内部按钮配置透传。 |
| [Editor](https://gpui-kit.com/component/editor/) | 部分 | [行号、局部重高亮、多光标/矩形选择、查找替换、折叠、语法感知括号配对、诊断/补全/悬停/定义跳转接口](../../ui/kit/code_editor.go) | 缺可随编辑跟踪的文本/几何装饰集合、可替换语言编辑规则、完整自定义搜索会话 API；高亮为 chroma，非 Tree-sitter。多光标、查找替换、折叠、括号配对已完成。 |
| [Empty](https://gpui-kit.com/component/empty/) | 主体已有 | [空状态富标题/描述、操作、媒体、尾部与分区样式](../../ui/kit/empty.go) | 第二十一、二十二批已补齐登记的媒体、富内容、尾部和样式缺口；默认保留 Surface 背景，分区样式通过 PartStyle 调整。使用单个 View 槽组合多个子项，非上游独立部件类型。 |
| [Focus Trap](https://gpui-kit.com/component/focus-trap/) | 主体已有 | [弹层焦点循环、关闭后返回焦点](../../ui/el/overlay.go) | 弹层通过 el.Layer.TrapFocus/Modal 覆盖；GPUI 还可在普通容器上独立包裹 FocusTrap，Keel 当前入口绑定浮层。 |
| [Form](https://gpui-kit.com/component/form/) | 部分 | [字段组织、校验、错误聚焦、异步提交/取消](../../ui/kit/form.go) | 缺多列网格、字段 col_span/col_start、字段描述/必填标识/可见性的声明式配置；已有校验、聚焦与异步提交状态。 |
| [GroupBox](https://gpui-kit.com/component/group-box/) | 主体已有 | [标题、描述、内容分组、四种外观、框外 footer 与标题/正文样式](../../ui/kit/group_box.go) | 第十六批已关闭登记缺口；默认保留 Keel 原有背景加边框，GroupBoxNormal 对应无装饰。样式回调作用于每帧新建元素，不应保留元素引用。 |
| [HoverCard](https://gpui-kit.com/component/hover-card/) | 主体已有 | [悬停卡片、延迟、定位、跨目标与取消](../../ui/kit/hover_card.go) | 第三批已补齐实例开关延时、方向/对齐及间距配置；默认仍为 700/300ms，键盘焦点立即打开，边缘避让保留。此表登记缺口已关闭。 |
| [Icon](https://gpui-kit.com/component/icon/) | 主体已有 | [内置矢量图标、自定义图标、尺寸与颜色](../../ui/kit/icon.go) | 实现路线不同：Keel 用 Gio/IconVG 图标；GPUI 文档提供 SVG 路径/字节与旋转接口。Keel 缺直接 SVG 加载和组件级旋转。 |
| [Image](https://gpui-kit.com/component/image/) | 主体已有 | [已解码图片、适配/裁剪/拉伸、圆角、预览、失败重试](../../ui/kit/image.go) | 已解码图片的绘制/适配/预览/重试已有；缺自定义 loading/fallback 槽和组件级 URL 加载/缓存。第七批复核更正：之前把应用/Markdown 的加载缓存算到了 kit.Image。 |
| [Input Group](https://gpui-kit.com/component/input-group/) | 主体已有 | [四方向/多附加内容、TextArea 组合、统一边框、标签聚焦与按钮操作](../../ui/kit/input_group.go) | 第九批已补齐 block addon 和独立附加内容配置；按钮直接使用 kit.Button。Textarea 最大行数/Token 与 Button 变体差异仍见各自条目，不计作已完成。 |
| [Input](https://gpui-kit.com/component/input/) | 部分 | [单行、密码、长度、前后缀、清空、校验、禁用、标签聚焦](../../ui/kit/input.go) | 缺格式化 mask、原子 inline token、可拦截富剪贴板的 on_paste 和专用上下文菜单配置；Filter 是字符白名单，不能当作 mask。 |
| [Kbd](https://gpui-kit.com/component/kbd/) | 主体已有 | [平台键帽、Plain、KbdFor 动作键位](../../ui/kit/kbd.go) | 平台键帽、Plain、KbdFor 动作绑定已完成；没有独立尺寸接口，旧表的“尺寸”应删除。 |
| [Label](https://gpui-kit.com/component/label/) | 部分 | [Text、字号/颜色、For 标签关联聚焦](../../ui/el/element.go) | el.Text 能排版和关联字段；缺 GPUI Label 的匹配区间高亮、masked 和 secondary 文案的专用接口。 |
| [List](https://gpui-kit.com/component/list/) | 部分 | [列表项、稳定 ID、单项禁用、多选/范围、键盘、拖动](../../ui/kit/list.go) | 缺分组头、自定义行/图标/行内操作、内建搜索与加载更多入口；当前是可选择、可拖动的文字列表。 |
| [Marker](https://gpui-kit.com/component/marker/) | 用途不同 | [几何标记、大小与颜色](../../ui/kit/marker.go) | 用途不同：GPUI 是带图标/文字、分隔线/边框、加载状态的消息标记行；Keel Marker 只绘制点/方块等几何标记，不能计作对齐。 |
| [Menu](https://gpui-kit.com/component/menu/) | 部分 | [菜单、子菜单、分隔线、长内容、方向键、焦点恢复](../../ui/kit/menu.go) | 缺图标/勾选项、组标题、任意自定义行与链接项 API；已有子菜单、快捷键、禁用、滚动和焦点行为。 |
| [MessageScroller](https://gpui-kit.com/component/message-scroller/) | 部分 | [可变高度虚拟化、跟随尾部、流式增高、历史加载锚点](../../ui/kit/message_scroller.go) | 虚拟化、尾部跟随、历史锚点与“最新”按钮已有；缺公开按消息跳转/初始未读定位、跟随状态查询和自定义跳转按钮。 |
| [Message](https://gpui-kit.com/component/message/) | 部分 | [消息内容、状态、操作栏、反应、失败重试](../../ui/kit/message.go) | 缺独立 avatar/header/footer 插槽、MessageGroup 和 ghost/content_inset 配置；当前是作者文字+内容+操作/反应。 |
| [Notification](https://gpui-kit.com/component/notification/) | 部分 | [通知队列、超时、关闭、暂停与原位更新](../../ui/kit/notification.go) | 缺系统通知投递、位置选择、操作按钮与任意富内容；当前只有应用内右上角标题/正文通知队列。 |
| [NumberInput](https://gpui-kit.com/component/number-input/) | 部分 | [数值解析、范围/步长/精度、草稿提交与取消](../../ui/kit/number_input.go) | 缺金额/千分位 mask、动态 step_by、前后内容槽；固定步长、精度、范围与输入草稿已有。 |
| [OtpInput](https://gpui-kit.com/component/otp-input/) | 主体已有 | [分格输入、粘贴、完成回调、密码遮罩、分组、尺寸与窄布局](../../ui/kit/otp_input.go) | 第二批已补齐 Masked/Groups/Size；默认两组，不能整除时前组多一位。此表登记的三个缺口已关闭。 |
| [Pagination](https://gpui-kit.com/component/pagination/) | 主体已有 | [页码、前后翻页、总数、窄布局换行](../../ui/kit/pagination.go) | 主体覆盖；缺 compact、visible_pages 与尺寸档接口。 |
| [Plot](https://gpui-kit.com/component/plot/) | 用途不同 | [成品散点/折线图、缩放、平移、拾取](../../ui/kit/plot.go) | 用途不同：GPUI 提供 ScaleLinear/Band/Point/Ordinal、Bar/Line/Area/Pie/Stack/Axis 等公共绘图基础件；Keel Plot 是可缩放平移的成品散点/折线图。 |
| [Popover](https://gpui-kit.com/component/popover/) | 主体已有 | [锚点定位、避让、长内容、外部点击/Esc、焦点恢复](../../ui/kit/popover.go) | 主体覆盖；缺箭头、实例 offset 与 mouse_button 配置；低层 Anchored 可设 Offset。 |
| [Progress](https://gpui-kit.com/component/progress/) | 主体已有 | [条形确定/不确定进度](../../ui/kit/progress.go)、[圆形进度与中心内容](../../ui/kit/progress_circle.go) | 第一批已补齐 ProgressCircle：真实进度、加载动画、减少动画、中心内容、大小/颜色。条形组件的高度/颜色等样式配置仍比 GPUI 少。 |
| [Questionnaire](https://gpui-kit.com/component/questionnaire/) | 部分 | [题型、答案模型、校验、分页、禁用与提交快照](../../ui/kit/questionnaire.go) | 缺单题条件禁用、跳过状态、自定义/外部校验、同题选项+自由输入、完整进度状态和快捷键配置；现有五种题型、必填校验与分页保留。 |
| [Radio](https://gpui-kit.com/component/radio/) | 主体已有 | [单选组、横纵布局、独立 Item、单项禁用、键盘](../../ui/kit/radio_group.go) | 单选主体覆盖；组内 Item 可单独放置。缺任意富标签和组件级大小配置。 |
| [Rating](https://gpui-kit.com/component/rating/) | 主体已有 | [评分、半星/小数展示、只读、键盘](../../ui/kit/rating.go) | 行为差异：GPUI 再点已填星会减分，Keel 直接设为该星序号；缺大小/颜色配置。Keel 另支持小数展示，交互仍是整星。 |
| [Resizable](https://gpui-kit.com/component/resizable/) | 部分 | [横纵分割、最小尺寸、拖动、键盘、取消与禁用](../../ui/kit/resizable.go) | 缺独立多面板 group、最大尺寸、条件显隐/把手外观配置；Keel 为双面板，可嵌套组合更多面板。 |
| [Root View](https://gpui-kit.com/component/root/) | 主体已有 | [根布局、统一浮层宿主、窗口快捷键](../../ui/el/root.go) | 架构差异：Keel 已有 root/overlay/focus/shortcut；Dialog/Sheet/Notifier 需应用挂载，GPUI 0.7 根视图自动挂载这些层。 |
| [Scrollable](https://gpui-kit.com/component/scrollable/) | 主体已有 | [ScrollX/ScrollY、滚动条拖动/轨道点击、定位与尾部跟随](../../ui/el/viewport.go) | 双轴滚动、滚动条与定位已有，通过 el 组合；缺组件级 Always/Hover/Scrolling 显示策略。没有独立类型本身不计功能缺失。 |
| [Select](https://gpui-kit.com/component/select/) | 主体已有 | [过滤、分组、多选、禁用项、万条虚拟化](../../ui/kit/select.go) | 单选主体覆盖，另有多选；缺自定义行/空内容/标题前缀、清空按钮与菜单宽高配置。分组、禁用项已实现。 |
| [Settings](https://gpui-kit.com/component/settings/) | 部分 | [设置分组、导航、搜索、窄布局](../../ui/kit/settings.go) | 缺页面下的多 Group 模型、resettable 重置、组 footer、独立搜索 keywords 和 Markdown 描述；已有分区导航、搜索与窄布局。 |
| [Sheet](https://gpui-kit.com/component/sheet/) | 部分 | [侧边抽屉、遮罩、长内容、焦点与禁用继承](../../ui/kit/sheet.go) | 缺拖动调整尺寸、独立 footer、顶部 margin、遮罩显示/点击关闭配置；四方向抽屉主体已有。 |
| [Shimmer](https://gpui-kit.com/component/shimmer/) | 用途不同 | [Skeleton 占位块扫光、减少动画](../../ui/kit/skeleton.go) | 用途不同：GPUI ShimmerText 保留可读文字并让高光扫过文字；Keel Skeleton.Shimmer 只扫过占位几何。缺文字效果与 duration/spread/reverse/once 配置。 |
| [Sidebar](https://gpui-kit.com/component/sidebar/) | 主体已有 | [嵌套分组、收起、选中、固定头尾、键盘滚动](../../ui/kit/sidebar.go) | 主体覆盖；缺右侧布局开关、自定义 item suffix/上下文菜单接口。已有 Badge 和固定 Header/Footer。 |
| [Skeleton](https://gpui-kit.com/component/skeleton/) | 主体已有 | [占位形状、尺寸、次级色阶、自定义圆角与加载动画](../../ui/kit/skeleton.go) | 第二十三批已关闭登记缺口；保留 Keel 的 1.5 秒明暗脉冲/可选扫光及减少动画，默认颜色来自 Subtle/SubtleHover，与上游独立 skeleton token、2 秒透明度动画不同。 |
| [Slider](https://gpui-kit.com/component/slider/) | 主体已有 | [单值/双端范围、横纵向、线性/对数、步长、拖动/键盘与结束回调](../../ui/kit/slider.go) | 第十批已关闭对数刻度和 Release 缺口；无效对数范围回退线性，取消不回滚已有值。轨道/滑块颜色与大小仍使用统一样式，未提供逐项外观配置。 |
| [Spinner](https://gpui-kit.com/component/spinner/) | 主体已有 | [不确定动画、减少动画、可访问名称、自定义图标/颜色/周期](../../ui/kit/spinner.go) | 第十九、二十批已补齐图标、颜色和速度配置；Period 为每周时长，默认一秒匀速。上游描述的默认 0.8 秒及缓动曲线不同。 |
| [StatusBar](https://gpui-kit.com/component/status-bar/) | 主体已有 | [固定状态栏、左右内容组、按优先级收起的溢出菜单](../../ui/kit/status_bar.go) | 左右内容与自定义 View 已覆盖；Keel 另有优先级溢出菜单，本轮未发现新的主要功能缺口。 |
| [Stepper](https://gpui-kit.com/component/stepper/) | 主体已有 | [横纵步骤、图标/富内容、尺寸、导航、键盘、滚动与单步禁用](../../ui/kit/stepper.go) | 第六批已补齐 Vertical、Size、StepperItem 与 SetItemDisabled。此表登记缺口已关闭；导航仍限已完成步骤，GPUI 文档的文本居中布局未提供独立开关。 |
| [Switch](https://gpui-kit.com/component/switch/) | 主体已有 | [布尔开关、标签、禁用与键盘](../../ui/kit/switch.go) | 布尔开关主体已有；缺大小/颜色/标签侧配置。当前无 Loading 接口，旧表误记；GPUI 此页也未将 loading 列为能力。 |
| [Table](https://gpui-kit.com/component/table/) | 部分 | [排序、行选择、单元格插槽、列宽调整；高级能力同 DataTable](../../ui/kit/table.go) | GPUI Table 是轻量 Header/Body/Footer/Caption 组合表，DataTable 才负责数据交互；Keel 两项共用 TableView，缺独立 footer/caption/任意行组合。 |
| [Tabs](https://gpui-kit.com/component/tabs/) | 主体已有 | [四种外观、图标/富标签、单项禁用、页面状态、溢出、关闭与焦点恢复、拖动排序](../../ui/kit/tabs.go) | 第十二、十三批已关闭登记的外观、禁用、内容、最大宽度及滚动接口缺口。默认仍为溢出菜单；Scrollable 开启时改为滚动轨道，ScrollTo 只定位不选择。自定义标签应为展示内容，宽度上限不包含独立关闭按钮。 |
| [Tag](https://gpui-kit.com/component/tag/) | 主体已有 | [语义/自定义颜色、描边、圆角、尺寸、富内容、移除与选中](../../ui/kit/tag.go) | 第十四批已关闭登记缺口；默认保留主题染色胶囊，实心底色可用 Appearance。Size 为最小高度，长文字仍换行；自定义内容限展示元素。 |
| [TextView](https://gpui-kit.com/component/text-view/) | 部分 | [Markdown、HTML 富文本、扩展 TeX、图片、选择复制、代码块、流式渲染](../../ui/markdown) | 缺富文本整体 max_lines/is_clamped、逐流式增量淡入、公开区间高亮/跳转、Markdown 插件与代码块操作扩展接口。HTML/扩展 TeX 已完成。 |
| [Textarea](https://gpui-kit.com/component/textarea/) | 部分 | [多行、只读、Rows 最小高度、错误显示](../../ui/kit/input.go) | 缺 inline token 和 auto_grow(min,max) 的最大行数控制；Rows 只设最小高度，旧表“最大可见行数”不成立。 |
| [Theme](https://gpui-kit.com/component/theme/) | 主体已有 | [语义配色、间距/字号/圆角/阴影刻度、浅深切换、注册与 JSON 主题、局部作用域、渐变、目录监听](../../ui/theme/registry.go) | 核心主题机制已完成；Keel 7 套内置（含 light/dark），GPUI 文档称 20+。Keel 渐变 JSON 为 from/to/angle，仅 Bg/Primary；GPUI 是可选背景 token 的 CSS 两色渐变，配置不兼容。 |
| [TimeField](https://gpui-kit.com/component/time-field/) | 主体已有 | [时分秒、步进/进位、12/24 小时、Tab 与本地化](../../ui/kit/time_field.go) | 分段、时分秒、12/24 小时、键盘修改已完成；缺组件级尺寸档。本轮未重做真机键盘验收。 |
| [TitleBar](https://gpui-kit.com/component/title-bar/) | 主体已有 | [自定义标题栏、窗口控制、macOS 双击偏好与失焦外观](../../ui/kit/title_bar.go) | 自绘标题栏与窗口控制已实现；macOS 窗口初始居中已实测。标题栏全部系统行为及 Windows/Linux 真机验收仍待完成。 |
| [Toggle](https://gpui-kit.com/component/toggle/) | 主体已有 | [状态按钮、图标、单选/多选组](../../ui/kit/toggle_group.go) | 单个开关按钮/单多选组已有；缺 ghost/outline/segmented 外观和大小档，旧表“尺寸”不成立。 |
| [Toolbar](https://gpui-kit.com/component/toolbar/) | 主体已有 | [左右区域、尺寸、工具分组、溢出与键盘](../../ui/kit/toolbar.go) | 主体覆盖；命令用 ToolbarItem，自定义内容用 Leading/Trailing，缺任意位置插入 compound 自定义组的接口。 |
| [Tooltip](https://gpui-kit.com/component/tooltip/) | 主体已有 | [通用提示、键盘焦点、延迟与取消](../../ui/kit/tooltip.go) | 第四批已补齐 Content、Action、Placement/Offset；动作键位自动跟随改绑，富内容不可交互。此表登记缺口已关闭。 |
| [Tree](https://gpui-kit.com/component/tree/) | 部分 | [虚拟化、展开、多选、单项禁用、键盘、动态数据与拖动](../../ui/kit/tree.go) | 缺公开行渲染器（图标/操作）和逐节点动态子项更新/展开加载回调；可整树 SetRoots。拖动自动滚动/展开仍缺，但不把它当作本页已证实的 GPUI 差距。 |
| [VirtualList](https://gpui-kit.com/component/virtual-list/) | 部分 | [等高及可变高度实现、稳定 key、尺寸缓存、插入保持锚点](../../ui/kit/variable_list.go) | 纵向等高/变高、尺寸缓存、锚点已有；缺横向虚拟列表与虚拟化轴切换。普通横向滚动不等于横向虚拟化。 |

验证记录见 [组件验收](component-acceptance-2026-10-02.md)。

## 跨文件复核线索

- 选择与异步：`combobox_options.go`、`command_search.go`、`select_options.go`、`list_items.go`、`tree_items.go`，均位于 `ui/kit`；因此多选/异步等没有被误判为未实现。
- 数据表：`ui/kit/table_{cells,columns,data,menu,selection}.go`；图表另读 `pie_chart.go`、`candlestick_chart.go`。
- 编辑器：`ui/kit/code_search.go`（公开搜索入口和 10,000 个匹配上限）、`code_editor_input.go`（内置配对表）、`code_fold.go`、`code_highlight.go`。搜索/替换存在，但自定义搜索控制接口不等同于 GPUI SearchSession。
- Dock：`ui/kit/dock_tree.go`、`dock_drag.go`、`dock_detach.go`；主题：`ui/theme/registry.go`、`watch.go`、`themes/*.json`。
- 文本与底层：`ui/markdown/markdown.go`、`html.go`、`render.go`、`ui/el/element.go`、`overlay.go`、`viewport.go`；头尾、分段等扩展见 `sidebar_items.go`、`tabs_reorder.go`、`time_field_segments.go`。

初次复核只更新比较报告；后续组件改动与验证按上方批次记录，未重做性能跑分。性能、视觉和系统读屏需独立验收，不能由上述组件覆盖数推导。
