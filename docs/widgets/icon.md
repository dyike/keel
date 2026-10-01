# 图标

```go
widget.Icon(widget.IconSearch).Size(18).Color(theme.Muted)
widget.VectorIcon(customGioIcon)
```
内置 Check、Close、Plus、Search、Copy、ChevronDown、ChevronRight，默认 18dp。自定义图标接受 Gio `widget.NewIcon` 解码的图标；nil 图标不占空间。装饰图标的可访问名称由所在控件提供。

验证入口：`go run ./examples/components -section icon`。
