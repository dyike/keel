# Input / TextArea

带标签的文本框，可加前后缀、清空按钮和错误提示。`kit.TextArea` 创建多行版本。

```go
search := kit.Input("搜索").Placeholder("客户或单号").Clearable().Prefix(searchIcon)
price := kit.Input("单价").Filter("0123456789.").Suffix(yuan)
note := kit.TextArea("备注").Rows(4)
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
