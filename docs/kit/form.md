# Form

两列表单：左边是标签，右边是控件，提交时统一校验。

```go
name := kit.Input("")
amount := kit.NumberInput("").Range(0, 1e6)
f := kit.Form().
    Field("客户", name, func() string { return kit.Required(name.Value(), "请填写客户") }).
    Field("金额", amount, func() string {
        if amount.Value() <= 0 {
            return "金额必须大于 0"
        }
        return ""
    })
kit.Button("创建", func() {
    if f.Validate(cx) {
        create()
    }
})
```

- 校验函数返回错误信息，合法时返回空字符串；可以传 nil 表示不校验这一项。
- `Validate(cx)` 运行全部校验、在各字段下显示错误、把焦点移到第一个不合法的字段，返回是否全部通过。必须在回调里调用。
- 错误会在用户修改字段后消失，下一次 `Validate` 时重新计算。
- 控件不需要再传标签：Form 会把行标签作为控件的无障碍名称。
- 可校验的控件实现了 `kit.Validatable`（`SetError`、`FocusID`），包括 Input、TextArea、Select、NumberInput、OtpInput、TimeField、Combobox、DatePicker。
- `kit.Required(s, msg)` 在 s 为空或全是空白时返回 msg。`LabelWidth(dp)` 设置标签列宽，默认 72。

Agent：容器角色 `form`，行标签是 `text`，控件以行标签为名字。

验证：`go run ./examples/components -section form`，加 `-theme dark` 检查深色。
