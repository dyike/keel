# Resizable

两个面板之间放一个可拖动的分隔条。

```go
split := kit.Resizable(fileTree, editor).Min(160, 320)
split.SetValue(240)                           // 第一个面板的宽度，单位 dp
stack := kit.Resizable(editor, terminal).Vertical() // 上下排列
```

- 拖动分隔条调整大小；分隔条可以获得焦点，方向键每次移动 16dp，Home / End 跳到最小或最大。
- 窗口大小变化时，第一个面板保持自己的尺寸，第二个面板占用剩下的空间。`Min(first, second)` 限制两边的最小尺寸，默认各 80dp。
- `Value()` / `SetValue(dp)`（不触发回调），`OnChange(fn)` 在用户拖动或按键调整时调用。
- 会撑满父容器给的空间，所以要放在有确定尺寸的地方。

Agent：分隔条的角色是 `separator`，名字是"调整大小"，`value` 是第一个面板的尺寸。

验证：`go run ./examples/components -section resizable`，加 `-theme dark` 检查深色。
