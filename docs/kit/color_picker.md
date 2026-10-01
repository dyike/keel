# ColorPicker

颜色选择器。

```go
picker := kit.ColorPicker().Alpha().Swatches(presets...).OnChange(func(c color.NRGBA) { … })
picker.SetValue(theme.Primary)
```

- 方块里拖动选择饱和度和亮度，色相条选择色相，`Alpha()` 后多一条不透明度条。方块和各个条都可以获得焦点，用方向键微调。
- 十六进制输入框接受 `#RGB`、`#RRGGBB`、`#RRGGBBAA`，回车或失去焦点时生效，输入不合法时恢复原值。
- `Swatches(...)` 在下方显示预设色块，每行 8 个。
- 颜色内部按 HSV 保存：把颜色拖成灰色或黑色后，色相不会丢失，再拖回来时还是原来的色相。
- 想放进弹层时，和 `kit.Popover` 组合使用，见示例。

Agent：容器角色 `group`，名字是当前颜色的十六进制值；方块和各个条的角色是 `slider`（名字为"饱和度与亮度""色相""不透明度"）；输入框名为 HEX；预设色块是以十六进制值命名的按钮。

验证：`go run ./examples/components -section color_picker`，加 `-theme dark` 检查深色。
