# Combobox

可输入、可筛选的下拉框。

```go
customer := kit.Combobox("客户", customers...).Placeholder("输入筛选")
tags := kit.Combobox("标签", "紧急", "VIP").AllowCustom()
```

- 输入时打开列表并筛选（不区分大小写，包含即匹配），点击选项选中。
- 回车的规则：
  - 输入的文字正好是某个选项，就选它；
  - 设置了 `AllowCustom` 时，保留输入的文字；
  - 否则选第一个匹配项。
- 没有设置 `AllowCustom` 时，离开输入框后文字如果不是选项，会恢复为上一次的选择。
- ↓ 打开列表并移动高亮，↑ 往回移动，回车选中高亮项。
- `Value()` / `SetValue`、`SetOptions`、`SetDisabled`、`SetError`。需要 `el.Root`。

Agent：容器角色 `combobox`，`value` 为当前选择；里面有 `textbox` 和展开按钮；列表是 `listbox` 和 `option`。

验证：`go run ./examples/components -section combobox`，加 `-theme dark` 检查深色。

`Multiple()` 把选项显示为可移除的标签，输入区留给下一次搜索；`AllowCustom()` 可与它组合。`Values()` 返回选择顺序的副本，`SetValues` 去重后替换选择，不触发回调；`OnValuesChange` 接收用户增删后的副本。多选下应使用 `Values()`，`Value()` 保留最近的主值。`SetValue` 替换成单项选择，程序赋值会关闭结果列表并使旧查询失效。构造与 `SetOptions` 复制选项切片；更新搜索结果不移除已选标签。

`OnSearch(func(query string, token uint64))` 接管结果来源。每次输入、打开或重试都有新 token；异步完成后用 `core.Update` 调用 `SetResults(token, options...)` 或 `SetSearchError(token, message)`。过期、已关闭的结果返回 false，远程结果不做二次本地筛选。加载或失败期间不能执行旧结果，失败后提供重试。关闭、程序赋值或禁用会使请求失效；禁用导致的失焦不会提交草稿。

结果使用虚拟列表和过滤缓存。↑ ↓、PageUp/PageDown 会滚动露出高亮项，结果区高度受窗口限制。示例包含万条客户、自由输入多标签，以及可模拟失败的异步多选。


`DisableOption(value, true)` 禁止用户选择指定值，候选仍显示并带禁用语义。点击、上下方向键、PageUp/PageDown、Enter 和失焦提交均遵守禁用；AllowCustom 不能绕过同值的禁用。全部候选禁用时没有高亮项，也不能提交候选。传 false 恢复可选。

配置按字符串值保存，过滤、SetOptions 和异步 SetResults 不会清掉配置；同值重复候选一起禁用，尚未出现在候选中的值也可预先配置。禁用已选值不会自动删除或触发回调，多选标签仍可移除。SetValue/SetValues 保留应用主动赋值能力。动态禁用当前高亮项时改为下一可用项，全部禁用则清除高亮。


多选列表中，再次点击或按 Enter 确认已选候选会取消选择，弹层保持打开；搜索文字清空，异步模式为清空后的查询申请新 token。每次切换触发一次 OnValuesChange；移除最近主值时 Value 回退到剩余选择的最后一项，无剩余则为空。OnChange 只在该主值变化时触发，移除其他已选项不会重复通知主值。禁用候选不能从列表切换，已选标签仍可用移除按钮删除。


`Footer(view)` 在候选滚动区下方显示持久操作区，加载、失败、空结果时也保留；`Footer(nil)` 移除。内容可含按钮或输入框，候选更新保留同一内容实例的焦点和状态；footer 操作不选中候选、不自动关闭，应用可调用 SetValue/SetValues 应用新选择并关闭。禁用触发字段或祖先会关闭整个弹层。

footer 最多占窗口扣除 80dp 后高度的三分之一，超高内容在自己的区域滚动；候选区为 footer 预留高度。首帧按上限预留，测量后下一帧收敛到实际高度。示例的“添加示例标签”按钮演示了主动更新选择。


`Clearable(true)` 在有已选值时显示独立清空按钮，单选／多选均可用。用户清空会关闭弹层、取消旧查询、清除草稿与错误，焦点返回输入框；OnChange 收到空字符串，多选 OnValuesChange 收到空切片，各通知一次。回调里重新 SetValue 的结果保留；回调参数描述本次清空事件。只剩搜索草稿而没有选择时不显示按钮，禁用状态遵循字段与祖先；`Clearable(false)` 隐藏按钮。程序 SetValue/SetValues 继续不触发回调。


`Size(dp)` 设置字段最小高度并同步缩放字号、间距、展开／清空按钮、多选标签和候选行，建议 28／36／48dp；0 恢复默认，负数和非有限值忽略。多选内容可换行，实际高度可能更高，标签保留 16dp 最小高度。footer 维持自身尺寸。打开期间修改尺寸保留输入焦点并重新露出高亮候选。

`CheckIcon(kit.Icon(kit.IconCheck))` 替换选中候选的图标，也接受 VectorIcon；保存图标配置副本，颜色和基础尺寸来自传入图标，随 Size 缩放。未选中候选保留同宽占位。`CheckIcon(nil)` 恢复默认；传入 IconNone 隐藏图案并保留尺寸。


