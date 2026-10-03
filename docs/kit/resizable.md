# Resizable

两个面板之间放一个可拖动的分隔条。

```go
split := kit.Resizable(fileTree, editor).Min(160, 320)
split.SetValue(240)                           // 第一个面板的宽度，单位 dp
stack := kit.Resizable(editor, terminal).Vertical() // 上下排列
```

- 拖动分隔条调整大小；分隔条可以获得焦点，方向键每次移动 16dp，Home / End 跳到最小或最大。
- 窗口大小变化时，第一个面板保持自己的尺寸，第二个面板占用剩下的空间。`Min(first, second)` 限制两边的最小尺寸，默认各 80dp。
- `SetDisabled(true)` 禁用分隔条及两侧内容，移除焦点并停止拖动、按键和用户回调；`SetValue` 仍可调整尺寸。
- `Value()` / `SetValue(dp)`（不触发回调），`OnChange(fn)` 在用户拖动或按键调整时调用。
- 会撑满父容器给的空间，所以要放在有确定尺寸的地方。

Agent：分隔条的角色是 `separator`，名字是"调整大小"，`value` 是第一个面板的尺寸。

验证：`go run ./examples/components -section resizable`，加 `-theme dark` 检查深色。

`Max(first, second)` 设置两侧最大尺寸，0 表示该侧不限。限制适用于拖动、方向键、Home/End 和 SetValue。负数/非有限上限按参数分别忽略；Min 和 SetValue 也忽略非有限值。上限低于同侧 Min 时以 Min 为准。

容器有足够空间时，同时满足两侧范围；两侧上限之和不足以填满容器时，第二面板后留空。空间不足以满足最小尺寸时，优先保留第二面板的最小空间，第一面板可缩至 0；容器小于分隔条时仍保留 6dp 把手。窗口测量变化后请求下一帧收敛，不调用 OnChange。配置 Min/Max 后在下一次 Render 收敛，SetValue 立即按当前已知容器尺寸约束。

`Visible(first, second)` 分别控制两侧显隐，默认均显示。仅一侧可见时，该面板填满容器，分隔条隐藏，暂不应用分割的 Min/Max；两侧隐藏时保留容器但不展示内容。面板内容继续声明，隐藏后不能聚焦或接收输入，恢复后保留输入状态；隐藏侧原焦点不会自动恢复。

显隐和隐藏期间的窗口变化不改写保存的分隔尺寸，也不触发 OnChange。恢复双面板时按当前容器和 Min/Max 收敛。隐藏期间 SetValue 仍可修改恢复尺寸，受第一面板 Min/Max 限制，容器约束延后至双面板恢复。

把手的命中区域仍为 6dp，默认视觉线宽为空闲 1dp、悬停/聚焦 2dp、按下 3dp、拖动 4dp。拖出命中区域仍保持拖动态；释放后恢复悬停或焦点态，取消保留最后有效尺寸。线宽过渡默认 150ms，遵守减少动画设置。

`HandleAppearance(fn)` 每帧传入当前主题默认值，可修改 `Idle/Hover/Pressed/Dragging` 线宽、`Color/ActiveColor` 和 `Duration`。线宽限制在 0–6dp，0 隐藏该状态，非有限值回退 1dp；Duration≤0 立即切换。nil 恢复默认。仅线宽渐变，状态颜色立即切换。此配置不改变命中宽度和面板范围。
