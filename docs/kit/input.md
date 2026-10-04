# Input / TextArea

带标签的文本框，可加前后缀、清空按钮和错误提示。`kit.TextArea` 创建多行版本。

```go
search := kit.Input("搜索").Placeholder("客户或单号").Clearable().Prefix(searchIcon)
price := kit.Input("单价").Filter("0123456789.").Suffix(yuan)
note := kit.TextArea("备注").Rows(4)
message := kit.TextArea("消息").AutoGrow(2, 8)
```

- `Value()` / `SetValue`；`OnChange` 在每次编辑后调用，`OnSubmit` 在单行框按回车时调用。
- `Password()` 遮盖内容，`MaxLength(n)` 限制字数，`Filter(chars)` 只接受这些字符（输入和粘贴都过滤）。
- `Clearable()` 在有内容时显示清空按钮，清空后焦点留在输入框里。
- `SetError(msg)` 在下方显示错误并把边框变红，用户再次编辑时自动清除；`Form` 用它显示校验结果。
- `SetDisabled`、`SetReadOnly`；只读时可以选择和复制，不能编辑。
- 文本框会撑满父容器给的宽度。`FocusID()` 返回文本框的元素 ID，可传给 `cx.Focus`。

Agent：角色 `textbox`，名字是标签（没有标签时是占位文字，在 Form 里是行标签），`value` 是内容；清空按钮名为"清空 标签"。

验证：`go run ./examples/components -section input`，加 `-theme dark` 检查深色。

用户修改单行或多行输入后都会清除当前错误，便于重新校验；程序赋值不隐式清除服务端错误，禁用时输入也不会清除错误或触发回调。

`AutoGrow(minRows, maxRows)` 按正文排版后的行数自动增高，包含软换行；超过上限后在编辑器内滚动，删除文字会缩回最小高度。行高随字体和显示缩放计算，不限制文本长度。参数必须满足 `minRows > 0` 且 `maxRows >= minRows`，非法参数忽略；两值相等可固定可见行数。`Rows(n)` 恢复原来的最小高度模式并取消 AutoGrow，单行 Input 忽略 AutoGrow。模式切换保留输入焦点与内容；长占位文字不参与自动增高。父级显式高度或空间约束仍优先。

`Mask(pattern)` 为单行输入设置模板掩码：`#` 是必填 ASCII 数字，`9` 是可选数字，`A` 是 Unicode 字母，`*` 是字母或数字；反斜线转义下一个字符，其余字符为固定文本。固定文本随已填槽位显示，不用占位符补满空槽；空输入保持空字符串。末尾单独的反斜线视为非法配置，保留旧掩码。

```go
phone := kit.Input("电话").Mask("(###)-###-####")
phone.SetValue("1234567890")
// Value(): (123)-456-7890；UnmaskedValue(): 1234567890
amount := kit.Input("金额").NumberMask(',', 2)
amount.SetValue("1234567.89") // 1,234,567.89
```

`NumberMask(separator, fraction)` 将整数部分每三位分组，允许开头负号，以句点作为小数点。separator 为 0 时不分组；fraction 为 -1 时小数位不限，0 时遇到小数点即截断，只保留整数部分，正数限制小数位数；超出部分截去，不四舍五入、不转换浮点数。只有负号或以小数点结尾时仍保留草稿。非法 separator 或 fraction 不改变配置。

`Value()`、`OnChange`、`OnSubmit` 使用格式化文本；`UnmaskedValue()` 返回模板中实际填入的字符或不含分组符的数字文本。`MaskComplete()` 检查必填槽是否填满，数字模式要求至少一个数字且不能以小数点结尾；不验证日期有效性、号码归属等业务规则。只有可选槽的空掩码值也可能为完整，应另加必填校验。

掩码启用时，Filter 约束实际输入字符，MaxLength 限制去格式后的 rune 数，不计算自动插入的模板文本或分组符。设置掩码、长度或 Filter 会重新格式化当前值，但不触发回调；`SetValue` 同样格式化。`Mask("")` 移除掩码并保留当前文本。TextArea 忽略 Mask 和 NumberMask。

