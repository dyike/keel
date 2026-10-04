# Button

基于 el 的操作按钮，支持鼠标、键盘、图标和加载状态。默认 Primary 外观、32dp 高度、6dp 圆角；颜色在 Render 中读取当前主题。

```go
save := kit.Button("保存", saveOrder).Icon(kit.IconPlus)
deleteButton := kit.Button("删除", deleteOrder).Variant(kit.ButtonDanger)
```

Variant 接受 ButtonVariant：ButtonPrimary（零值）、ButtonSecondary、ButtonGhost、ButtonDanger、ButtonLink、ButtonText、ButtonSuccess、ButtonWarning、ButtonInfo。仅用 Variant 配置外观，不提供 Danger 等快捷入口。Size(float32) 指定高度 dp，推荐 28 / 32 / 40；非正数和非有限数忽略。窄容器限制为单行，遵守可用宽度。

Outline(true) 叠加描边外观；Compact(true) 缩小水平内边距，保留高度。Link 默认无水平内边距，Text 保留内边距；两者背景透明，悬停时改变文字颜色。只有图标、没有文案的按钮为正方形，使用 Name 设置可访问名称。

Content(el.View) 替换可见文案和图标，传 nil 恢复。内容应为展示元素，不嵌套按钮或输入框；构造时的文案（或 Name）仍作为可访问名称。加载时隐藏内容的绘制并保留布局，在中央显示进度环。

Appearance(func(ButtonAppearance) ButtonAppearance) 在每次 Render 时修改当前变体配色，可配置正常、悬停、按下、边框和焦点颜色。传 nil 恢复主题配色；自定义配色不使用 PrimaryGradient。成功、警告、信息变体按底色选择黑字或白字。禁用背景、文字和边框仍使用主题禁用色。

Icon(IconName) 配置前置图标。Loading(bool) 设置加载状态，SetLoading 可在回调中更新。加载时复用 Spinner 的动画：有图标时替换图标，无图标时保留文案的测量尺寸并在其上居中绘制 Spinner，因此切换前后按钮尺寸不变。减少动画设置开启时显示静止进度环。

SetText 更新文案，SetDisabled 更新禁用状态，均不触发点击回调。后台更新必须放在 core.Update 中。加载时保持焦点和 Tab 顺序，只忽略激活；不显示悬停、按下外观，使用默认光标。禁用时不参与 Tab 导航；父元素的 Disabled 同样限制按钮交互，同时禁用和加载时按禁用处理。Ghost、Link、Text 和 Outline 禁用时保持透明背景。

Tab / Shift+Tab 聚焦；Space / Enter 松开时激活一次。Hover、Active 和 FocusStyle 分别提供悬停、按下、焦点外观，不改变布局。

Agent 角色为 button，name 为文案；加载时 value 为 loading，disabled 为 false；只有自身或父元素禁用时 disabled 为 true。图标和 Spinner 是按钮内部装饰，不单独列入快照。

验证：`go run ./examples/components -section button`，加 `-theme dark` 检查深色。示例覆盖九种变体、描边/紧凑配置、自定义内容/配色、三个尺寸、图标、禁用、加载切换和窄容器。

## 选中状态

`Selected(true)` / `SetSelected(on)` 把按钮显示为已选中，例如工具栏里当前的视图、已启用的筛选；Agent 看到的 `selected` 为 true。实心按钮颜色加深；次要、轻量、文字和描边按钮换成选中底色 `Highlight` 和 `PrimaryText` 文字，描边按钮的边框也变成 `PrimaryText`。点击不会自动切换，需要在 OnClick 里设置；一组互斥选项直接用 [ToggleGroup](toggle_group.md)。

按钮组见 [ButtonGroup](button_group.md)。
