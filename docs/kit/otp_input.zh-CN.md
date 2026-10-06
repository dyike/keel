# OtpInput

[English](otp_input.md) | 简体中文

分格输入固定位数的验证码。

```go
code := kit.OtpInput("验证码", 6).OnComplete(func(s string) { verify(s) })
```

- 只接受数字。输入后自动跳到下一格，Backspace 回到上一格，粘贴整串验证码会一次填满所有格子。
- 输入最后一位时调用 `OnComplete`；`OnChange` 在每次编辑后调用。
- 实现方式：一个看不见的文本框盖在格子上面，焦点、编辑、粘贴都由它处理，格子只负责显示。因此读屏和 Agent 只看到一个文本框。
- `Value()` / `SetValue`（丢弃非数字和多余的位数）、`SetDisabled`、`SetError`。

`Masked(true)` 遮住可见数字与语义值，`Masked(false)` 恢复显示；原始值仍通过 `Value` 和回调交给应用。`Groups(n)` 设置组数，限制到 1–位数，默认两组（单数字为一组），不能整除时前面的组多一位，例如 7 位分 3 组为 3–2–2。组间间距为格内间距的两倍。

`Size(dp)` 设置格高，默认 48dp；格宽、间距和字号按比例变化，忽略非正数和非有限数。修改布局或遮罩不改变值，不触发编辑回调。

Agent：角色 `textbox`，名字是标签；普通模式的 `value` 是已输入的数字，遮罩模式为等长圆点。

验证：`go run ./examples/components -section otp_input`，加 `-theme dark` 检查深色。

窄容器中各位等宽收缩，实际编辑区与可见格子的宽度一致。默认六位在 224dp 内容宽度下仍完整显示；1× / 2× 的数字边界和整段粘贴有回归测试。
