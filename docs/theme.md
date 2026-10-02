# 主题

`ui/theme` 管颜色、尺寸刻度和字体。组件每次 Render 都从 theme 读取，所以切换主题后所有窗口下一帧就换成新样子，不用重建视图。

## 切换主题

```go
theme.Use("nord")          // 按名字应用已注册的主题，并记住名字
theme.CurrentName()        // "nord"
theme.Names()              // 已注册的全部名字，可以直接给主题选择框用
theme.Apply(myPalette)     // 应用一个没有名字的调色板
```

`Use` 和 `Apply` 要在打开窗口前调用，或者在界面回调里调用（持有帧锁）。其他 goroutine 用 `core.Update(func() { theme.Use("dark") })`。

内置主题：`light`、`dark`、`nord`、`paper`、`solarized-dark`、`high-contrast`，以及带渐变的 `aurora`。

## 主题文件

主题文件是 JSON：选一个基础主题，只写要改的颜色。颜色名就是 `Palette` 的字段名，不区分大小写，值可以写成 `#rgb`、`#rgba`、`#rrggbb` 或 `#rrggbbaa`。

```json
{
  "name": "forest",
  "base": "light",
  "colors": {
    "bg": "#e8f3ea",
    "primary": "#2f7d4a",
    "primaryHover": "#26693d",
    "chart": ["#2f7d4a", "#c2410c"]
  }
}
```

```go
name, palette, err := theme.ParseTheme(data)
theme.Register(name, palette)
```

`chart` 最多 8 个颜色，没写的沿用基础主题。写错颜色名或颜色值时 `ParseTheme` 返回错误。

## 渐变

`bgGradient` 给窗口背景，`primaryGradient` 给主按钮、进度条和用户的聊天气泡：

```json
"bgGradient":      {"from": "#10142b", "to": "#1f1843", "angle": 120},
"primaryGradient": {"from": "#7c5cff", "to": "#2fb8ff", "angle": 0}
```

`angle` 的单位是度：0 从左到右，90 从上到下。主按钮悬停和按下时整体压暗，渐变不消失。

渐变只是外观，`Bg` 和 `Primary` 仍然是这两处的纯色，选中边框、焦点框、链接等其他地方都用纯色。所以要把 `bg`、`primary` 设成渐变里的某个颜色，看起来才协调。没设渐变（零值）时完全不影响原来的纯色。

自己画的元素也能用：`el.Div().BgGradient(theme.Gradient{From: a, To: b, Angle: 90})`。传零值什么都不改，所以可以直接传 `theme.PrimaryGradient`，没配渐变的主题下它就是纯色。之后再调 `Bg` 会替换掉渐变。

## 监听主题目录

设计师改主题文件时，不用重启程序就能看到效果：

```go
w, err := theme.WatchThemes("./themes", time.Second, func(names []string, err error) {
    if err != nil {
        log.Print(err) // 某个文件写错了；其他文件照常加载
    }
})
defer w.Stop()
theme.Use("forest")
```

- 启动时立刻注册目录里所有的 `*.json`，所以紧接着就能 `Use` 它们。
- 之后按间隔检查文件的修改时间。有变化的文件在下一帧（持有帧锁时）重新注册；正在使用的主题（`CurrentName`）变了，会立刻重新应用。
- 用轮询而不是文件系统事件，不需要额外依赖，各平台行为一致。
- 删掉文件不会注销已注册的主题。
- 组件库示例可以这样试：`go run ./examples/components -themes ./mythemes -theme forest`。

## 尺寸刻度

组件从刻度里取值，不自己写数字，同一种用途在各处看起来一样，要改设计也只改一处。

| 刻度 | 常量 |
| --- | --- |
| 间距（dp） | `SpaceXxs` 2、`SpaceXs` 4、`SpaceSm` 6、`SpaceMd` 8、`SpaceLg` 12、`SpaceXl` 16、`Space2xl` 24、`Space3xl` 32 |
| 圆角（dp） | `RadiusSm` 4、`RadiusMd` 6、`RadiusLg` 8、`RadiusXl` 12、`RadiusFull` |
| 字号（sp） | `TextXs` 11、`TextSm` 12、`TextMd` 13、`TextControl` 14、`TextBody` 15、`TextLg` 17、`TextXl` 20、`TextHeading` 22 |
| 阴影 | `ElevationSm`、`ElevationMd`、`ElevationLg`，颜色是 `theme.Shadow` |

间距的用法：图标和文字之间用 `SpaceSm`，一行里的控件之间用 `SpaceMd`，控件内边距和表单字段之间用 `SpaceLg`，卡片内边距用 `SpaceXl`，对话框和页面内边距用 `Space2xl`。

```go
el.Div().Row().Gap(theme.SpaceMd).P(theme.SpaceXl).Rounded(theme.RadiusLg)
```

kit 的间距都已改用这套刻度；剩下的 10、14、20dp 是刻意的视觉微调，比如让输入框文字和按钮文字对齐。

## 局部主题

一个窗口里的某一块想换颜色（比如浅色窗口里的深色侧栏），用 `cx.Themed`，见[元素与视图 · 局部主题](el.md#局部主题)。
