# Slider

[English](slider.md) | 简体中文

拖动或用键盘选择数值或双端范围。

```go
volume := kit.Slider("音量", 0, 100).Step(5)
price := kit.RangeSlider("价格区间", 0, 100).Step(5)
price.SetValues(20, 80)
level := kit.RangeSlider("竖向区间", 0, 100).Vertical(180)
```

- 拖动滑块，或在轨道上任意位置按下后拖动。
- 键盘：← ↓ 减一步，→ ↑ 加一步，PageUp / PageDown 移动 10 步，Home / End 跳到两端。
- `Step(s)` 让数值对齐到 `min + k·s`；为 0 时连续取值，键盘每次移动范围的 1%。
- `RangeSlider` 初始选中整个范围。`Values()` / `SetValues(low, high)` 读取或设置两端；程序设置会排序、对齐和钳制，不触发 `OnRangeChange`。用户拖动时选择距离最近的一端，碰到另一端就停止，不交换端点身份。
- 双端各自支持 Tab 焦点和按键；Home / End 受另一端约束。上下限重叠时，在重叠处按下选择上限，可向右或向上重新拉开；下限也可用 Tab 独立操作。
- `Vertical(height)` 设置竖向轨道高度（dp），底部为最小值、顶部为最大值；↑ 增加、↓ 减少。
- `Value()` / `SetValue`（自动对齐和限制在范围内）、`SetRange`、`SetDisabled`。

Agent：单端角色为 `slider`；双端分别是本地化“下限 / 上限 + 标签”命名的 `slider`，`value` 是该端数值。`FocusID()` 在双端模式中指向下限。

验证：`go run ./examples/components -section slider`，加 `-theme dark` 检查深色。

`Scale(kit.SliderLogarithmic)` 使用对数刻度，单值/双端、横向/竖向均可使用。有效范围要求 `0 < min < max`；不满足时按线性刻度显示，之后设置有效范围会恢复对数刻度。默认 `SliderLinear`。

```go
frequency := kit.Slider("频率 Hz", 20, 20000).
    Scale(kit.SliderLogarithmic).
    OnRelease(func(value float64) { applyFrequency(value) })
```

例如 1–1000 的对数范围，轨道正中间约为 31.62，四分之一处约为 5.62。未设置正步长时，方向键移动轨道的 1%，PageUp/Down 移动 10%；显式 `Step(s)` 仍按原始数值增减和对齐。双端拖动按轨道上的距离选择最近端点。

`OnRelease(func(float64))` 接收单值操作结束时的值；`OnRangeRelease(func(low, high float64))` 接收双端范围。鼠标/触摸释放触发一次；导航按键重复按下只改变值，释放最后操作的导航键时触发一次。没有数值变化的有效点击也会触发。`OnChange` / `OnRangeChange` 仍在值变化时连续触发。

程序赋值、拖动取消、失焦后的按键释放及禁用操作不触发结束回调；取消不撤销已经产生的值变化。可用 OnChange 更新预览，用 OnRelease 提交开销较大的操作。传 nil 可移除对应回调。

`Appearance(func(*kit.SliderAppearance))` 配置当前实例的轨道与滑块外观；每次渲染先读取当前主题默认值，再运行回调，传 nil 恢复默认。双端共用外观，配置不会改变值或触发值回调。

```go
price.Appearance(func(a *kit.SliderAppearance) {
    a.TrackSize, a.ThumbSize = 8, 24
    a.TrackRadius, a.ThumbRadius = 2, 4
    a.FillColor, a.ThumbBorderColor = theme.Success, theme.Success
})
```

颜色包括 `TrackColor`、`FillColor`、`ThumbColor`、`ThumbBorderColor`；尺寸为 dp。默认轨道 4、滑块 16、边框 2；轨道/滑块尺寸必须为有限正数，无效值回退默认，最大 1024dp。边框与圆角允许 0，无效值回退默认；边框和滑块圆角最大为滑块尺寸的一半；轨道圆角最大 1024dp。横向最小宽度、竖向最小高度为滑块尺寸，指针数值映射使用滑块中心之间的距离。自身禁用时填充和滑块边框改为主题 Muted，键盘焦点仍显示主题焦点色。

已验证横纵方向、双端范围、双倍率布局、默认恢复、禁用与键盘操作，并覆盖自定义颜色的窗口像素回归。
