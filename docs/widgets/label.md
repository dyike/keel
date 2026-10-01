# 标签关联

```go
name := widget.Input("")
label := widget.Text("用户名").For(name)
```
`For` 接受实现 `Focus(core.C)` 的目标；点击（含触摸）会转交焦点，不增加独立的 Tab 焦点。禁用输入框不会接受聚焦。保留 Text、Heading、Muted 的字号、颜色、换行与 `SetText` 行为。

验证入口：`go run ./examples/components -section label`。
