# 输入框

```go
search := widget.Input("搜索").Prefix(widget.Icon(widget.IconSearch)).Suffix(widget.Kbd("mod+k").Plain()).Clearable()
search.SetDisabled(true)
```
前后缀接受任意组件，放在同一边框内；非空内容显示清空按钮，点击触发一次 `OnChange("")` 并归还编辑焦点。只读/禁用时不能清空。自带标签可点击聚焦，没有额外 Tab 焦点；`Focus(gtx)` 供应用使用。

保留密码、长度限制、提交及静默 `SetValue`。

验证入口：`go run ./examples/components -section input`。
