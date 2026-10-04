# ColorPicker

颜色选择器。

```go
picker := kit.ColorPicker().Alpha().Swatches(presets...).OnChange(func(c color.NRGBA) { … })
picker.SetValue(theme.Primary)
```

- 方块里拖动选择饱和度和亮度，色相条选择色相，`Alpha()` 后多一条不透明度条。方块和各个条都可以获得焦点，用方向键微调。
- 十六进制输入框接受 `#RGB`、`#RRGGBB`、`#RRGGBBAA`，回车或失去焦点时生效，输入不合法时恢复原值。
- `Swatches(...)` 在下方显示预设色块，根据可用宽度自动换行，并复制预设切片。
- 颜色内部按 HSV 保存：把颜色拖成灰色或黑色后，色相不会丢失，再拖回来时还是原来的色相。
- `Popup(true)` 使用内置触发按钮与锚定弹层；默认内联，也仍可自行组合 `kit.Popover`。

Agent：容器角色 `group`，名字是当前颜色的十六进制值；方块和各个条的角色是 `slider`（名字为"饱和度与亮度""色相""不透明度"）；输入框名为 HEX；预设色块是以十六进制值命名的按钮。

验证：`go run ./examples/components -section color_picker`，加 `-theme dark` 检查深色。

`SetDisabled(true)` 禁用色板、滑块、预设色和 HEX 输入，并取消尚未提交的 HEX 草稿。父容器禁用引起的失焦不会提交草稿。`SetValue` 在禁用期间仍可更新颜色且不调用 `OnChange`。

默认宽度 240dp，放入窄容器时缩到可用宽度。滑块指示环根据实际布局定位，黑白双描边确保在亮色和暗色区域都可辨认。色相和透明度条的命中高度为 24dp，焦点有独立边框。

方向键微调，Shift + 方向键加速十倍；色相/透明度条支持 Home/End 到端点、PageUp/PageDown 十步调整。色块的选中状态同时用勾形图标和 Agent `selected` 表达，图标按合成后的背景选择黑或白。重复点击当前色块不重复触发回调。拖动取消保留最后一次有效颜色，不用取消事件的坐标覆盖它。

`Label(text)` 在选择器或触发器上方显示标签，并作为内置按钮的可访问名称。无标签时按钮使用当前 HEX。`Icon(name)` 替换触发器中的颜色块，`IconNone` 恢复颜色块；按钮始终显示当前 HEX。颜色修改继续使用原 OnChange。

`Size(ColorPickerSizeXSmall/Small/Medium/Large)` 同步调整弹层/内联面板宽度、色板高度、HEX 字段和触发按钮尺寸；面板宽度依次为 192/216/240/280dp，按钮高度为 24/28/32/40dp。默认 Medium 保留原内联布局，非法档位忽略。

`SetOpen(bool)` 程序设置弹层，`IsOpen()` 查询；仅 Popup 模式有效，不触发颜色回调。用户打开时焦点移到色板；Esc 或点击外部关闭，焦点回到触发按钮。关闭或禁用取消未提交的 HEX 草稿，已提交的颜色保留。切回内联也关闭弹层。键盘、禁用、草稿取消、尺寸切换和值保留通过 1×/2× 自动测试；真机弹层手感仍待验收。
