# Menu

[English](menu.md) | 简体中文

命令菜单，从触发元素旁边弹出。

```go
export := kit.Menu().Item("PDF", "", exportPDF).Item("CSV", "", exportCSV)
more := kit.Menu().
    Item("复制", "mod+c", copy).
    Separator().
    Sub("导出", export).
    Item("删除", "delete", remove)
more.SetItemDisabled("复制", !hasSelection)
more.Trigger(kit.Button("更多", more.Toggle).Variant(kit.ButtonGhost))
```

- `Item(label, shortcut, action)`：`shortcut` 使用 `core.ParseShortcut` 的写法，只用 `kit.Kbd` 显示，**不注册**快捷键；不需要时传空字符串。
- `Sub(label, menu)` 添加子菜单，`Separator()` 添加分隔线，`SetItemDisabled(label, bool)` 禁用或启用菜单项。
- 菜单是模态的：打开时页面的其余部分不响应点击，焦点限制在菜单内，打开后聚焦第一个可用项。点击外部或按 Esc 关闭，关闭后焦点回到触发元素。
- 键盘：
  - ↑ ↓ 在可用项之间移动，跳过禁用项和分隔线，首尾循环；
  - Home / End 跳到第一项或最后一项；
  - Enter / Space 执行当前项；
  - → 打开子菜单，← 或 Esc 只关闭当前这一层子菜单。
- 执行任意一项后，整个菜单（包括所有子菜单）都会关闭，然后再调用 action。
- 子菜单默认显示在右侧，放不下时翻到左侧。
- `Value()` / `SetValue(bool)` 读取或设置是否打开，`Toggle` 用作触发元素的点击回调，`Width(dp)` 设置最小宽度（默认 220）。

Agent：菜单容器的角色是 `menu`（子菜单的名字是它在父菜单中的标题），菜单项的角色是 `menuitem`；有子菜单的项 `value` 为 `submenu`；禁用的项报告 `disabled`。

验证：`go run ./examples/components -section menu`。

长菜单会限制在窗口内并纵向滚动。方向键、Home / End 会把目标项滚入可见范围；打开时若前面多项禁用，也会显示第一个可用项。宽度过大时受窗口宽度约束。

`SetDisabled(true)` 关闭并禁用整个菜单；禁用正在展开的父项或子菜单会关闭对应分支。关闭顶层时会递归清除展开状态，重新打开不会恢复旧的深层子菜单。`Sub` 忽略 nil、循环引用和重复挂载的子菜单实例；不同分支请分别创建实例。

右键触发可以给 `Trigger` 传入自定义 View，在元素的 `OnContextMenu` 中调用 `Toggle`。键盘入口由触发 View 的 `OnKey` 定义，示例使用 F10；弹层依然锚定触发区域，外部点击和 Esc 的关闭行为相同。

`Placement(side, align)` 设置顶层菜单方向（Top/Bottom/Left/Right）和对齐（Start/Center/End），默认 Bottom/Start；非法组合忽略。`Offset(dp)` 设置间距，默认 4dp，支持 0 和负值重叠，非有限值忽略。打开期间可更新，空间不足时沿用浮层翻转和窗口内限制。子菜单仍按 Right/Start、2dp 展开。

DropdownButton 的这两个方法直接配置传入的 Menu；普通模式锚定整按钮，分体模式锚定箭头。共用同一 Menu 的调用方也会看到配置变化。

`IconItem(label, shortcut, icon, action)` 添加图标命令，`SetItemIcon(label, icon)` 可更新普通项、勾选项或子菜单入口的图标；IconNone 移除。含前置标记的菜单统一预留标记列，文字保持对齐。

`CheckItem(label, shortcut, checked, onChange)` 添加勾选项。点击/Enter/Space 切换内部状态，关闭整个菜单链，再调用 onChange(bool)；禁用项不切换。`SetItemChecked(label, checked)` 不触发回调；`ItemChecked(label)` 返回状态和是否找到。更新方法作用于当前菜单中所有同名项，查询返回首个匹配项，建议使用唯一标签。

`CheckSide(el.Left/Right)` 设置当前菜单的勾号位置；默认左侧替代该项图标，右侧可同时显示图标。未知方向忽略，子菜单单独配置。Agent 中角色为 menuitemcheckbox，并报告 checked 状态。

`Label(text)` 插入不可交互的分组标题，空文字忽略。标题使用次级色和较小的粗体字，固定占 30dp 行高，长文字单行截断；不响应点击，也不会成为方向键、Home/End 或文字搜索的目标。长菜单定位计入标题高度。标题的 Agent 角色为 heading；它只是视觉分节标题，不创建独立子菜单或嵌套 group。

`ContentItem(label, shortcut, content, action)` 添加自定义内容行，可组合多行文字、说明、图标等展示元素；不要嵌套按钮/输入框。label 保留为可访问名称和文字搜索依据。整行点击、Enter/Space 都执行同一个 action，并关闭菜单链。