编辑通过 el.TransformEdit 同步格式化文本与选区，并使用编辑层的撤销重做记录；删除单个自动插入的分隔符时，按 Backspace/Delete 方向删除邻近的可编辑字符，避免分隔符反复插回导致无法删除。自动测试已覆盖模板/数字分组、Unicode、过滤/长度、分隔符删除、选区、撤销重做及菜单剪切；带数字的固定前缀按完整片段识别，避免吞掉原始输入。

Input 和 TextArea 默认提供右键编辑菜单（触屏上长按输入框打开），包含复制、剪切、粘贴和全选；无选区时复制/剪切禁用，只读时剪切/粘贴禁用，密码框的菜单复制/剪切禁用。菜单锚定在输入框下方，关闭后编辑命令将焦点送回输入框；选择操作使用打开菜单之前保留的编辑器选区。

`ContextMenu(menu)` 使用自定义 Menu，传 nil 恢复默认菜单；`ContextMenuEnabled(false)` 同时关闭默认及自定义菜单。菜单实例属于该输入，其 Trigger 不使用，条目、子菜单、位置、宽度等仍由 Menu 配置。组件或祖先禁用/隐藏时由浮层归属规则关闭菜单。自定义菜单可用 `cx.InputAction(field.FocusID(), el.InputCopy / InputCut / InputPaste / InputSelectAll)` 调用编辑命令，并用 `cx.Focus(field.FocusID())` 恢复焦点。创建一次 Menu 并复用，避免每帧替换导致打开状态丢失。

菜单编辑命令与键盘粘贴共用过滤、格式化及编辑回调，图片/文件拦截见下文。自动测试覆盖右键命中、焦点恢复、只读/密码限制、自定义菜单和掩码剪切；触屏选择菜单仍未实现。

`OnPaste(func(core.ClipboardData) bool)` 拦截键盘或菜单粘贴：返回 true 表示应用已接收，不再插入文本；false 把 Data.Text 交给正常过滤/掩码/回调流程。Images 包含 MIME 与编码数据，Files 为路径，组件不会自行打开文件。`PasteReader(core.ClipboardReader)` 由应用提供异步读取；不配置时回调接收 Gio 文本粘贴。`OnPasteError` 在 UI 线程报告错误，富读取失败后回退到 Gio 文本读取；Gio 文本读取超过 16MiB 或失败则拒绝插入。

组件负责把异步完成送回 UI 线程；等待期间输入文本或选区改变、组件禁用/只读时丢弃旧结果。重复发起粘贴以最新请求为准。平台 reader 必须调用完成回调，不能在 UI 线程同步等待平台主线程。

组件库备注示例已适配 `native/clipboard.Read`：macOS 支持文本、PNG/TIFF 与文件 URL，Windows 支持 Unicode 文本、PNG/DIB 和文件路径；Linux 在 Wayland 下通过窗口自己的连接读取（示例调用 `clipboard.UseWaylandDisplay(mainWindow.WaylandDisplay())`），X11 下用独立连接，都支持 UTF-8 文本、编码图片和本机文件 URI；Wayland 路径尚未在真机运行。收到图片/文件后显示数量并消费粘贴；尚未支持的平台回退文本。自动测试覆盖单行/多行消费与回退、旧文本/选区拒绝、重复完成、只读、超限及流关闭。已在 macOS 主线程运行桥接，成功读取当前图片剪贴板；文件 URL、PNG/TIFF 各类型及真实窗口粘贴尚未逐项验收。CodeEditor 同类钩子已验证多光标和撤销。Windows 格式解析与交叉编译、X11 格式和分块传输测试通过，两平台系统剪贴板和真实窗口粘贴待真机验收。

## 尺寸

