# ui/locale

Keel 框架自己显示或报告给 Agent 的文字：确定、取消、复制、关闭、请选择、"36 行"等。用法和 `theme` 一样，切换后所有窗口重绘，`cx.Cache` 自动失效。

```go
locale.Apply(locale.English())                      // 开窗前或回调中
core.Update(func() { locale.Apply(locale.English()) }) // 后台 goroutine

s := locale.Chinese() // 只改几处：从预设复制再改
s.OK = "好的"
locale.Apply(s)
```

- 预设：`Chinese()`（默认）、`English()`。`Apply` 替换全部文字，所以自定义时要从预设开始改。
- 计数用函数：`Rows(n)`，英文会区分 1 row 和 2 rows。
- `Name(action, target)` 拼接无障碍名称，例如"关闭 保存成功"、"Close Saved"。
- **只管框架文字。** 应用自己的文案由应用负责：可以按 `Current().Lang` 选择自己的文案表，在 Render 里读取，切换语言后会自动重绘；自建缓存的 key 要带上 `Revision()`。
- 组件在 Render 或 Layout 时读取 `locale.Current()`，不能在构造时保存文字。

- 依赖：`ui/internal/loop`（仅用于通知重绘）。
- 被依赖：`el`（缓存失效）、`kit`、`widget`、`markdown`。
