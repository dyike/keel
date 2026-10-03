# Dialog

模态对话框。打开时页面变暗，不响应指针，焦点限制在对话框内；关闭后焦点回到打开之前的位置。

**标准消息**：复用同一个实例。

```go
dlg := kit.Dialog("")
dlg.Confirm("保存修改", "离开前保存吗？", save)                  // 取消 / 确定，焦点在确定上
dlg.ConfirmDanger("删除订单", "删除后不能恢复。", "删除", remove)  // 焦点在取消上
dlg.Alert("导出完成", "共 36 条记录。", nil)                      // 只有确定
```

**自定义内容**：

```go
edit := kit.Dialog("编辑客户").Body(form).Footer(cancelButton, saveButton).Width(480)
edit.SetValue(true)
```

- 对话框要放在视图树里渲染：`dlg.Render(cx)` 在原位置不显示任何内容，只是声明浮层。需要 `el.Root`。
- 关闭方式：Esc、点击遮罩、取消按钮。关闭时调用 `OnClose(fn)`；标准消息的"确定"不算关闭，只运行它自己的回调。
- `ConfirmDanger` 和 `Persistent()` 的对话框点击遮罩不会关闭，防止误触；Esc 等同于"取消"。
- 自定义对话框的按钮由调用方提供，按钮回调里自己调用 `SetValue(false)` 关闭对话框。
- `Value()` / `SetValue(bool)` 读取或设置是否打开；`SetTitle` 修改标题；`Width(dp)` 设置宽度，默认 420，窄窗口中不超过窗口宽度。

Agent：普通对话框的角色是 `dialog`，`ConfirmDanger` 和 `Persistent()` 的角色是 `alertdialog`，名字是标题，里面的元素单独列出。对话框打开期间，快照里看不到下面的页面。

验证：`go run ./examples/components -section dialog`。

对话框限制在窗口内，长正文单独滚动，窄窗口的页脚按钮改为纵向排列；极小窗口允许整个面板滚动。非有限宽度被忽略，Footer 保存视图列表的副本。

`SetDisabled(true)` 直接关闭，且阻止 SetValue / 标准消息重新打开；不触发用户关闭回调。所在容器禁用、隐藏或不再提供所属元素时，已声明的模态层会请求关闭并调用一次 OnClose。关闭后不会因恢复启用而重新弹出。

正文和页脚可以包含 Menu、Popover 或另一个 Dialog。父浮层先登记，子浮层位于上方；Esc 逐层关闭，每层恢复对应的先前焦点。示例“自定义”里的客户预设用于验证嵌套菜单。

关闭配置可独立设置，并在复用标准消息时保留：

- `Keyboard(false)` 禁止 Esc 关闭，仍消耗该键，避免误关下层对话框；默认开启。
- `Overlay(false)` 隐藏遮罩颜色，仍阻挡背景操作并约束焦点；默认显示。
- `OverlayClosable(bool)` 显式设置外部点击是否关闭，优先于 Persistent/ConfirmDanger 的默认值。
- `CloseButton(true)` 显示标题栏关闭按钮，调用与取消/Esc 相同的关闭逻辑。默认隐藏，保留旧布局；按钮名称随语言切换。无标题也能显示。

关闭按钮、标题、正文和页脚使用稳定身份；切换显示配置不会重建正文输入状态。上述配置不限制程序调用 `SetValue(false)`，也不限制自定义 Footer 按钮。

`BeforeConfirm(func() bool)` 在标准 Confirm / ConfirmDanger / Alert 的确定操作前运行。返回 false 保持打开、保留焦点且不执行原 onOK，适用于同步校验或等待后台任务；返回 true 后按原顺序先关闭再执行 onOK，不调用 OnClose。nil 清除校验，复用标准消息时保留配置。取消/Esc/遮罩/关闭按钮不受此校验影响；自定义 Footer 仍由应用控制。

```go
dlg.BeforeConfirm(func() bool { return formIsValid() })
dlg.Confirm("提交", "确认提交？", submit)
```

校验回调中若调用 SetValue、SetDisabled 或打开另一条标准消息，旧确认操作不会继续关闭或执行 onOK。校验不会自动启动 goroutine，也不会自动显示加载状态。