`Size(InputSizeXSmall/Small/Medium/Large)` 可用于 Input 和 TextArea。Medium 保留原有主题尺寸；其余档调整输入字号、框的最小高度和内边距。TextArea 的固定 Rows 高度随字号变化，AutoGrow 继续按实际文字测量；标签、前后缀和自定义内容保留自己的样式。InputGroup 内仍由组控制外框和内边距。双倍率尺寸、焦点和值保持、AutoGrow 增长/收缩及固定 Rows 字号缩放已验证。

## 原子引用

```go
field := kit.Input("引用")
draft, err := kit.NewInputContent("查看 docs/input.md", kit.InputTokenSpan{
    Range: kit.InputRange{Start: len("查看 "), End: len("查看 docs/input.md")},
    Token: kit.InputToken{ID: "input-doc", Text: "docs/input.md", Label: "输入组件文档"},
})
if err == nil {
    err = field.SetContent(draft)
}
field.OnTokenActivate(func(token kit.InputToken) { /* 应用打开 token.ID */ })
```

`Content()` 返回包含提交文本和引用元数据的独立草稿；范围使用 UTF-8 字节，须落在字素边界且不能重叠。Text 必须与范围内的文字一致，ID 不得为空；同一个 ID 可以多次出现。Label 省略时显示 Text。显示名称和提交文本可以不同，`Value()`、选区复制、Form 取值均使用提交文本。

`SetContent` 恢复草稿并清空撤销；`SetValue` 即使文字相同也移除全部引用并清空撤销。`ReplaceWithToken` 替换当前选区，记录撤销并触发 OnChange；禁用、只读或输入法组合输入期间返回错误。普通输入或粘贴不会把相同文字自动识别成引用；替换引用为相同显示文字也会移除引用 ID。删除和非空选区覆盖引用的任意部分时按整个引用处理，撤销恢复元数据和选区。

点击引用会选择它并调用 `OnTokenActivate`；拖选、Shift 点击和禁用状态不会调用，只读状态允许查看。应用可以把 `ActivateToken()` 绑定到快捷键，激活完整选中的引用。显示背景复用主题颜色；Agent 可以读取引用名称及字段提交文本。

Input、TextArea 和 InputGroup 内的输入可以使用这条编辑路径。带密码、掩码、Filter 或 MaxLength 的字段拒绝 SetContent；随后启用这些模式会退出引用编辑，保留当前提交文本。富粘贴仍使用 PasteReader/OnPaste，恢复草稿后拒绝迟到的粘贴结果。

引用现在按单个有独立宽度的对象参与编辑器排版，软换行时整体移到下一行，包括位于文本末尾的引用。默认显示标签背景块；宽于输入视口时按可用宽度限制内容，不拆分为多个引用。真实 Gio 事件测试覆盖删除、撤销、光标、组合输入事务、异步粘贴和点击/拖选；原生输入法候选位置、组合下划线和跨平台交互尚未验收。

原子引用输入现在维护组合范围，按实际排版位置绘制组合下划线，并向平台上报可见组合边界和光标基线；单行横向滚动、多行换行/裁剪以及 1×/2× 已有自动检查。组合结束、失焦、禁用、转只读和重置草稿都会清除旧范围。平台候选窗口的最终显示仍需原生窗口验收。

`TokenRenderer(func(gtx core.C, token InputToken) core.D)` 自定义引用内容，可绘制图标和文本，返回像素尺寸及距底部的基线。回调先在禁用上下文测量，再按测得大小绘制；须遵守 Constraints，不修改应用状态、不注册交互处理器。引用激活仍通过 OnTokenActivate。传 nil 恢复默认标签块。示例使用 el.Embed 构造图标加文字的被动视图，同时展示 TextArea.AutoGrow 下的整块换行。

编辑器局部字体只负责对象几何，不注册到全局字体；普通字符和已加载字体保留回退路径。内部占位字符避开原文及标签已有字符；输入法上报的是可读标签，复制/Form 使用原文，平台编辑区间和组合范围会转换为对应的布局坐标。1×/2× 测试覆盖自定义尺寸、窄宽度、文本末尾、私用字符冲突、引用后方组合输入和撤销；触屏长按菜单及原生 IME 验收仍未完成。