`SetItemContent(label, content)` 可替换当前菜单同名普通项、勾选项或子菜单入口的显示内容；nil 恢复文字。行身份保持不变，禁用、勾选、快捷键和子菜单箭头沿用原项配置。自定义行最小高度 30dp，按内容增高；长菜单按实际布局高度定位，滚动到可见位置后才转移键盘焦点。

`Link(label, url)` 添加链接项，保留 menuitem 角色，语义 value 为 URL；点击或键盘执行后先关闭整条菜单链，再打开链接。默认通过 `core.OpenURL` 调用平台浏览器/邮件处理程序，仅接受绝对 HTTP、HTTPS 和 mailto URL。`SetItemIcon` 可添加前置图标，`ExternalLinkIcon(false)` 隐藏右侧外链图标。

`OnLink(func(string))` 可接管打开行为和 URL 策略，子菜单优先使用自己的回调，否则沿父菜单查找。默认打开失败时通过 `OnLinkError(func(error))` 通知，同样沿父菜单查找；未配置时忽略错误。默认入口只报告校验和进程启动错误，不表示网页加载成功。Web 使用 window.open，受浏览器弹窗策略限制。


## 按目标区域显示快捷键

`ActionItem` 根据触发器的 `KeyContext` 祖先解析键位，子菜单继承顶层目标。也可以用 `ActionContext("editor")` 指定一个元素 ID，显示该元素所在区域的键位；指定目标不存在、隐藏或禁用时不显示提示。空字符串恢复触发器目标。提示在当前帧构建完成后、测量之前生成，因此首次打开或同帧切换上下文不需要多等一帧。

```go
core.Bind("save", "mod+s")
core.BindIn("editor", "save", "mod+shift+s")
menu := kit.Menu().ActionContext("document").ActionItem("保存", "save", save)
// Render 中：
cx.ActionAt("document", "save", save)
return el.Div().ID("document").KeyContext("editor").Child(input.Render(cx), menu.Render(cx))
```

`core.BindIn` 对一个动作设置区域覆盖，优先顺序为内层、外层、全局。不传键位表示在此区域禁用该动作的绑定；`ClearBindingIn` 删除覆盖并恢复继承。无效键位不会更改原绑定，改绑会请求所有窗口刷新。`Keymap/LoadKeymap` 仍只读写全局绑定，区域绑定由应用配置。

### 上下文条件表达式

`BindIn` 的第一个参数可以是条件表达式，和 GPUI 的 keymap 写法一致：

```go
core.BindIn("Editor && !ReadOnly", "format", "mod+shift+f") // 可编辑的编辑器
core.BindIn("Pane > Editor", "close", "mod+w")               // 面板里的编辑器
core.BindIn("Terminal || Shell", "clear", "mod+k")          // 两者之一
```

- 焦点路径上的每一层是一个 `el.KeyContext`，名字可以包含多个用空格分隔的标识，如 `KeyContext("Editor ReadOnly")`。
- 标识检查当前匹配的这一层；`a > b` 表示这一层满足 b、外面某一层满足 a。优先级从高到低：`!`、`>`、`&&`、`||`，括号分组。
- 解析时从最内层往外找，第一个有绑定成立的层胜出；同一层有多个成立时，后绑定的生效。都不成立时用全局绑定。
- 写错的表达式 `BindIn` 返回错误，什么都不改。单个名称就是只有一个标识的表达式，原来的写法照常生效。

`cx.ActionAt` 仅在指定元素内有焦点时处理解析出的按键。嵌套目标中，较深的目标优先，同层按声明顺序；内层显式空绑定也会阻止外层同动作处理。不要同时用全局 `cx.Action` 注册同一动作。`ActionItem` 传了函数时，点击只调用这个函数；传 nil 时，点击按动作名路由：从触发器（或 `ActionContext` 指定的元素）往外找最内层的 `cx.ActionAt` 处理器，找不到再用 `cx.Action` 的全局处理器，和在那里按快捷键效果一样，而且不需要绑定任何键。这样菜单、快捷键和命令面板可以共用一份命令实现。路由不会切换焦点。底层是 `cx.Perform(targetID, action)`，命令面板等也可以直接调用。普通 `Item` 的显式快捷键和独立 `KbdFor` 保持原有行为。底层 `el.KeyHint` 可用于相同规则的展示内容，不负责注册快捷键。

## 滚动条显示方式

使用 `kit.Menu().Scrollbars(el.ScrollbarAlways)`，可在菜单内容溢出时始终显示滚动条。
其他模式包括 `ScrollbarHover`（鼠标位于菜单内或拖动滚动条时显示）、
`ScrollbarScrolling`（滚动时及停止后短暂显示）和 `ScrollbarSystem`（跟随系统）。
未设置时使用全局 `el.SetScrollbarDefault`。子菜单继承最近父菜单的显式配置，
也可以单独覆盖；无效值会被忽略。
短菜单使用左右对称的行边距；溢出的菜单在滚动条淡出后仍保留点击区域，
避免滚动过程中菜单项左右跳动。组件示例中的长菜单使用 `ScrollbarAlways`。
