# 键帽

```go
widget.Kbd("mod+shift+p") // macOS ⇧⌘P，Windows/Linux Ctrl+Shift+P
widget.Kbd("mod+s").Plain().Size(widget.Small)
```
使用 `core.ParseShortcut` 的 `+` 分隔语法。`mod` 按当前平台显示，无法解析的组合按原文显示；Escape 用 `Esc`，避免缺字。`Size` 调整尺寸，`Plain` 去掉边框，`SetShortcut` 更新提示。只显示，不注册快捷键，也不查询动作绑定。

验证入口：`go run ./examples/components -section kbd`。
