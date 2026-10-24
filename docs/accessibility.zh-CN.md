# 无障碍

[English](accessibility.md) | 简体中文

在 macOS 上，VoiceOver 等辅助应用可以读取 Keel 窗口并按下其中的控件。应用不需要为此写代码：读屏软件看到的就是 Agent 在[自动化测试](automation.zh-CN.md#元素)里读到的元素列表，角色、名称、值和状态都相同。

Linux（AT-SPI）、Windows（UI Automation）、iOS 和 Android 尚未接入。

## 读屏软件拿到什么

每个元素成为窗口内容视图下的一个 `NSAccessibilityElement`：

| 元素角色 | AppKit 角色 |
| --- | --- |
| `button`、`columnheader`、`gridcell` | `AXButton` |
| `link` | `AXLink` |
| `textbox` | `AXTextField` |
| `checkbox` | `AXCheckBox`；值为 0、1，半选时为 2 |
| `switch`、`toggle` | `AXCheckBox`，子角色 `AXSwitch` / `AXToggle` |
| `radio` | `AXRadioButton` |
| `tab`、`tablist` | `AXRadioButton`（`AXTabButton`）、`AXTabGroup` |
| `select`、`combobox` | `AXPopUpButton` |
| `slider`、`progressbar` | `AXSlider`、`AXProgressIndicator` |
| `image`、`avatar` | `AXImage` |
| `menu`、`menuitem` | `AXMenu`、`AXMenuItem` |
| `list`、`listbox` | `AXList` |
| `table`、`row`、`tree`、`treeitem` | `AXTable`、`AXRow`、`AXOutline`、`AXRow` |
| `disclosure` | `AXDisclosureTriangle`；展开时值为 1 |
| `toolbar`、`separator`、`tooltip` | `AXToolbar`、`AXSplitter`、`AXHelpTag` |
| 滚动容器（`ScrollY`、`ScrollX`） | `AXScrollArea` |
| 其他容器（`dialog`、`group`、`form` 等） | `AXGroup` |
| `text` 及其他角色 | `AXStaticText`，文字作为值 |

容器把内部元素作为子元素，所以 VoiceOver 进入对话框或分组的方式和原生应用一样。禁用和选中状态会保留。按下（VO-空格）会点击元素可见部分的中心，执行的处理函数与鼠标点击相同。位置在被询问时才计算，窗口移动后仍然正确。

滚动区域包含其中的全部内容，包括滚出视野的部分，位置就是它们的实际位置。VoiceOver 可以移到这样的元素上并把它滚进视野（`AXScrollToVisible`）；滚动区域也能按页滚动（`AXScrollUpByPage` 及其他方向）。两者都像鼠标滚轮那样在区域上滚动。元素滚进视野后身份不变，VoiceOver 不会丢失位置。Agent 仍然只看到显示出来的内容：视野外的元素和滚动区域只提供给无障碍。

要让自定义组件可读，按自动化测试的要求给它角色和名称即可，见[扩展指南](extending.zh-CN.md)和 `core.Role`。

## 开销

在辅助应用向窗口询问元素之前，什么都不运行。此后该窗口每画一帧，就在主线程排队一次刷新：根据这一帧的语义树重建元素列表，只更新变化了的对象，并发送 `AXValueChanged` 和 `AXLayoutChanged` 通知。空闲窗口不画帧，也就没有无障碍开销。

描述滚出视野的内容会让辅助应用活跃期间的每一帧变慢。在 Apple M4 上，一个 500 行、可见约 20 行的列表，el 每帧由约 710 µs 变为约 856 µs（`ui/el` 的 `BenchmarkScrollFrame*`，+20%），没有额外分配；没有辅助应用时每帧不变。由 Gio 组件（`el.Widget`）绘制的内容和虚拟列表只在布局到的地方存在，因此只描述其可见部分。

## 原理

Gio 把每帧的语义树放在主线程上，AppKit 也在主线程上询问无障碍信息。Keel 的 Gio 副本在这个线程上暴露语义树（`Window.AppendSemantics`）并接受合成输入（`Window.ProcessInput`），见 `third_party/README.zh-CN.md` 的补丁记录。`ui/window/accessibility_darwin.go` 给 Gio 的视图类加上 `accessibilityChildren`、`accessibilityRole` 和 `accessibilityHitTest:`，并注册一个元素类，每个进程只做一次。这些回调不获取帧锁。

## 局限

- 不报告键盘焦点（`accessibilityFocusedUIElement`）。VoiceOver 跟随自己的光标，而不是 Keel 的焦点。
- 文本框对辅助应用只读：能读取值，不能设置，也不暴露文本范围和选区。
- 滑块不支持增减操作；没有实时区域播报。
- 内容是会变化的文字的元素（计数器、“保存”变成“已保存”）会得到新的身份，VoiceOver 可能回到窗口开头。
- 真实窗口测试（`KEEL_DESKTOP=1 go test -run RealWindowsAccessibility ./ui/window`）在同一进程内通过 AppKit 的接口询问视图。读取、按下和滚动还在另一个进程里通过 VoiceOver 所用的 `AXUIElement` 接口，在 `examples/hello` 和 `examples/components` 上验证过。尚未在开启 VoiceOver 的情况下验证。
