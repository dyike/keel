# 无障碍

[English](accessibility.md) | 简体中文

在 macOS（VoiceOver）、Windows（讲述人等 UI Automation 客户端）和 Linux（Orca 等 AT-SPI 客户端）上，辅助技术可以读取 Keel 窗口并按下其中的控件。应用不需要为此写代码：读屏软件看到的就是 Agent 在[自动化测试](automation.zh-CN.md#元素)里读到的元素列表，角色、名称、值和状态都相同。

iOS 和 Android 尚未接入。

## macOS 上读屏软件拿到什么

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

容器把内部元素作为子元素，所以 VoiceOver 进入对话框或分组的方式和原生应用一样。禁用和选中状态会保留。按下（VO-空格）会点击元素可见部分的中心，执行的处理函数与鼠标点击相同。位置在被询问时才计算，窗口移动后仍然正确。持有键盘焦点的元素就是视图的 `accessibilityFocusedUIElement`，焦点移动时发出 `AXFocusedUIElementChanged`，所以 VoiceOver 会跟随 Tab。设置元素的 `accessibilityFocused` 会把键盘焦点移到它上面。

滚动区域包含其中的全部内容，包括滚出视野的部分，位置就是它们的实际位置。VoiceOver 可以移到这样的元素上并把它滚进视野（`AXScrollToVisible`）；滚动区域也能按页滚动（`AXScrollUpByPage` 及其他方向）。两者都像鼠标滚轮那样在区域上滚动。元素滚进视野后身份不变，VoiceOver 不会丢失位置。Agent 仍然只看到显示出来的内容：视野外的元素和滚动区域只提供给无障碍。

要让自定义组件可读，按自动化测试的要求给它角色和名称即可，见[扩展指南](extending.zh-CN.md)和 `core.Role`。

## Linux（AT-SPI）

第一个窗口在后台连接会话声明的无障碍总线（`org.a11y.Bus.GetAddress`），并向注册表登记应用（`Socket.Embed`）。没有无障碍总线时什么都不做。应用、每个窗口（角色 frame）和每个元素都是实现 `Accessible` 和 `Component` 的 D-Bus 对象；能按下的元素还实现 `Action`，只有一个动作 "click"。角色沿用同一套元素角色，例如 push button、check box、entry、link、page tab、combo box、list item、table row、scroll pane 和 label（`ui/internal/atspi/roles.go`，数值对照 `atspi-constants.h` 检查）。状态包括 enabled、showing、checked、selected、expanded 和 editable。

`Component.ScrollTo` 把元素滚进视野，`DoAction` 点击它。持有键盘焦点的元素带 focused 状态；`Component.GrabFocus` 把焦点移到元素上。变化通过 `ChildrenChanged`、`StateChanged`（checked、selected、enabled、showing、focused、active、defunct）和 `PropertyChange`（名称）通知，窗口焦点通过 `Window.Activate` 和 `Deactivate` 通知。

调用在 D-Bus 协程上到达，在锁内读取上一帧的元素。动作变成排给窗口的输入（`Window.QueueInput`），由 Gio 的驱动在下一个事件之前投递：在 X11 和 Wayland 上，驱动运行在窗口协程的 `Window.Event` 内部，只有那里可以操作它的输入状态。

## Windows（UI Automation）

窗口用一个根提供者回答 `WM_GETOBJECT`，Windows 会把它和窗口自己的提供者（标题、位置）合并。每个元素都是实现 `IRawElementProviderSimple` 和 `IRawElementProviderFragment` 的提供者，并按角色提供 Invoke（按钮、链接、菜单项、行）、Toggle（复选框、开关）、ScrollItem（滚动区域内的所有元素）和只读 Value（文本框、选择框、滑块、进度）模式。控件类型沿用元素角色：Button、CheckBox、Edit、Hyperlink、TabItem、ComboBox、ListItem、DataItem，滚动区域和对话框为 Pane，以及 Text、Group。`IsOffscreen` 标出滚出视野的元素。持有键盘焦点的元素 `HasKeyboardFocus` 为真，也是片段根 `GetFocus` 的结果；`SetFocus` 把焦点移到元素上。结构变化时发出 `StructureChanged`，切换时发出 `ToggleState` 属性变化，焦点移动时发出 `AutomationFocusChanged`。

提供者是不用 cgo 构建的 COM 对象：每个对象是一块原生堆内存，每个接口一个虚表指针；虚表项是约 40 个只创建一次的回调，Go 侧按内存块地址找到对象。第一次 `WM_GETOBJECT` 之前什么都不运行；之后 Gio 在窗口线程上把每帧的语义树交给观察者（`Window.ObserveSemantics`），按下和滚动变成排队的输入（`Window.QueueInput`），因为 UI Automation 在它自己的线程上调用。

## 开销

在辅助应用询问元素之前，什么都不运行。此后每画一帧就刷新一次：macOS 在主线程排队刷新，Linux 在帧结束后的窗口协程上刷新。刷新时根据这一帧的语义树重建元素列表，只更新变化了的对象，并发出变化通知（macOS 上是 `AXValueChanged` 和 `AXLayoutChanged`，Linux 上是上文的 AT-SPI 事件）。空闲窗口不画帧，也就没有无障碍开销。

描述滚出视野的内容会让辅助应用活跃期间的每一帧变慢。在 Apple M4 上，一个 500 行、可见约 20 行的列表，el 每帧由约 710 µs 变为约 856 µs（`ui/el` 的 `BenchmarkScrollFrame*`，+20%），没有额外分配；没有辅助应用时每帧不变。由 Gio 组件（`el.Widget`）绘制的内容和虚拟列表只在布局到的地方存在，因此只描述其可见部分。

## 原理

在 macOS 上，Gio 把每帧的语义树放在主线程上，AppKit 也在主线程上询问无障碍信息。Keel 的 Gio 副本在这个线程上暴露语义树（`Window.AppendSemantics`）并接受合成输入（`Window.ProcessInput`），见 `third_party/README.zh-CN.md` 的补丁记录。`ui/window/accessibility_darwin.go` 给 Gio 的视图类加上 `accessibilityChildren`、`accessibilityRole` 和 `accessibilityHitTest:`，并注册一个元素类，每个进程只做一次。这些回调不获取帧锁。

## 局限

- 设置焦点（`setAccessibilityFocused:`、AT-SPI 的 `GrabFocus`、UIA 的 `SetFocus`）只对自身接受键盘焦点或包含输入框的元素有效，对其他元素不起作用。
- 文本框对辅助应用只读：能读取值，不能设置，也不暴露文本范围和选区。
- 滑块不支持增减操作；没有实时区域播报。
- 接受输入的元素（按钮、复选框）在标签和状态变化时保持身份。文字会变化的纯文本（计数器）会变成新元素。
- Windows：命中测试（`ElementProviderFromPoint`）按鼠标位置回答，因为它的坐标是 double，Windows 回调收不到。值不能设置，也没有 Text 和 ExpandCollapse 模式。
- Linux：没有 Text 和 Value 接口，Orca 能读出输入框的名称但读不出内容，滑块和进度条的值只作为描述提供。屏幕坐标就是窗口坐标，因为 Wayland 不报告窗口位置。CI 在 Xvfb 下检查 X11；Wayland 运行同样的代码，但没有用客户端检查。
- 真实窗口测试（`KEEL_DESKTOP=1 go test -run RealWindowsAccessibility ./ui/window`）在同一进程内通过 AppKit 的接口询问视图。读取、按下和滚动还在另一个进程里通过 VoiceOver 所用的 `AXUIElement` 接口，在 `examples/hello` 和 `examples/components` 上验证过。在 Linux 上，CI 在带 at-spi2-core 的 Xvfb 下运行 `ui/window/testdata/atspi`：它作为无障碍总线上的客户端，在注册表的桌面上找到应用，读取角色和名称，按下按钮，把一行滚进视野后按下；`pyatspi_check.py` 通过 Orca 使用的 libatspi 读取 `examples/hello` 并勾选其中的复选框。在 Windows 上，CI 不开窗口直接通过虚表调用提供者，对照 Windows SDK 头文件检查每个 IID、方法顺序和 ID，并在另一个进程里用 .NET 的 UI Automation 客户端读取 `examples/hello` 和 `examples/components`（`ui/window/testdata/uia/check.ps1`）：切换复选框、调用按钮，把视野外的链接滚进视野后调用。VoiceOver、讲述人和 Orca 都还没有实际运行过。
