# Form

[English](form.md) | 简体中文

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
- `Validate(cx)` 运行可见字段的全部校验、在各字段下显示错误、把焦点移到第一个不合法的字段，返回是否全部通过。必须在回调里调用。
- 自带错误展示的控件按各自编辑规则清除错误；其他控件的错误由 Form 显示，保留到下次校验或异步结果更新。`Errors()` 返回各字段当前错误的副本。
- 控件不需要再传标签：Form 会把行标签作为控件的无障碍名称。
- 自带错误展示的控件实现了 `kit.Validatable`（`SetError`、`FocusID`），包括 Input、TextArea、Select、NumberInput、OtpInput、TimeField、Combobox、DatePicker。
- `kit.Required(s, msg)` 在 s 为空或全是空白时返回 msg。`LabelWidth(dp)` 设置标签列宽，默认 72。
- `Actions(views...)` 放在字段下方的按钮行，和控件列左对齐；提交中仍可点击，方便取消。

Agent：容器角色 `form`，行标签是 `text`，控件以行标签为名字。

验证：`go run ./examples/components -section form`，加 `-theme dark` 检查深色。

所有可见字段的非 nil 校验函数都会执行。Checkbox、Switch、Radio、Slider、Rating 等非文本字段的错误由 Form 放在控件下方；有 `FocusID` 时聚焦控件，否则聚焦错误字段容器。禁用字段不会成为焦点目标。NumberInput、TimeField 和 Combobox 的可用编辑草稿在校验前提交，避免校验读取旧值。Required 将 Unicode 空白（含全角空格）视为空值。

异步校验与提交共用一次请求：

```go
token := f.BeginSubmit(cx)
if token == 0 { return } // 同步校验失败、禁用或已有请求
payload := customer.Value() // 在 UI 回调里捕获，后台只使用快照
go func() {
    errors := validateAndSave(payload) // []string，按 Field 顺序；nil 表示成功
    core.Update(func() { f.FinishSubmit(token, errors) })
}()
```

`Submitting()` 为 true 时字段不可编辑，重复 `BeginSubmit` 返回 0；外部提交按钮可用 `.Loading(f.Submitting())` 显示忙碌状态。`FinishSubmit` 恢复编辑、展示各字段错误，并在字段恢复可用后聚焦首个错误。错误切片可短于字段数，剩余字段按无错误处理；过长切片、过期或取消的 token 返回 false，不改变状态。

`CancelSubmit()` 使未完成结果失效并恢复编辑；它不会取消业务 goroutine 的网络请求，应用需要自行取消 I/O。`SetDisabled(true)`、祖先禁用或再次同步 `Validate` 也会使请求失效。程序在提交期间替换字段值、或移除表单时，应先调用 `CancelSubmit`。后台只能通过 `core.Update` 调用完成接口。

## 多列与字段配置

`Columns(n)` 将字段排成 n 个等宽网格列，至少一列；`VerticalLabels(true)` 把标签放到控件上方。两者独立，默认仍为单字段列、标签在左。响应式断点由应用决定，可在 Render 中按可用宽度调用 Columns。`Gap(dp)` 设置字段间距，`LabelTextSize(sp)` 设置标签字号（0 恢复继承）。控件大小由各控件自身配置。

```go
f := kit.Form().Columns(2).VerticalLabels(true).
    FieldWithOptions("姓名", name, validateName, kit.FormFieldOptions{
        Required: true, Description: "公开显示的姓名",
    }).
    FieldWithOptions("邮箱", email, validateEmail, kit.FormFieldOptions{
        Required: true,
    }).
    FieldWithOptions("介绍", bio, nil, kit.FormFieldOptions{ColSpan: 2}).
    Footer(kit.Button("保存", save))
```

`ColSpan` 默认 1，限制在当前列数以内；`ColStart` 从 1 计数，0 表示顺序排列。指定起始列已被当前行占用时从下一行开始；起始列和跨度超过网格边界时收缩到可用列。隐藏字段不占网格位置。Actions 继续按原有方式排列；Footer 独占底部全宽并靠右对齐，可与 Actions 共用，提交中保持可用。

`Required` 只显示星号，不替代校验器。`Description` 显示控件下的辅助文字；`DescriptionContent` 可提供富内容或动态 View，优先于纯文本。`Hidden` 默认 false；隐藏字段不渲染、不提交草稿、不执行校验，也不会成为错误聚焦目标。其他字段的插入索引和输入状态不受影响。

`SetFieldOptions(index, options)` 按添加顺序替换配置，非法索引返回 false。改变 Hidden 会取消正在进行的提交并清除此字段错误；其他展示配置变更不会取消提交。异步错误数组仍按完整字段顺序传入，隐藏字段对应的错误被忽略。不要直接修改提交快照中的业务数据。

自动测试覆盖 1×/2× 网格定位、跨列、起始列、隐藏、缩为单列、输入状态、必填提示/说明/页尾，以及隐藏字段校验和异步失效。示例展示响应式两列，尚未完成真机视觉验收。
