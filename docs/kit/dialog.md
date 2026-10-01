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
