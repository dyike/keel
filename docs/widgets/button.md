# 按钮

```go
save := widget.Button("保存", onSave).Icon(widget.Icon(widget.IconCheck)).Size(widget.Large)
save.SetLoading(true)
```
保留 Primary、Secondary、Danger 和禁用行为。`Size(Small/Medium/Large)` 默认 Medium，图标由按钮统一着色和定尺寸。`SetLoading` / `Loading` 控制加载指示；加载期间鼠标和键盘都不能重复触发回调，结束后由应用调用 `SetLoading(false)`。`ComponentSize` 与内部加载绘制供后续控件共用。

验证入口：`go run ./examples/components -section button`。
