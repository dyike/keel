# OtpInput

分格输入固定位数的验证码。

```go
code := kit.OtpInput("验证码", 6).OnComplete(func(s string) { verify(s) })
```

- 只接受数字。输入后自动跳到下一格，Backspace 回到上一格，粘贴整串验证码会一次填满所有格子。
- 输入最后一位时调用 `OnComplete`；`OnChange` 在每次编辑后调用。
- 实现方式：一个看不见的文本框盖在格子上面，焦点、编辑、粘贴都由它处理，格子只负责显示。因此读屏和 Agent 只看到一个文本框。
- `Value()` / `SetValue`（丢弃非数字和多余的位数）、`SetDisabled`、`SetError`。

Agent：角色 `textbox`，名字是标签，`value` 是已输入的数字。

验证：`go run ./examples/components -section otp_input`，加 `-theme dark` 检查深色。

窄容器中各位等宽收缩，实际编辑区与可见格子的宽度一致。默认六位在 224dp 内容宽度下仍完整显示；1× / 2× 的数字边界和整段粘贴有回归测试。
