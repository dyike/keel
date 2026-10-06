# Checkbox

[English](checkbox.md) | 简体中文

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

`Size(dp)` 设置方框边长，正值限制在 12–64dp，0 恢复默认 18dp；勾号和半选横线同步缩放。`TextSize(sp)` 单独设置标签字号，正值限制在 8–128sp，0 恢复父元素字号。负数和非有限值忽略；修改外观不改变值、半选状态或现有焦点。

```go
agree.Size(24).TextSize(18).TabIndex(2)
```

`TabStop(false)` 仅跳过 Tab 遍历，仍可鼠标和程序聚焦。`TabIndex(n)` 按升序遍历，相同值保持树顺序，负值跳过。复用 el 的单 root 规则，浮层内独立循环，不跨独立 Embed 或原生 Gio 控件排序。焦点轮廓仍沿标签和方框整行绘制。
