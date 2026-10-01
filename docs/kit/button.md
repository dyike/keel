# Button

基于 el 的操作按钮，支持鼠标、键盘、图标和加载状态。默认 Primary 外观、32dp 高度、6dp 圆角；颜色在 Render 中读取当前主题。

```go
save := kit.Button("保存", saveOrder).Icon(kit.IconPlus)
deleteButton := kit.Button("删除", deleteOrder).Variant(kit.ButtonDanger)
```

Variant 接受 ButtonVariant：ButtonPrimary（零值）、ButtonSecondary、ButtonGhost、ButtonDanger。仅用 Variant 配置外观，不提供 Danger 等快捷入口。Size(float32) 指定高度 dp，推荐 28 / 32 / 40；非正数和非有限数忽略。窄容器限制为单行，遵守可用宽度。

Icon(IconName) 配置前置图标。Loading(bool) 设置加载状态，SetLoading 可在回调中更新。加载时复用 Spinner 的动画：有图标时替换图标，无图标时保留文案的测量尺寸并在其上居中绘制 Spinner，因此切换前后按钮尺寸不变。减少动画设置开启时显示静止进度环。

SetText 更新文案，SetDisabled 更新禁用状态，均不触发点击回调。后台更新必须放在 core.Update 中。加载时保持焦点和 Tab 顺序，只忽略激活；不显示悬停、按下外观，使用默认光标。禁用时不参与 Tab 导航；父元素的 Disabled 同样限制按钮交互，同时禁用和加载时按禁用处理。Ghost 禁用时保持透明背景。

Tab / Shift+Tab 聚焦；Space / Enter 松开时激活一次。Hover、Active 和 FocusStyle 分别提供悬停、按下、焦点外观，不改变布局。

Agent 角色为 button，name 为文案；加载时 value 为 loading，disabled 为 false；只有自身或父元素禁用时 disabled 为 true。图标和 Spinner 是按钮内部装饰，不单独列入快照。

验证：`go run ./examples/components -section button`，加 `-theme dark` 检查深色。示例覆盖四种外观、三个尺寸、图标、禁用、加载切换和窄容器。Button 的接口在 M2 第二节完成后交 review，再继续后续组件。