`SetItems(...ComboboxItem)` 接受 `Value`、`Label`、`Disabled`，将稳定值与显示名称分离。空 Value 忽略，空 Label 使用 Value，重复 Value 保留首项，输入切片会复制。本地搜索同时匹配名称和值；回调、Value/Values 及程序赋值使用稳定值，候选和多选标签显示名称。候选语义的 name 是显示名称，value 是稳定值。

`SetItemResults(token, items...)` 用于结构化异步结果，和 SetResults 一样校验请求状态，不做本地二次筛选。替换候选不会删除选择；已选值不在新结果时保留原名称，返回候选后更新名称。SetOptions/SetResults 恢复字符串模式。条目 Disabled 与 DisableOption 取逻辑或，修改条目状态须重新 SetItems/SetItemResults。

提交输入文字时，精确值优先，其次是首个名称匹配；同名不同值建议从候选选择。未改动的单选显示文字在失焦时保持原值，避免名称碰巧等于另一条目的值时误选。正在编辑的草稿不会被名称更新覆盖。


`RenderItem(func(ComboboxItem, bool) el.View)` 自定义候选正文，参数包含稳定值、显示名称、合并后的禁用状态及是否选中。回调为 nil，或返回 nil 时恢复默认名称。行的选择语义、禁用、背景和尾部勾选仍由组件维护。正文内独立按钮不会连带选择整行；普通文本或空白处点击仍选择候选。禁用行同时禁用其子操作。

仅可见虚拟行调用渲染函数。需要保留输入／按钮状态时，按 Value 复用 View 实例，避免每帧创建新的状态对象；候选重排使用值身份，同值重复字符串按出现次序区分。离开虚拟视口后的焦点不保证保留，持久输入值应放在应用状态中。

`RowHeight(dp)` 设置含上下间距的统一候选行高，正值独立于 Size，0 恢复随 Size 缩放的 30dp；非法值忽略。最小高度为 1dp 正文加当前上下间距。富内容不会自动测出不同的逐行高度，应用应设置足够行高；超出行高的内容受虚拟视口裁剪。


`SetGroups(...ComboboxGroup)` 设置有序分组，每组包含稳定 ID、Label 和 Items；空 ID 忽略，空 Label 使用 ID，重复组 ID 和跨组重复候选值均保留首项。输入切片复制，选择不清除。组标题只展示，不参与选择；无候选的组和搜索后无匹配项的组自动隐藏。搜索匹配候选名称和值，不匹配组标题。

`SetGroupResults(token, groups...)` 提交异步分组结果，仍校验请求 token；SetItems/SetOptions 及其异步入口恢复无分组模式。组标题和候选均在虚拟列表内，使用同一 RowHeight；键盘跳过标题及禁用项，滚动定位包含标题占用空间。标题随内容滚动，不固定在顶部。自定义 RenderItem 只处理候选正文。


`Searchable(false)` 用只选择的按钮替换输入框，默认 Searchable(true)。点击或 Enter 打开，方向键导航、Enter／Space 确认高亮候选，Esc 关闭；多选仍可切换候选并移除标签。输入文字不会更改查询，候选不做本地过滤，AllowCustom 在此模式下不生效。

切换模式关闭弹层、丢弃草稿并使旧异步请求失效，保留已选值；重复设置相同模式不改变状态。无搜索模式打开或重试时，OnSearch 收到空查询，返回结果仍受 token 校验；恢复 Searchable(true) 后重新显示输入框。清空按钮、footer、自定义候选和分组继续可用。


`OnConfirm(func([]string))` 在用户结束一次已打开的选择会话时通知一次：选中单值、Esc、外部点击、展开按钮关闭，以及打开期间清空均适用。多选逐项切换只触发变更回调，关闭时才确认；未打开弹层的清空不触发确认。确认在本次 OnValuesChange/OnChange 之后，参数是本次完成时的选择快照，修改切片不影响组件。

确认不表示选择一定发生变化，也不改变已有的 Esc／外部关闭草稿提交规则。程序赋值、禁用、移除和 Searchable 切换不触发确认。回调中可以重新赋值，原操作不会在回调后覆盖新值；多选变更回调关闭或禁用组件时也不会启动新的搜索。`OnConfirm(nil)` 移除监听。

`RenderTrigger(func(ComboboxTriggerContext) el.View)` 替换整个默认字段外观；nil 回调或 nil 内容恢复默认。上下文提供 Selection 副本（值、名称、候选禁用状态）、Open、组件自身 Disabled、Size、Placeholder，以及可在 UI 事件中调用的 Toggle／Clear。自定义触发器自行绘制边框、标签和清空入口；默认多选标签、展开图标及 Clearable 按钮不再自动插入。组件仍提供标签/错误文案、选择语义、焦点入口与键盘行为，祖先禁用由外层统一执行。

普通展示内容点击背景即可开关；自定义按钮可绑定上下文动作，子按钮不连带触发背景。动作仅用于事件，不能在渲染过程中调用。Enter/Space/方向键可从触发器打开；开启搜索时，面板内出现搜索框并自动聚焦，选择后恢复到触发器（多选仍打开时回到搜索框）。关闭搜索时直接导航候选。首次启用或移除 RenderTrigger 会取消草稿并关闭旧弹层；替换非 nil 渲染器保留选择与展开状态。复用有状态 View 以保留子控件状态。
