# Collapsible

[English](collapsible.md) | 简体中文

单节内容的展开/收起状态，可整体使用，也可把触发器与内容分别放进布局。

```go
section := kit.Collapsible("高级筛选", form)
section.SetValue(true)
// 默认带边框的面板：section.Render(cx)
row.Child(section.Trigger().Render(cx))
body.Child(section.Content().Render(cx))
```

`Value` / `SetValue` 读取和设置状态，程序设置不回调；`OnChange` 在用户切换时调用。`SetDisabled` 同时禁用独立触发器和内容。每个实例的 Trigger / Content 各渲染一次，不要同时再调用整体 Render。

`Heading(view)` 自定义标题的展示内容，构造时的 label 仍作为无障碍名称。标题内部放文字、图标等展示内容，避免嵌套另一个交互控件。

触发器支持 Tab、Enter / Space，角色为 disclosure，值为 expanded / collapsed。关闭正在编辑的内容时，焦点回到触发器，输入状态保留。收起中的内容不能操作。

展开与收起使用 180ms 高度动画；重复切换从当前高度反转，正文维持自然布局，不挤压文字。减少动画时立即切换，初始展开状态也直接显示。

验证：`go run ./examples/components -section collapsible`，加 `-theme dark` 检查深色。
