# GPUI Kit 实现进度 · 2026-10-02

更新日期：2026-10-03（原报告 2026-10-02，代码基准 `2fe8d1d`，本轮逐页复核 77 项公开文档及 Keel 公共接口/相关实现）。来源：[GPUI Kit 组件目录](https://gpui-kit.com/component/)（页面版本 v0.7.0），按导航中的独立组件链接去重，共 **77 项**。组件分类参考该站，说明和实现判断根据 Keel 当前工作区重写；源站文档采用 [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)。这是一份能力对照，不要求复制 Rust API。

补齐后状态：**48 项主体已有、27 项部分覆盖、2 项用途不同**（初次复核为 38/36/3）。这是按文档列出的功能判断，不是功能完成百分比，也不是视觉成熟度评分。

- **主体已有**：核心用途覆盖；末列仍列出配置、交互或组合方式差异，不能读成全部功能相同。
- **部分**：已有可运行的主体，但缺源站明确提供的扩展功能或组合能力；已实现部分继续记为完成。
- **用途不同**：旧表拿同名或近似效果代替了不同用途的组件，需重新建立对应关系。

本轮是文档与源码静态复核，检查了跨文件扩展接口；没有重跑 77 项真机交互，也没有逐项证明视觉一致。缺口指组件当前未提供的接口/行为，应用通过 `el` 自行组合不自动算作 kit 已封装。GPUI 文档示例不等于上游源码和所有平台均已验证。

## 本轮更正与重点差距

1. 初次复核更正 Marker、Plot、Shimmer 的用途混淆；第六十五批新增 ShimmerText 后，文字扫光已建立独立实现，Marker 和 Plot 的用途差异仍保留。
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
- [x] 第二十三批（`41a9982`）：Skeleton 增加 Secondary 与 Rounded，整体透明度减半、任意有限非负圆角并按短边限制；Circle/Rounded 后调用者生效。像素测试覆盖浅深主题、直角/圆角/极大圆角、普通/次级的脉冲与扫光、减少动画及装饰语义；全量构建、vet、测试和浅色 1×/深色 2× 截图通过。
- [x] 第二十四批（`260995a`）：Rating 增加 Size/Color，并按上游 0.7 源码对齐已填星点击：点第 i 颗已填星设置 i−1 分，否则设置 i 分；悬停预览对应目标。覆盖 1×/2× 尺寸、清零、键盘、只读/禁用继承、Agent 数值及自定义颜色/小数填充像素；全量构建、vet、测试及浅深色截图检查通过。
- [x] 第二十五批（`20e5731`）：Kbd 增加 Size 与 Style；独立字号同步缩放内边距，0 恢复继承，样式回调可调颜色/背景/边框。覆盖 1×/2×、Plain 尺寸、样式恢复、动作改绑/解绑；全量构建、vet、测试及浅深色截图检查通过。
- [x] 第二十六批（`de57b50`）：条形 Progress 增加 Height、Color、Rounded 与 TrackStyle；自定义颜色覆盖主题渐变，轨道支持背景和边框，进度块填满内部高度。同时修正独立渲染的轨道宽度和 NaN 值。像素/Agent 测试覆盖填充比例、配色、高度、动画与减少动画；全量构建、vet、测试及浅深色截图检查通过。
- [x] 第二十七批（`f122562`）：条形/圆形 Progress 加入 200ms 数值过渡，连续更新从当前显示值衔接；首次显示、退出不确定模式和减少动画立即显示目标。复用展开组件的插值逻辑，结束时精确归位。帧时钟像素测试覆盖正向/反向更新、重定向连续性、立即目标语义和静止；展开组件回归及全量构建、vet、测试通过。
- [x] 第二十八批（`47cae5a`）：Popover 增加 Offset，支持默认 4dp、零间距、负值重叠及打开期间重新定位；忽略非有限值。四方向布局、动态更新、回调次数及 Esc 关闭通过测试；全量构建、vet、测试通过。箭头、鼠标触发配置和面板样式仍待补齐。
- [x] 第二十九批（`2182b86`）：Popover 增加 Appearance 与 PanelStyle，默认装饰可关闭，面板可配置配色、边框、圆角和间距。保留稳定身份、dialog 语义、窗口约束与滚动；1×/2× 测试覆盖样式恢复、输入状态与焦点，像素测试覆盖配色与重置；全量构建、vet、测试通过。
- [x] 第三十批（`3acdc4b`）：Popover 增加 RightClick，右键切换开关，保留触发元素左键/键盘行为；可动态关闭右键处理。测试覆盖重复切换、回调次数、Esc、组件及祖先禁用；全量构建、vet、测试通过。任意鼠标按键选择与箭头仍待完成。
- [x] 第三十一批（`8c16322`）：Popover 增加 MouseButton，支持左/右/中键与手动模式；RightClick 沿用兼容入口。底层新增 OnMousePress，保留子元素事件和禁用继承。测试覆盖三种按键的交叉过滤、重复开关、动态恢复手动模式、非法值、多键同时按下和禁用；全量构建、vet、测试通过。
- [x] 第三十二批（`7468891`）：Popover 增加 Arrow，箭头跟随实际方向翻转，按对齐方式定位并避开圆角；Offset 测量到尖端。箭头独立于内容滚动，命中区域不会触发外部关闭；几何/交互测试覆盖翻转、窄面板、箭头点击，像素测试覆盖四方向、自定义颜色与关闭；全量构建、vet、测试通过。箭头使用纯色背景，不单独绘制边框/阴影或采样渐变。
- [x] 第三十三批（`ef0db2c`）：Pagination 增加 Compact、VisiblePages、Size 和 SetDisabled；稳定按钮身份保留切换模式后的焦点。测试覆盖页码预算/省略号、最大整数、1×/2× 尺寸、紧凑导航、键盘、禁用继承和总数收敛；默认页码策略保持兼容。全量构建、vet、测试及浅色 1×/深色 2× 截图检查通过。
- [x] 第三十四批（`d70fc17`）：Switch 增加 Small/Medium 尺寸、Color/ClearColor 和 LabelSide；自定义选中色在自身禁用时降低 alpha，切换标签位置保留键盘焦点。测试覆盖 1×/2× 布局、标签点击、键盘、禁用继承、回调与颜色像素；全量构建、vet、测试及浅色 1×/深色 2× 截图检查通过。重新核对上游后，将滑块动画、焦点环及 Tab 配置补记为未完成。
- [x] 第三十五批（`bd8826a`）：Switch 增加 180ms 滑块过渡，快速反向从当前显示位置衔接；首次显示/减少动画直接归位，值和语义立即更新。帧时钟像素测试覆盖大小两档、开始/中间位置、反向衔接、最终归位和减少动画；全量构建、vet、测试通过。
- [x] 第三十六批（`98b70ba`）：Switch 增加 FocusRing，关闭/恢复焦点轮廓不改变尺寸与键盘操作。窗口像素测试覆盖 Tab 聚焦、隐藏/恢复、Space/Enter 切换和禁用；全量构建、vet、测试通过。轮廓仍沿整行绘制，轨道级轮廓与 Tab 配置继续记录为差异。
- [x] 第三十七批（`b13354e`）：el 和 Switch 增加 TabStop/TabIndex；显式配置启用单 root 顺序遍历，负索引或关闭停靠时保留鼠标/程序聚焦。窗口测试覆盖升序/同值树序、输入框、隐藏/禁用、动态配置、Tab/Shift+Tab 与模态初始焦点及循环；全量构建、vet、测试通过。跨独立 Embed/原生 Gio 排序及直接 Router.MoveFocus 不在此入口范围。
- [x] 第三十八批（`0decad0`）：Checkbox 增加 Size/TextSize 与 TabStop/TabIndex；勾号和半选横线随方框缩放。测试覆盖 1×/2×、尺寸恢复、半选语义、键盘焦点、禁用、Tab 排序/跳过与反向遍历；全量构建、vet、测试及浅色 1×/深色 2× 截图检查通过。
- [x] 第三十九批（`cf93625`）：RadioGroup 增加 Size/TextSize 与按选项配置的 Content；圆环和选中点按比例缩放，富标签保留原值、名称和键盘身份，独立 Item 共享配置。测试覆盖 1×/2×、替换/恢复后的焦点、描述点击、禁用跳过、祖先禁用、重排及删除清理；全量构建、vet、测试和浅色 1×/深色 2× 截图检查通过。逐项尺寸与组件级 Tab 配置仍待补齐。
- [x] 第四十批（`65ffdfe`）：RadioGroup 增加 ItemSize，支持逐项圆点/字号覆盖和分别继承组配置；重排保留，删除清理。1×/2× 测试覆盖尺寸隔离、恢复继承、动态组字号、焦点、非法值及生命周期；构建、vet、全量测试和浅深色截图检查通过。组件级 Tab 配置仍待补齐。
- [x] 第四十一批（`14defb2`）：RadioGroup 增加 TabStop/TabIndex 和逐项 ItemTab/ClearItemTab；默认单停靠点，显式逐项配置可覆盖，删除选项清理覆盖。窗口测试覆盖正反向、组跳过、负索引、鼠标选择、方向键、未选中项停靠且不改变值、禁用与配置清理；构建、vet、全量测试通过。排序范围限单 el root。
- [x] 第四十二批（`ea1ec2d`）：Dialog 增加 Keyboard、Overlay、OverlayClosable、CloseButton，适用于普通及警告对话框；隐藏遮罩保持模态，禁止 Esc 不会穿透关闭下层。关闭按钮默认隐藏并使用本地化名称；正文/页脚身份稳定。测试覆盖独立关闭路径、嵌套、遮罩像素与焦点返回后再打开；构建、vet、全量测试通过。确认回调阻止关闭仍待补齐。
- [x] 第四十三批（`3481982`）：Dialog 增加 BeforeConfirm，标准确认/危险确认/提示可同步拒绝确认，保留打开状态与焦点并跳过原 onOK；nil 清除，复用保留。回调中显式改变打开状态、禁用或替换消息时中止旧确认。测试覆盖回车重试、程序重开、点击、取消隔离及消息替换；构建、vet、全量测试通过。取消路径的可拒绝回调仍待完成。
- [x] 第四十四批（`ee1e9d2`）：Dialog 增加 BeforeCancel，统一校验 Esc、外部点击、取消及关闭按钮；底层 BeforeDismiss 在接受后才标记浮层已关闭。拒绝保留焦点和模态，所属元素失效清理绕过校验。测试覆盖四路径重复拒绝/重试、回调次数、清理、消息替换和键盘焦点；构建、vet、全量测试通过。OnClose 仍沿用仅取消通知的兼容语义。
- [x] 第四十五批（`e93c1a5`）：DropdownButton 增加 Button 配置透传、共享 Size 与主操作 Loading；未指定共享变体/尺寸时继承内层按钮，渲染不修改源实例，内部身份稳定。分体箭头可在主操作加载时使用。测试覆盖主操作回调、加载与键盘焦点、禁用、清除内层按钮和配置隔离；构建、vet、全量测试及浅色 1×/深色 2× 截图检查通过。菜单定位配置仍待完成。
- [x] 第四十六批（`0b5f614`）：Menu/DropdownButton 增加 Placement/Offset，四方向、三种对齐和有限正负间距，打开期间可重新定位；普通按钮锚定整体，分体按钮锚定箭头，子菜单保留原策略。测试覆盖方向/对齐/间距组合、非法值、边缘翻转限制和 Esc 后焦点恢复；构建、vet、全量测试通过。
- [x] 第四十七批（`e7558d0`）：Menu 增加 IconItem、CheckItem、图标/勾选状态更新与查询、CheckSide；统一前置标记列，勾选后关闭菜单链再通知应用。Agent 增加 menuitemcheckbox 的 checked 状态。测试覆盖回车/点击、禁用、程序赋值、重新打开、子菜单关闭链及 Agent 状态；构建、vet、全量测试通过。组标题、自定义行及链接项仍待完成。
- [x] 第四十八批（`c756970`）：Menu 增加 Label 分组标题，空标题忽略，单行截断，不参与点击、方向键和文字搜索；长菜单定位计入标题行高。测试覆盖 heading 语义、点击隔离、键盘/搜索跳过以及 30 组菜单的 End 定位与执行；构建、vet、全量测试通过。自定义行和链接项仍待补齐。
- [x] 第四十九批（`e284361`）：Menu 增加 ContentItem/SetItemContent，多行展示内容按实际高度布局，保留可访问名称、搜索、禁用及原操作；滚动定位改为测量实际行高并在可见后转移焦点。测试覆盖内容点击、替换恢复、禁用以及 1×/2× 变高菜单的 Home/End 定位与执行；构建、vet、全量测试通过。链接项仍待补齐。
- [x] 第五十批（`c1f89c9`）：Menu 增加 Link、ExternalLinkIcon、OnLink/OnLinkError 及子菜单回调继承，默认通过 core.OpenURL 打开 HTTP/HTTPS/mailto；URL 写入语义值。测试覆盖键盘/点击、禁用、菜单关闭、回调覆盖及非法地址；构建、vet、全量测试和 core 的 Windows/Linux/wasm 交叉构建通过。未执行系统浏览器真机打开验收；快捷键焦点上下文解析仍待补齐。
- [x] 第五十一批（`6d6c3c5`）：Sheet 增加独立 Footer 和 Keyboard/Overlay/OverlayClosable/CloseButton，正文独立滚动，页脚复制切片并支持换行。测试覆盖四方向长正文下页脚可见/执行、关闭按钮动态恢复、Esc 配置和切片隔离；构建、vet、全量测试通过。仍缺拖动调整尺寸和顶部间距；页脚需保留足够面板高度。
- [x] 第五十二批（`5ce5bde`）：Sheet 增加 MarginTop，四方向按剩余高度布局，顶部滑入绘制/点击区域同步裁剪，遮罩保持全窗口。测试覆盖 1×/2×、打开期间调整/清零、空间不足、非有限值、顶部点击和动画；构建、vet、全量测试通过。重新查阅上游后补记面板自定义配色/间距接口差异。
- [x] 第五十三批（`893fd76`）：Sheet 增加 PanelStyle，支持配色、边框、圆角、阴影、内边距与间距，nil 恢复默认；尺寸和语义身份由组件保持。测试覆盖 1×/2× 布局、动态样式与输入焦点/内容保留、页脚交互及背景/边框像素；构建、vet、全量测试通过。拖动调整尺寸仍待完成。
- [x] 第五十四批（`86d193a`）：Sheet 增加默认开启的 Resizable、PanelSize 和 OnResize，四方向内侧把手、键盘调整、窗口约束、取消恢复与关闭/禁用清理。测试覆盖 1×/2× 拖动、Home/End、起始尺寸受限、取消、祖先禁用及中途关闭调整；构建、vet、全量测试通过。Sheet 升为主体已有，总计 47/27/3。
- [x] 第五十五批（`bfb6c4e`）：Attachment 增加 Media 与 Vertical，支持展示型图片/媒体槽、横纵布局及默认图标恢复；操作区独立并保持打开区域身份。测试覆盖 1×/2× 窄布局、预览打开、切换后键盘焦点、上传取消/重试、移除隔离和禁用；构建、vet、全量测试通过，浅深色示例截图已检查。仍缺附件组；重新查阅上游后补记生命周期、分区组合/样式、尺寸档和图片状态效果差异。
- [x] 第五十六批（`d824945`）：AttachmentGroup 增加横向排列/滚动、Gap、Name、禁用继承及复制切片的 SetItems/Items。测试覆盖 1×/2× 溢出尺寸、滚动后打开/移除命中、动态删除后的偏移收敛、其他附件状态保留、空组及非法间距；构建、vet、全量测试通过。生命周期和分区配置仍待补齐。
- [x] 第五十七批（`a7bfd35`）：Attachment 增加显式生命周期及查询，覆盖待上传/上传/处理/失败/完成并保留取消状态；中英文文案、默认媒体忙碌指示、操作限制与旧进度/错误接口统一。测试覆盖状态转换、错误清除恢复、处理中取消/重试、非法状态、双语言与 Agent 值；构建、vet、全量测试通过。分区配置、尺寸与图片状态效果仍待完成。
- [x] 第五十八批（`98de78e`）：Attachment 增加 Content、Actions 和六分区 PartStyle，默认元信息可恢复，自定义操作与打开隔离，根语义身份保持。测试覆盖 1×/2× 内容替换、切片隔离、动态样式与输入焦点/状态、内置操作保留、禁用及背景像素；构建、vet、全量测试通过。尺寸档和状态视觉效果仍待补齐。
- [x] 第五十九批（`8405714`）：Attachment 增加 XSmall/Small/Medium/Large 四档，统一卡片宽度、默认媒体、字号、间距与内置按钮；默认 Medium 宽度调整为 232dp，PartStyle 优先。测试覆盖 1×/2× 尺寸递增、焦点保留、非法值、样式覆盖及窄竖排上传操作；构建、vet、全量测试及浅深色截图检查通过。补记组边缘渐隐/公开滚动控制及竖排预览/操作布局差异。
- [x] 第六十批（`c62acee`）：AttachmentGroup 增加 ScrollTo/ScrollState，支持首次绘制前定位、隐藏期间保留请求、按实际内容限制偏移及应用分页按钮。测试覆盖 1×/2×、超大/负数/非有限值、删除时待处理请求、隐藏恢复及禁用时程序滚动；构建、vet、全量测试通过。组边缘渐隐仍待完成。
- [x] 第六十一批（`7e91538`）：AttachmentGroup 增加 EdgeFade/ClearEdgeFade，按左右隐藏内容绘制 24dp 渐隐，窄窗口限制半宽并避开滚动条。像素与交互测试覆盖起点/中间/终点、移除后不溢出、关闭渐隐及渐隐区域点击；构建、vet、全量测试通过。附件状态视觉与竖排布局差异仍待补齐。
- [x] 第六十二批（`c40cc59`）：Attachment 待上传使用圆角虚线边框，失败使用危险色边框及默认媒体错误/禁止图标；新增 el.BorderDashed，样式回调可覆盖状态默认值。像素测试覆盖虚线分布、实线覆盖、失败边框及完成恢复；构建、vet、全量测试及浅深色截图检查通过。图片状态遮罩和标题扫光仍待完成。
- 后续差异继续以 77 项表中末列为准。

- [x] 第六十三批（`162518e`）：Attachment 自定义媒体状态遮罩，上传显示白色确定进度环，处理显示不确定环，失败显示加深遮罩和重试/禁止图标；完成后恢复预览。覆盖状态像素、语义、重试状态及祖先禁用测试；构建、vet、全量测试和浅色 1×/深色 2× 截图检查通过。

- [x] 第六十四批（`9b7399a`）：Attachment 增加 MediaOverlay，自定义内容绘制在媒体状态遮罩上方，不改变预览尺寸；支持独立按钮操作、禁用继承、动态清除及状态切换后的焦点保留。1×/2× 交互、遮罩上方像素、Agent 语义测试和浅深色截图检查通过；构建、vet、全量测试通过。

- [x] 第六十五批（`b73ba37`）：新增 ShimmerText 和 el 字形扫光绘制，支持 Duration/Spread/Reverse/Once/Restart、配色、字号/行数和减少动画；Attachment 上传与处理中标题接入。固定帧时间测试覆盖高光移动、反向、单次/重播、减少动画、语义去重、混排/粗体/行高/截断及附件完成后恢复；构建、vet、全量测试通过，浅色 1×/深色 2× 静态截图已检查。

- [x] 第六十六批（`8953b3e`）：Attachment 增加 Title/Description 的 PartStatus/ClearPartStatus，以及 Description/ClearDescription。覆盖只影响默认元信息展示，保留真实生命周期、进度条和重试操作；支持恢复继承、父状态更新后保留和自定义 Content 恢复。状态/重试回归、描述颜色像素与 Agent 语义测试通过；构建、vet、全量测试及浅色 1×/深色 2× 截图检查通过。

- [x] 第六十七批（`3f4306d`）：Attachment 竖排默认方形预览，新增 MediaAspectRatio（0 保留自然尺寸），已加载 Image 居中裁剪且不修改源实例；操作区移至卡片右上角，切换方向保留焦点。el 增加 AspectRatio，高度显式值及约束优先。测试覆盖 1×/2×、窄窗口、比例恢复/非法值、显式尺寸约束、源图片隔离、点击/焦点及 cover 像素；构建、vet、全量测试和浅深色截图检查通过。

- [x] 第六十八批（`e729248`）：Attachment 增加 ShowMedia/ShowContent/ShowActions；竖排隐藏元信息时使用边框内铺满的纯图片卡片，只有操作时恢复普通流布局。纯图片模式的键盘焦点轮廓绘制在图片上方，隐藏分区保留配置但不参与布局/交互。1×/2× 窄布局、分区恢复、打开/操作隔离、状态保持、Agent 隐藏及焦点轮廓像素测试通过；构建、vet、全量测试及浅深色截图检查通过。

- [x] 第六十九批：Attachment 增加 MediaSource、MediaLoading/MediaError 和 RetryMedia；后台解码、15 秒期限、取消与版本校验，支持网络、data URL 和本地来源。图片加载失败与上传失败分别管理，切换 Media 取消旧请求，加载状态不改变预览尺寸。HTTP 取消/替换、重复来源、失败重试、禁用/打开隔离、迟到结果、成功像素及 Agent 状态测试通过；构建、vet、全量测试通过，浅深色加载状态截图已检查。

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
| [AlertDialog](https://gpui-kit.com/component/alert-dialog/) | 主体已有 | [提示/确认/危险对话框、焦点约束与恢复](../../ui/kit/dialog.go) | 第四十二批已补齐 Keyboard、Overlay、OverlayClosable、CloseButton；Persistent 默认只禁止遮罩关闭，显式配置可覆盖。第四十三批增加 BeforeConfirm，可返回 false 保持打开并跳过 onOK；允许后仍先关闭再执行原回调。第四十四批增加 BeforeCancel，支持拒绝用户取消。程序关闭和所属元素失效清理绕过校验；OnClose 仍只通知取消关闭，与上游确认后也通知的约定不同。 |
| [Alert](https://gpui-kit.com/component/alert/) | 主体已有 | [行内/横幅提示、级别、四档尺寸、可替换图标、富正文、关闭按钮](../../ui/kit/alert.go) | 第十五批已关闭登记缺口；Content 可组合 Markdown 与操作按钮。横幅没有独立标题行，无正文时使用标题作为消息；自定义内容的内部样式由内容自身控制。 |
| [Attachment](https://gpui-kit.com/component/attachment/) | 部分 | [附件卡片、进度、取消、重试、错误状态](../../ui/kit/attachment.go) | 第五十五批补齐 Media 媒体槽和 Vertical 横纵布局。第五十六批补齐 AttachmentGroup 横向排列与滚动。第五十七批补齐显式生命周期和默认媒体忙碌指示。第五十八批补齐 Content/Actions 和六分区 PartStyle。第五十九批补齐四档尺寸。第六十二批补齐状态边框与默认失败图标。第六十三批补齐媒体上传/处理中遮罩、进度环和失败重试/禁止图标。第六十四批补齐 MediaOverlay 自定义叠加层及独立交互。第六十五批补齐上传/处理中标题扫光。第六十六批补齐 Title/Description 独立状态覆盖、恢复继承及自定义描述。第六十批补齐组 ScrollTo/ScrollState 公开滚动控制，第六十一批补齐 EdgeFade 边缘渐隐。第六十七批补齐默认方形竖排预览、MediaAspectRatio、已加载图片 cover 裁剪与右上角操作。第六十九批补齐 MediaSource 后台加载、取消/替换保护和独立图片重试；第六十八批补齐可选分区开关及纯图片无元信息布局。仍缺按悬停显示移除按钮和标题扫光样式的独立配置入口。 |
| [Avatar](https://gpui-kit.com/component/avatar/) | 主体已有 | [图片/首字母回退、URL 加载与重试、尺寸、状态标记](../../ui/kit/avatar.go)、[叠放头像组/上限/+N/省略号](../../ui/kit/avatar_group.go) | 第七批已关闭原登记缺口；加载不跨实例缓存。外观仍为圆形和主题色回退，GPUI 的自定义占位图标、边框/圆角等样式接口及配色算法不同。 |
| [Badge](https://gpui-kit.com/component/badge/) | 主体已有 | [数字/圆点/图标、上限、尺寸、自定义颜色/名称与角标容器](../../ui/kit/badge.go) | 第十七批已关闭登记缺口；图标模式不依赖计数，数字/圆点仍在 count≤0 时隐藏。Size 为 dp，圆点按比例缩放；Tone 清除固定颜色覆盖。 |
| [Bubble](https://gpui-kit.com/component/bubble/) | 部分 | [可复用内容气泡、Mine 对齐](../../ui/kit/message.go) | 缺 ghost 等外观变体、气泡组和独立反应槽；当前只接 content + Mine，操作/反应在 Message 层。 |
| [Button](https://gpui-kit.com/component/button/) | 主体已有 | [九种变体、描边/紧凑、自定义内容/配色、禁用、尺寸、图标、加载](../../ui/kit/button.go) | 第十一批已关闭登记的变体、样式和内容缺口。Outline/Compact 为叠加配置；自定义内容限展示元素。Tooltip 可外部组合，本项不表示与上游所有组合接口完全相同。 |
| [Calendar](https://gpui-kit.com/component/calendar/) | 主体已有 | [年月切换、多月、范围、禁用日期、键盘](../../ui/kit/calendar.go) | 主体覆盖；禁用日期用函数、年份限制可用 Bounds 表达。缺组件尺寸档，API 组织不同。 |
| [Carousel](https://gpui-kit.com/component/carousel/) | 部分 | [轮播、指示器、键盘、禁用与定时暂停](../../ui/kit/carousel.go) | 缺竖向轨道、同屏多项、可组合前后控件；Keel 每次只显示一张，另有自动播放。 |
| [Chart](https://gpui-kit.com/component/chart/) | 部分 | [折线/柱状/面积/饼图/环图/蜡烛图、图例与数据表](../../ui/kit/chart.go) | 缺 RadarChart、SankeyChart；已有折线/柱/面积/饼环/蜡烛图。轴域、刻度数量、参考线、线型和 tooltip 内容的公共配置较少。 |
| [Checkbox](https://gpui-kit.com/component/checkbox/) | 主体已有 | [布尔选择、半选、回调、禁用](../../ui/kit/checkbox.go) | 第三十八批已补齐 Size/TextSize、TabIndex/TabStop，保留半选能力。尺寸使用连续 dp/sp，Tab 排序限单 el root；焦点轮廓仍沿整行。 |
| [Clipboard](https://gpui-kit.com/component/clipboard/) | 主体已有 | [通用复制按钮、提示与连续复制反馈](../../ui/kit/copy_button.go) | 第五批已补齐 OnCopied、Content 与反馈状态查询；回调表示已提交写入请求，非操作系统成功确认。此表登记缺口已关闭。 |
| [Collapsible](https://gpui-kit.com/component/collapsible/) | 主体已有 | [独立 Trigger/Content、动画、焦点恢复](../../ui/kit/collapsible.go) | 主体覆盖：拆分 Trigger/Content、状态控制与动画；本轮未发现新的主要功能缺口。 |
| [ColorPicker](https://gpui-kit.com/component/color-picker/) | 主体已有 | [HSV、透明度、HEX、预设、键盘、禁用](../../ui/kit/color_picker.go) | 颜色编辑主体已有；GPUI 自带触发器/弹层，Keel 是内联选择器，弹层需组合 Popover；缺触发图标、标签与尺寸配置。 |
| [Combobox](https://gpui-kit.com/component/combobox/) | 部分 | [过滤、多选标签、异步结果、重试、虚拟化](../../ui/kit/combobox.go) | 缺分组、单项禁用、自定义行/触发器、footer；目前候选数据是 string 列表。多选与异步搜索已完成。 |
| [Command](https://gpui-kit.com/component/command/) | 部分 | [模糊过滤、分组、快捷键、异步结果、虚拟化](../../ui/kit/command.go) | 缺内联模式、关闭搜索的模式、自定义行/header/footer；当前固定为带搜索的模态命令面板。 |
| [DataTable](https://gpui-kit.com/component/data-table/) | 部分 | [横向滚动、冻结列、列管理、多选/单元格选择、复制、筛选、分页加载](../../ui/kit/table.go) | 主要数据表能力已有；缺独立整列选择模式、列级 selectable/resizable/movable 限制，以及 stripe/密度等公开配置。 |
| [DatePicker](https://gpui-kit.com/component/date-picker/) | 部分 | [日历弹层、范围、多月、取消草稿、键盘](../../ui/kit/date_picker.go) | 缺日期+时间联动、快捷日期/范围预设、组件级 date_format 与清空按钮；已有独立 TimeField 不等于 DatePicker 已集成。 |
| [DescriptionList](https://gpui-kit.com/component/description-list/) | 主体已有 | [多列/跨列、横纵标签、富值插槽、分隔线、边框、字号与标签宽度](../../ui/kit/description_list.go) | 第八批已关闭登记缺口。Columns 由调用方设置，不按窗口宽度自动切换；默认仍为无边框单列，保留原用法。 |
| [Dialog](https://gpui-kit.com/component/dialog/) | 主体已有 | [可组合内容、嵌套浮层、长内容、焦点约束与恢复](../../ui/kit/dialog.go) | 第四十二批已补齐遮罩显示、外部点击关闭、Esc、关闭按钮的独立开关。关闭按钮默认隐藏以保持兼容。Body/Footer 可组合，但非 GPUI 的完整 compound parts API。 |
| [Dock](https://gpui-kit.com/component/dock/) | 部分 | [边缘与中心区标签组、嵌套分割、拖放、布局保存、最大化、跨窗口分离](../../ui/kit/dock.go) | 中心/边缘嵌套分割、拖放、最大化已完成；缺 GPUI 的面板工厂注册/面板自有状态恢复和独立 DockSkin。分离由 OnDetach 交给应用开窗，恢复布局不会重开分离窗口。 |
| [DropdownButton](https://gpui-kit.com/component/dropdown_button/) | 主体已有 | [按钮菜单、分体按钮、键盘与焦点恢复](../../ui/kit/dropdown_button.go) | 第四十五批补齐 Button 配置透传、Loading 和共享 Size；默认继承内层变体/高度，分体主按钮加载不阻挡箭头。第四十六批补齐 Placement/Offset，普通模式锚定整按钮，分体模式锚定箭头；内部 ID 由组件管理。菜单仍按 Keel 的边缘翻转策略定位。 |
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
| [Kbd](https://gpui-kit.com/component/kbd/) | 主体已有 | [平台键帽、Plain、动作键位、独立字号与自定义样式](../../ui/kit/kbd.go) | 第二十五批已补齐登记的尺寸缺口及样式回调；Size 单位 sp，默认仍继承。KbdFor 读取动作首个绑定，不提供上游按焦点/上下文查询绑定的独立入口。 |
| [Label](https://gpui-kit.com/component/label/) | 部分 | [Text、字号/颜色、For 标签关联聚焦](../../ui/el/element.go) | el.Text 能排版和关联字段；缺 GPUI Label 的匹配区间高亮、masked 和 secondary 文案的专用接口。 |
| [List](https://gpui-kit.com/component/list/) | 部分 | [列表项、稳定 ID、单项禁用、多选/范围、键盘、拖动](../../ui/kit/list.go) | 缺分组头、自定义行/图标/行内操作、内建搜索与加载更多入口；当前是可选择、可拖动的文字列表。 |
| [Marker](https://gpui-kit.com/component/marker/) | 用途不同 | [几何标记、大小与颜色](../../ui/kit/marker.go) | 用途不同：GPUI 是带图标/文字、分隔线/边框、加载状态的消息标记行；Keel Marker 只绘制点/方块等几何标记，不能计作对齐。 |
| [Menu](https://gpui-kit.com/component/menu/) | 部分 | [菜单、子菜单、分隔线、长内容、方向键、焦点恢复](../../ui/kit/menu.go) | 第四十七批已补齐图标、勾选项及勾号左右位置，含状态更新/查询与 Agent checked。第四十八批补齐不可交互的组标题 Label。第四十九批补齐 ContentItem/SetItemContent 展示内容行及变高定位。第五十批补齐 Link、外链图标开关、系统打开及应用回调。快捷键仍取动作首个绑定，缺按触发器焦点上下文解析绑定；默认系统打开未做各平台真机验收。 |
| [MessageScroller](https://gpui-kit.com/component/message-scroller/) | 部分 | [可变高度虚拟化、跟随尾部、流式增高、历史加载锚点](../../ui/kit/message_scroller.go) | 虚拟化、尾部跟随、历史锚点与“最新”按钮已有；缺公开按消息跳转/初始未读定位、跟随状态查询和自定义跳转按钮。 |
| [Message](https://gpui-kit.com/component/message/) | 部分 | [消息内容、状态、操作栏、反应、失败重试](../../ui/kit/message.go) | 缺独立 avatar/header/footer 插槽、MessageGroup 和 ghost/content_inset 配置；当前是作者文字+内容+操作/反应。 |
| [Notification](https://gpui-kit.com/component/notification/) | 部分 | [通知队列、超时、关闭、暂停与原位更新](../../ui/kit/notification.go) | 缺系统通知投递、位置选择、操作按钮与任意富内容；当前只有应用内右上角标题/正文通知队列。 |
| [NumberInput](https://gpui-kit.com/component/number-input/) | 部分 | [数值解析、范围/步长/精度、草稿提交与取消](../../ui/kit/number_input.go) | 缺金额/千分位 mask、动态 step_by、前后内容槽；固定步长、精度、范围与输入草稿已有。 |
| [OtpInput](https://gpui-kit.com/component/otp-input/) | 主体已有 | [分格输入、粘贴、完成回调、密码遮罩、分组、尺寸与窄布局](../../ui/kit/otp_input.go) | 第二批已补齐 Masked/Groups/Size；默认两组，不能整除时前组多一位。此表登记的三个缺口已关闭。 |
| [Pagination](https://gpui-kit.com/component/pagination/) | 主体已有 | [页码、前后翻页、总数、窄布局换行](../../ui/kit/pagination.go) | 第三十三批已补齐紧凑模式、数字按钮上限、尺寸和整体禁用。Size 为连续 dp；VisiblePages 正值限制在 3–101，0 恢复 Keel 原窗口策略，默认策略与上游五按钮不同。 |
| [Plot](https://gpui-kit.com/component/plot/) | 用途不同 | [成品散点/折线图、缩放、平移、拾取](../../ui/kit/plot.go) | 用途不同：GPUI 提供 ScaleLinear/Band/Point/Ordinal、Bar/Line/Area/Pie/Stack/Axis 等公共绘图基础件；Keel Plot 是可缩放平移的成品散点/折线图。 |
| [Popover](https://gpui-kit.com/component/popover/) | 主体已有 | [锚点定位、避让、长内容、外部点击/Esc、焦点恢复](../../ui/kit/popover.go) | 第二十八、二十九批已补齐实例 Offset、默认外观开关及面板样式。第三十至三十二批补齐左/右/中键选择及箭头。箭头用纯色背景，边框/阴影/渐变不延伸到箭头；Keel 在空间不足时翻转，上游保持锚点方向并限制位置，定位策略不同。 |
| [Progress](https://gpui-kit.com/component/progress/) | 主体已有 | [条形确定/不确定进度](../../ui/kit/progress.go)、[圆形进度与中心内容](../../ui/kit/progress_circle.go) | 第一、二十六、二十七批已补齐圆形进度、条形样式及数值过渡。Keel 数值范围 0–1，条形默认带标签/百分比；图形过渡 200ms，语义立即报告目标值，减少动画立即归位。 |
| [Questionnaire](https://gpui-kit.com/component/questionnaire/) | 部分 | [题型、答案模型、校验、分页、禁用与提交快照](../../ui/kit/questionnaire.go) | 缺单题条件禁用、跳过状态、自定义/外部校验、同题选项+自由输入、完整进度状态和快捷键配置；现有五种题型、必填校验与分页保留。 |
| [Radio](https://gpui-kit.com/component/radio/) | 主体已有 | [单选组、横纵布局、独立 Item、单项禁用、键盘](../../ui/kit/radio_group.go) | 第三十九批已补齐 Size/TextSize 与按选项配置的富标签 Content，独立 Item 共享配置。第四十批增加 ItemSize，逐项覆盖尺寸/字号，0 继承组配置；尺寸采用连续 dp/sp。第四十一批补齐组级 TabStop/TabIndex 和逐项 ItemTab/ClearItemTab；默认单停靠点，显式逐项配置可覆盖。排序限单 el root，跨原生 Gio/独立 Embed 不支持。 |
| [Rating](https://gpui-kit.com/component/rating/) | 主体已有 | [评分、已填星减分、尺寸/颜色、半星/小数展示、只读与键盘](../../ui/kit/rating.go) | 第二十四批已关闭登记缺口；按上游源码明确为点已填第 i 星设 i−1 分，并非总分减一。Size 为 dp，默认 22；小数只用于展示，编辑仍选整星。 |
| [Resizable](https://gpui-kit.com/component/resizable/) | 部分 | [横纵分割、最小尺寸、拖动、键盘、取消与禁用](../../ui/kit/resizable.go) | 缺独立多面板 group、最大尺寸、条件显隐/把手外观配置；Keel 为双面板，可嵌套组合更多面板。 |
| [Root View](https://gpui-kit.com/component/root/) | 主体已有 | [根布局、统一浮层宿主、窗口快捷键](../../ui/el/root.go) | 架构差异：Keel 已有 root/overlay/focus/shortcut；Dialog/Sheet/Notifier 需应用挂载，GPUI 0.7 根视图自动挂载这些层。 |
| [Scrollable](https://gpui-kit.com/component/scrollable/) | 主体已有 | [ScrollX/ScrollY、滚动条拖动/轨道点击、定位与尾部跟随](../../ui/el/viewport.go) | 双轴滚动、滚动条与定位已有，通过 el 组合；缺组件级 Always/Hover/Scrolling 显示策略。没有独立类型本身不计功能缺失。 |
| [Select](https://gpui-kit.com/component/select/) | 主体已有 | [过滤、分组、多选、禁用项、万条虚拟化](../../ui/kit/select.go) | 单选主体覆盖，另有多选；缺自定义行/空内容/标题前缀、清空按钮与菜单宽高配置。分组、禁用项已实现。 |
| [Settings](https://gpui-kit.com/component/settings/) | 部分 | [设置分组、导航、搜索、窄布局](../../ui/kit/settings.go) | 缺页面下的多 Group 模型、resettable 重置、组 footer、独立搜索 keywords 和 Markdown 描述；已有分区导航、搜索与窄布局。 |
| [Sheet](https://gpui-kit.com/component/sheet/) | 主体已有 | [侧边抽屉、遮罩、长内容、焦点与禁用继承](../../ui/kit/sheet.go) | 第五十一批补齐独立 Footer 及 Keyboard/Overlay/OverlayClosable/CloseButton。第五十二批补齐 MarginTop 及动画裁剪。第五十三批补齐 PanelStyle 面板样式。第五十四批补齐四方向拖动调整尺寸及回调；当前登记缺口已关闭。把手默认开启，用户最小尺寸 80dp，最大为可用窗口尺寸，支持键盘和取消恢复；不表示各平台真机验收完成。 |
| [Shimmer](https://gpui-kit.com/component/shimmer/) | 主体已有 | [可读文字扫光、周期、宽度、反向、单次、重播、减少动画](../../ui/kit/shimmer_text.go) | 第六十五批新增独立 ShimmerText；文字保持字体/字重/行高与截断，单次结束恢复普通文字。默认 2 秒、半宽 0.3、主题 PrimaryText 高光，支持自定义配色。彩色位图字形保留原色，不参与高光着色。 |
| [Sidebar](https://gpui-kit.com/component/sidebar/) | 主体已有 | [嵌套分组、收起、选中、固定头尾、键盘滚动](../../ui/kit/sidebar.go) | 主体覆盖；缺右侧布局开关、自定义 item suffix/上下文菜单接口。已有 Badge 和固定 Header/Footer。 |
| [Skeleton](https://gpui-kit.com/component/skeleton/) | 主体已有 | [占位形状、尺寸、次级色阶、自定义圆角与加载动画](../../ui/kit/skeleton.go) | 第二十三批已关闭登记缺口；保留 Keel 的 1.5 秒明暗脉冲/可选扫光及减少动画，默认颜色来自 Subtle/SubtleHover，与上游独立 skeleton token、2 秒透明度动画不同。 |
| [Slider](https://gpui-kit.com/component/slider/) | 主体已有 | [单值/双端范围、横纵向、线性/对数、步长、拖动/键盘与结束回调](../../ui/kit/slider.go) | 第十批已关闭对数刻度和 Release 缺口；无效对数范围回退线性，取消不回滚已有值。轨道/滑块颜色与大小仍使用统一样式，未提供逐项外观配置。 |
| [Spinner](https://gpui-kit.com/component/spinner/) | 主体已有 | [不确定动画、减少动画、可访问名称、自定义图标/颜色/周期](../../ui/kit/spinner.go) | 第十九、二十批已补齐图标、颜色和速度配置；Period 为每周时长，默认一秒匀速。上游描述的默认 0.8 秒及缓动曲线不同。 |
| [StatusBar](https://gpui-kit.com/component/status-bar/) | 主体已有 | [固定状态栏、左右内容组、按优先级收起的溢出菜单](../../ui/kit/status_bar.go) | 左右内容与自定义 View 已覆盖；Keel 另有优先级溢出菜单，本轮未发现新的主要功能缺口。 |
| [Stepper](https://gpui-kit.com/component/stepper/) | 主体已有 | [横纵步骤、图标/富内容、尺寸、导航、键盘、滚动与单步禁用](../../ui/kit/stepper.go) | 第六批已补齐 Vertical、Size、StepperItem 与 SetItemDisabled。此表登记缺口已关闭；导航仍限已完成步骤，GPUI 文档的文本居中布局未提供独立开关。 |
| [Switch](https://gpui-kit.com/component/switch/) | 主体已有 | [布尔开关、标签、禁用与键盘](../../ui/kit/switch.go) | 第三十四批已补齐大小、选中颜色和标签侧配置。第三十五批补齐滑块过渡；第三十六批补齐 FocusRing；第三十七批补齐单 root 的 TabStop/TabIndex。焦点轮廓沿整行而非仅轨道，跨 root 排序不支持；Tooltip 可组合。当前无 Loading 接口，上游此页也未列为能力。 |
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
