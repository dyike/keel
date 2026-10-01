# 角标

```go
unread := widget.Badge(3).Max(99).Child(widget.Button("通知", onOpen))
unread.SetCount(120) // 显示 99+，Count() 返回 120
widget.Badge(1).Dot().Child(widget.Text("在线"))
widget.Badge(1).Icon(widget.Icon(widget.IconCheck)).Color(theme.Primary, theme.OnColor)
```
数字、圆点和图标都在计数大于 0 时显示；隐藏时不占额外空间。`Max` 默认 99，忽略非正数配置。`Size` 支持三种尺寸；`Color` 指定背景/前景色。数字按实际字形在角标内垂直居中。`Child` 在子组件右上角显示角标，预留上下和右侧空间，使子组件与同行控件保持垂直居中，不截获子组件点击。

验证入口：`go run ./examples/components -section badge`。
