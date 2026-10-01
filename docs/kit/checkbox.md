# Checkbox

带标签的复选框，点击、Space、Enter 都会切换。

```go
agree := kit.Checkbox("同意条款", false).OnChange(func(on bool) { … })
all.SetMixed(true) // "全选"在部分选中时显示半选
```

- `Value()` / `SetValue(bool)` 读取或设置是否勾选，`SetValue` 不触发回调，并会清除半选状态。
- `SetMixed(true)` 显示半选，点击后变为勾选。
- `SetDisabled(true)` 后不能操作，`SetLabel` 修改文字。
- 需要用户点"提交"才生效的选项用 Checkbox；立即生效的开关用 Switch。

Agent：角色 `checkbox`，`checked` 表示是否勾选；半选时 `value` 为 `mixed`。

验证：`go run ./examples/components -section checkbox`，加 `-theme dark` 检查深色。
