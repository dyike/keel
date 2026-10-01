# DropdownButton

带下拉菜单的按钮。

```go
kit.DropdownButton("导出", formats)              // 整个按钮打开菜单
kit.DropdownButton("保存", more).Split(save)     // "保存"执行 save，旁边的箭头打开菜单
```

- 菜单是普通的 `kit.Menu`，键盘、子菜单、禁用项的行为都和 Menu 一致。
- `Variant(kit.ButtonSecondary)` 等设置按钮外观，分体样式下两个部分使用同一外观。
- `SetDisabled(true)` 同时禁用按钮和箭头，并关闭已打开的菜单。
- 需要 `el.Root`。

Agent：整体样式下，按钮名字就是标题；分体样式下，主操作按钮名为标题，箭头按钮名为"标题 更多选项"。

验证：`go run ./examples/components -section dropdown_button`。
